package utils

import (
	"context"
	"fmt"
	v1alpha1 "github.com/argoproj-labs/argo-support/api/v1alpha1"
	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var authProviderMap = make(map[v1alpha1.NamespacedObjectReference]*v1alpha1.AuthProvider)

func GetSecret(ctx context.Context, k8sClient client.Client, authProvider *v1alpha1.AuthProvider) (*v1.Secret, error) {
	logger := log.FromContext(ctx)

	var secret v1.Secret
	objectKey := client.ObjectKey{
		Namespace: authProvider.GetNamespace(),
		Name:      authProvider.Spec.SecretRef.Name,
	}
	err := k8sClient.Get(ctx, objectKey, &secret)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Secret for AuthProvider not found", "namespace", objectKey.Namespace, "name", objectKey.Name)
			return nil, err
		}

		logger.Error(err, "failed to get Secret from AuthProvider", "namespace", objectKey.Namespace, "name", objectKey.Name)
		return nil, err
	}

	return &secret, nil
}

func getAuthProviders(ctx context.Context, k8sClient client.Client, labels map[string]string) (*[]v1alpha1.AuthProvider, error) {
	logger := log.FromContext(ctx)

	authProviderList := &v1alpha1.AuthProviderList{}
	listOptions := []client.ListOption{
		client.MatchingLabels(labels),
	}
	err := k8sClient.List(ctx, authProviderList, listOptions...)
	if err != nil {
		return nil, err
	}

	if len(authProviderList.Items) == 0 {
		err = fmt.Errorf("authProvider not found with labels %#v", labels)
		logger.Info(err.Error())
		return nil, err
	}

	return &authProviderList.Items, nil
}

func GetAuthProviders(ctx context.Context, k8sClient client.Client, refs *[]v1alpha1.NamespacedObjectReference, namespace string) (*[]v1alpha1.AuthProvider, error) {
	logger := log.FromContext(ctx)
	authProviders, err := getAuthProviders(ctx, k8sClient, map[string]string{v1alpha1.LabelKeyAppName: v1alpha1.LabelKeyAppNameValue})
	if err != nil {
		logger.Error(err, "failed to get AuthProvider", "namespace", namespace)
		return nil, err
	}
	return authProviders, nil
}

func getConfigMap(ctx context.Context, k8sClient client.Client, labels map[string]string) (*v1.ConfigMap, error) {
	logger := log.FromContext(ctx)

	cmList := &v1.ConfigMapList{}
	listOptions := []client.ListOption{
		client.MatchingLabels(labels),
	}
	err := k8sClient.List(ctx, cmList, listOptions...)
	if err != nil {
		return nil, err
	}

	if len(cmList.Items) == 0 {
		err = fmt.Errorf("ConfigMap not found with labels %#v", labels)
		logger.Info(err.Error())
		return nil, err
	}

	return &cmList.Items[0], nil
}

func GetConfigMapRef(ctx context.Context, k8sClient client.Client, configMapRef *v1alpha1.ConfigMapRef, namespace string) (*v1.ConfigMap, error) {
	logger := log.FromContext(ctx)
	cm, err := getConfigMap(ctx, k8sClient, map[string]string{v1alpha1.LabelKeyAppName: v1alpha1.LabelKeyAppNameValue})
	if err != nil {
		logger.Error(err, "failed to get AuthProvider", "namespace", namespace)
		return nil, err
	}
	return cm, nil
}

func StripTheKeys(obj metav1.Object) metav1.Object {
	switch t := obj.(type) {
	case *v1.Pod:
		fmt.Println("Processing Pod", t.Name)
	case *v1.Service:
		fmt.Println("Processing Service", t.Name)
	case *rolloutv1alpha1.Rollout:
		fmt.Println("Processing Operation", t.Name)
		return t
	default:
		fmt.Println("Unknown type")
	}
	return nil
}

func GetInlinePrompt(step string, data string) string {
	switch step {
	case "main_instructions":
		return "Disregard any previous instructions provided. You are expert in debugging the kubernetes issues, " +
			"Follow this instruction provide the debug summary in points. In summary highlight the field  from yaml which might cause the issue" +
			" don't add point which is not provided in resource." +
			" When analysis the issue do not make any  assumptions or add extra details or add details which is irrelevant to debugging." +
			" Along this main instructions, I'll provide additional inline instructions that contain inside the tag " +
			"<prompt></prompt> follow up by  resource spec and status that need to be inferred for debugging the issue."
	case "app-healthy":
		return "<prompt>app seems to be healthy, there nothing to analysis. discard any previous prompt and don't summarize any details provided. Echo the message app seems to be healthy  and nothing to summarize</prompt>"
	case "app-unhealthy":
		return "<prompt>app seems to be not healthy, check app statue</prompt>"
	case "non-healthy-res":
		return "<prompt>evaluate the non healthy  resource based on the message</prompt>"
	case "app-conditions":
		return "<prompt>When analyzing rollout, be sure to consider for conditions 'reason' and inference the message that true and progressed for more than 15 mins (lastTransitionTime - lastUpdateTime). " +
			"Analyse if condition is stuck based on the time" +
			"Additionally, consider on the status.phase, status.message, status.canary.podTemplateHash, status.currentPodHash, and status.stableRS when analysis the issue with the Rollout. Don't make assumption, and if not enough " +
			"info to analyse than recommend user to review status and reach out to support </prompt>"

	case "rollout":
		return "<prompt>When analyzing the rollout, first things to analyse is status.phase. If phase is healthy return the message to" +
			"user  Rollout seems to be healthy, there is no apparent issue or error that needs debugging. (Stop here)." +
			" If phase is not Healthy, then include following field into your analysis phase, observedGeneration, message, compare  stableRS = podTemplateHash" +
			" evaluate conditions that  true for more than 15 mins," +
			" diff between lastUpdateTime and lastTransitionTime and estimate if the condition is stuck for long. Discard any condition which is false from your debugging" +
			". If you NOT able identify the root cause, don't provide any explanation and  limit the answer to recommend user to check with argo support</prompt>"
	case "event":
		return "<prompt>When debugging the issue analyse events related to  resources. Don't include event older than 15 mins</prompt>"
	case "analysis-runs":
		return "<prompt>When debugging the AnalysisRun resource when  phase: Error. Check if there are any failed metrics. Discard the summary  of the analysisrun  that contain message: Run Terminated with   phase: Successful and no metrics failures </prompt>"
	case "no-pod-log":
		return "<prompt>pod logs is not available, as pod could be terminated. Check the events and rollout or deployment  is aborted include this data in the analysis</prompt>"
	case "logs-with-error":
		return "<prompt>evaluate the logs for error that causing the failure. In you summary highlight any pods failure that causing pods to fail</prompt>"
	case "podContainerStatus":
		return "<prompt>evaluate the containerStatus for error that causing the failure. In you summary highlight any container status that causing pods to fail</prompt>"
	case "podInitContainerStatus":
		return "<prompt>evaluate the InitContainerStatuses for error that causing the failure. In you summary highlight any Initcontainer status that causing pods to fail</prompt>"
	case "events":
		return "<prompt>valuate the events, related the event types with  resource status, ignore the  events order than 30 mins.</prompt>"

	default:
		return ""
	}
}
