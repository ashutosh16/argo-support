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
	supportv1alpha1 "github.com/argoproj-labs/argo-support/api/v1alpha1"
	"github.com/argoproj-labs/argo-support/internal/wf_operations"
	"github.com/argoproj-labs/argo-support/internal/wf_operations/genai"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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
	var support supportv1alpha1.Support
	err = r.Get(ctx, req.NamespacedName, &support, &client.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Argo support operation not found", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Argo AI support operation")
		return ctrl.Result{}, err
	}
	/*
		TODO
			// Do not attempt to further reconcile the ApplicationSet if it is being deleted.
			if support.ObjectMeta.DeletionTimestamp != nil {
				support := support.ObjectMeta.Name
				logger.Info("DeletionTimestamp is set on %s", support)
				controllerutil.RemoveFinalizer(&support, v1alpha1.ResourcesFinalizerName)
				if err := r.Status().Update(ctx, &support); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}
	*/
	_ = log.FromContext(ctx)

	if support.ObjectMeta.Generation == support.Status.ObservedGeneration {
		return ctrl.Result{}, nil
	}
	for _, wf := range support.Spec.Workflows {

		if support.Status.LastTransitionTime != nil && wf.InitiatedAt.After(support.Status.LastTransitionTime.Add(-15*time.Minute)) {
			continue
		}
		now := metav1.Now()
		support.Status.LastTransitionTime = &now
		support.Status.Phase = supportv1alpha1.ArgoSupportPhaseRunning

		// Pass support as an argument to getWfExecutor
		wfExecutor, err := r.getWfExecutor(ctx, &wf, &support)

		if wfExecutor != nil {
			// Pass support as an argument to wfExecutor.Process
			obj, err := wfExecutor.Process(ctx, &support)
			if err != nil {
				logger.Info("Failed to process workflow")
				support.Status.Phase = supportv1alpha1.ArgoSupportPhaseFailed
				continue
			}

			if obj != nil && len(obj.Status.Results) > 1 {
				sort.SliceStable(obj.Status.Results, func(i, j int) bool {
					return (obj.Status.Results[i].FinishedAt.Time).After(obj.Status.Results[j].FinishedAt.Time)
				})

				if len(obj.Status.Results) > 2 {
					support.Status.Results = obj.Status.Results[:2]
				}
			}
			now := metav1.Now()
			support.Status.LastTransitionTime = &now
			support.Status.ObservedGeneration = support.ObjectMeta.Generation
		} else {
			support.Status.Phase = supportv1alpha1.ArgoSupportPhaseFailed
			logger.Error(err, "Failed to get workflow executor")
		}
	}

	// Update support object in Kubernetes with latest status
	if err := r.Status().Update(ctx, &support); err != nil {
		logger.Error(err, "Failed to update Argo AI support status to completed")
		return ctrl.Result{}, err
	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

func (r *SupportReconciler) getWfExecutor(ctx context.Context, wf *supportv1alpha1.Workflow, obj metav1.Object) (wf_operations.Executor, error) {

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

func (r *SupportReconciler) handleFinalizer(ctx context.Context, ops *supportv1alpha1.Support) error {
	// name of our custom finalizer

	// examine DeletionTimestamp to determine if object is under deletion
	if ops.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(ops, supportv1alpha1.FinalizerName) {
			controllerutil.AddFinalizer(ops, supportv1alpha1.FinalizerName)
			if err := r.Update(ctx, ops); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(ops, supportv1alpha1.FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(ops, supportv1alpha1.FinalizerName)
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
		For(&supportv1alpha1.Support{}).
		Complete(r)
}
