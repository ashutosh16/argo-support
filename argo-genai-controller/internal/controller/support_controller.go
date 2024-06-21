/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	v1alpha1 "github.com/argoproj-labs/argo-support/api/v1alpha1"
	"github.com/argoproj-labs/argo-support/internal/wf_operations"
	"github.com/argoproj-labs/argo-support/internal/wf_operations/genai"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sort"
	"time"
)

const ReconcileRequeueOnValidationError = 10 * time.Second

// SupportReconciler reconciles a Support object
type SupportReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	DynamicClient dynamic.DynamicClient
	KubeClient    kubernetes.Interface
}

//+kubebuilder:rbac:groups=support.argoproj.extensions.io,resources=supports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=support.argoproj.extensions.io,resources=supports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=support.argoproj.extensions.io,resources=supports/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Support object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile

func (r *SupportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var err error
	var support v1alpha1.Support
	err = r.Get(ctx, req.NamespacedName, &support, &client.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Argo support operation not found", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Argo support")
		return ctrl.Result{}, err
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	if support.ObjectMeta.DeletionTimestamp != nil {
		logger.Info("Argo support is being deleted", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, nil
	}

	if support.Status.Phase == v1alpha1.ArgoSupportPhaseFailed ||
		support.Status.Phase == v1alpha1.ArgoSupportPhaseRunning {
		logger.Info("spec is  not observed", "support.ObjectMeta.Generation", support.Status.ObservedGeneration)
	} else if support.ObjectMeta.Generation == support.Status.ObservedGeneration {
		logger.Info("skipping..no change to spec version with observed version", "support.ObjectMeta.Generation", support.Status.ObservedGeneration)
		return ctrl.Result{}, nil
	}

	finalizerErr := r.handleFinalizer(ctx, &support)
	if finalizerErr != nil {
		logger.Error(finalizerErr, "Failed to handle finalizer")
		return ctrl.Result{}, finalizerErr
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if &support != nil && support.Annotations[v1alpha1.ArgoSupportWFFeedbackAnnotationKey] != "" {
			return nil
		}
		var obj v1alpha1.Support
		err := r.Get(ctx, types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		}, &obj)
		if err != nil {
			return err
		}

		now := metav1.Now()
		support.Status.LastTransitionTime = &now
		support.Status.Count++
		support.Status.Phase = v1alpha1.ArgoSupportPhaseRunning
		return r.Status().Update(ctx, &support)
	})

	for _, wf := range support.Spec.Workflows {

		if support.Status.Count >= wf.RetryLimit {
			support.Status.Phase = v1alpha1.ArgoSupportPhaseError
			continue
		}

		wfExecutor, err := r.getWfExecutor(ctx, &wf, &support)
		var originalCopy = support.DeepCopy()

		if wfExecutor != nil {

			if &support != nil && support.Annotations[v1alpha1.ArgoSupportWFFeedbackAnnotationKey] != "" {
				var f *v1alpha1.Feedback
				f, err = wfExecutor.UpdateWorkflow(ctx, &support)
				if f != nil {
					logger.Info("feedback collect:", "workflow name", wf.Name, "feedback", &f)
					for _, result := range support.Status.Results {
						if result.Name == f.Name {
							result.Feedback = f
						}
					}
				}
				r.Patch(ctx, &support, client.MergeFrom(originalCopy))
				continue
			} else {

				err = wfExecutor.RunWorkflow(ctx, &support)
			}
			if err != nil {
				logger.Error(err, "Failed to process workflow "+wf.Name)
				support.Status.Phase = v1alpha1.ArgoSupportPhaseFailed
				continue
			}
			if support.Status.Phase == v1alpha1.ArgoSupportPhaseRunning {
				continue
			}
			if &support != nil && len(support.Status.Results) > 1 {
				sort.SliceStable(support.Status.Results, func(i, j int) bool {
					return (support.Status.Results[i].FinishedAt.Time).After(support.Status.Results[j].FinishedAt.Time)
				})

				if len(support.Status.Results) > 2 {
					support.Status.Results = support.Status.Results[:2]
				}
			}
			now := metav1.Now()
			support.Status.LastTransitionTime = &now
		} else {
			support.Status.Phase = v1alpha1.ArgoSupportPhaseFailed
			logger.Error(err, "Failed to get workflow executor for "+wf.Name)
		}
	}

	if support.Status.Phase == v1alpha1.ArgoSupportPhaseCompleted || support.Status.Phase == v1alpha1.ArgoSupportPhaseError {
		support.Status.Count = 0
		support.Status.ObservedGeneration = support.ObjectMeta.Generation
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var obj v1alpha1.Support
		err := r.Get(ctx, types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		}, &obj)
		if err != nil {
			return err
		}
		obj.Status.Count++
		obj.Status = support.Status
		return r.Status().Update(ctx, &obj)
	})

	if err != nil {
		logger.Error(err, "Failed to update support status"+support.Name+" in namespace "+support.Namespace)
	}

	logger.Info("reconciliation completed successfully", "support", support.Name, "namespace", support.Namespace, "phase", support.Status.Phase)
	if support.Status.Phase == v1alpha1.ArgoSupportPhaseFailed {
		logger.Error(err, "Failed to get workflow executor")
		return ctrl.Result{RequeueAfter: ReconcileRequeueOnValidationError}, nil
	} else {
		return ctrl.Result{}, nil

	}
}

func (r *SupportReconciler) getWfExecutor(ctx context.Context, wf *v1alpha1.Workflow, obj metav1.Object) (wf_operations.Executor, error) {

	switch {
	case wf.Name == "gen-ai":
		ops, err := genai.NewGenAIOperations(ctx, r.Client, r.DynamicClient, r.KubeClient, wf, obj.GetNamespace())
		if err != nil {
			return nil, err
		}
		return ops, nil
	default:
		return nil, nil
	}
}

func (r *SupportReconciler) handleFinalizer(ctx context.Context, ops *v1alpha1.Support) error {
	// name of our genstudio finalizer

	// examine DeletionTimestamp to determine if object is under deletion
	if ops.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(ops, v1alpha1.FinalizerName) {
			controllerutil.AddFinalizer(ops, v1alpha1.FinalizerName)
			if err := r.Update(ctx, ops); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(ops, v1alpha1.FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(ops, v1alpha1.FinalizerName)
			if err := r.Update(ctx, ops); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SupportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Support{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
