package genai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/argo-support/api/v1alpha1"
	"github.com/argoproj-labs/argo-support/internal/services/ai_provider"
	"github.com/argoproj-labs/argo-support/internal/services/ai_provider/genstudio"

	"github.com/argoproj-labs/argo-support/internal/utils"
	"github.com/argoproj-labs/argo-support/internal/wf_operations"
	"github.com/argoproj-labs/argo-support/internal/wf_operations/common"
	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/go-logr/logr"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

const (
	genAIEndPointSuffix  = "/analyze"
	argoV1ResAPI         = "/api/v1/applications"
	responseType         = "/analyses"
	argocdEndPointSuffix = "/api/v1/applications/"
	rolloutRevision      = "rollout.argoproj.io/revision"
)

type GenAIOperator struct {
	k8sClient     client.Client
	genAIClient   ai_provider.AIProvider
	dynamicClient dynamic.DynamicClient
	argoCDClient  *ai_provider.ArgoCDClient
	kubeClient    kubernetes.Interface
	configMap     *v1.ConfigMap
}

var (
	_ wf_operations.Executor = &GenAIOperator{}
)

func NewGenAIOperations(ctx context.Context, k8sClient client.Client, dynamicClient dynamic.DynamicClient, kubeClient kubernetes.Interface, wf *v1alpha1.Workflow, namespace string) (*GenAIOperator, error) {
	authProviders, err := getAuthProviders(ctx, k8sClient, &wf.Ref, namespace)
	var genClient ai_provider.AIProvider
	var argoCDClient *ai_provider.ArgoCDClient

	cm, err := getCMRef(ctx, k8sClient, &wf.ConfigMapRef, namespace)
	if err != nil {
		return nil, err
	}

	for _, authProvider := range *authProviders {
		if &authProvider != nil {
			switch authProvider.Name {
			case "genai-auth-provider":
				genClient, err = getGenAIProvider(ctx, k8sClient, &authProvider, namespace)
				if err != nil {
					return nil, err
				}
			case "argocd-auth-provider":
				argoCDClient, err = getArgoCDClient(ctx, k8sClient, &authProvider, namespace)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &GenAIOperator{
		k8sClient:     k8sClient,
		genAIClient:   genClient,
		argoCDClient:  argoCDClient,
		dynamicClient: dynamicClient,
		kubeClient:    kubeClient,
		configMap:     cm,
	}, nil
}

func getAuthProviders(ctx context.Context, k8sClient client.Client, refs *[]v1alpha1.NamespacedObjectReference, namespace string) (*[]v1alpha1.AuthProvider, error) {
	return utils.GetAuthProviders(ctx, k8sClient, refs, namespace)
}

func getCMRef(ctx context.Context, k8sClient client.Client, configMapRef *v1alpha1.ConfigMapRef, namespace string) (*v1.ConfigMap, error) {
	return utils.GetConfigMapRef(ctx, k8sClient, configMapRef, namespace)
}
func getGenAIProvider(ctx context.Context, k8sClient client.Client, authProviders *v1alpha1.AuthProvider, namespace string) (ai_provider.AIProvider, error) {

	switch *authProviders.Spec.Provider {
	case "genstudio":
		return genstudio.GenAIClientWithSecret(ctx, k8sClient, authProviders, namespace), nil
	default:
		return nil, fmt.Errorf("unsupported genai provider: %s", authProviders.Spec.Provider)
	}
}

func getArgoCDClient(ctx context.Context, k8sClient client.Client, authProvider *v1alpha1.AuthProvider, namespace string) (*ai_provider.ArgoCDClient, error) {
	return ai_provider.GetArgoCDClienWithSecret(ctx, k8sClient, authProvider, namespace)
}

func (g *GenAIOperator) Process(ctx context.Context, obj metav1.Object) (*v1alpha1.Support, error) {
	logger := log.FromContext(ctx)
	support := obj.(*v1alpha1.Support)
	if support == nil {
		return nil, fmt.Errorf("failed to process: recieved nil genai object")
	}
	var app *common.Application
	var err error
	if support != nil && support.Annotations[v1alpha1.ArgoSupportGenAIAnnotationKey] == "" {
		support.Status.Phase = v1alpha1.ArgoSupportPhaseRunning
		return support, nil
	} else {
		annotationValue := support.Annotations[v1alpha1.ArgoSupportGenAIAnnotationKey]
		annotationValue = strings.ReplaceAll(annotationValue, "\n", "")
		if annotationValue != "" {
			err := json.Unmarshal([]byte(annotationValue), &app)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal annotation: %v", err)
			}
		}
	}

	if app == nil {
		labels := obj.GetLabels()
		fullUrl := fmt.Sprint(g.argoCDClient.BaseURL + argocdEndPointSuffix + labels["app.kubernetes.io/instance"])

		app, err = g.argoCDClient.GetArgoApp(fullUrl)
		if err != nil {
			logger.Error(err, "unable to connect to argocd instance", "argocd url", fullUrl)
		}
	}

	t, _ := g.buildAITokens(ctx, app, obj)

	failures := common.Failures{
		Failures: []common.Failure{
			{Context: t},
		},
	}

	// Marshal the struct to JSON
	tokens, _ := json.Marshal(failures)
	logger.Info("tokens to be processed", "tokens length", len(tokens))

	res, err := g.genAIClient.SubmitTokensToGenAI(ctx, string(tokens), genAIEndPointSuffix)
	if err != nil {
		return nil, fmt.Errorf("failed to post request: %v", err)
	}

	summary, ok := res.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("type assertion to map[string]interface{} failed")
	}

	value, exists := summary["analyses"]
	if !exists {
		return nil, fmt.Errorf("key 'analyses' not found in the result")
	}

	analysesSlice, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("type assertion for 'analyses' as []interface{} failed")
	}

	// Assuming that each element in analysesSlice is a map[string]interface{} that contains an "analysis" key
	var genSummary string
	for _, analysis := range analysesSlice {
		analysisMap, ok := analysis.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("type assertion for individual analysis failed")
		}
		genSummary, ok = analysisMap["analysis"].(string)
		if !ok {
			return nil, fmt.Errorf("type assertion for 'analysis' as string failed")
		}
		break
	}

	argoOpsobj, ok := obj.(*v1alpha1.Support)
	if !ok {
		return nil, fmt.Errorf("type assertion to *v1alpha1.ArgoSupportSpec failed")
	}
	var slackSupport string
	slackSupport, _ = g.configMap.Data["slackSupport"]

	help := v1alpha1.Help{
		SlackChannel: slackSupport,
	}
	now := metav1.Now()
	epochTime := now.Unix()

	argoOpsobj.Status = v1alpha1.SupportStatus{
		Results: append(argoOpsobj.Status.Results, v1alpha1.Result{
			Name: fmt.Sprintf("%s-%d", argoOpsobj.Spec.Workflows[0].Name, epochTime),
			Summary: v1alpha1.Summary{
				MainSummary: genSummary,
			},
			Help:       help,
			FinishedAt: &now,
			Message:    "Gen AI request completed",
		}),
		Phase: v1alpha1.ArgoSupportPhaseCompleted,
	}
	return argoOpsobj, nil
}

func (g *GenAIOperator) buildAITokens(ctx context.Context, app *common.Application, o metav1.Object) (string, error) {
	logger := log.FromContext(ctx)
	var builder strings.Builder
	builder.WriteString(utils.GetInlinePrompt("main_instructions", ""))

	if app != nil {
		g.processApplicationStatus(&builder, app, logger)
	} else {
		logger.Info("app info seems to be missing and should not be included in the analysis")
	}

	rolloutLister := rolloutListFromClient(g.dynamicClient)
	res, err := rolloutLister(o.GetNamespace(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, r := range res {
		g.processRollout(&builder, r, o, logger)
	}

	g.collectEventData(ctx, &builder, o, logger)
	builder.WriteString(utils.GetInlinePrompt("end_instructions", ""))

	return builder.String(), nil
}

func (g *GenAIOperator) processApplicationStatus(builder *strings.Builder, app *common.Application, logger logr.Logger) {
	if app.Status.Health.Status == common.HealthStatusHealthy && len(app.Status.Conditions) == 0 {
		builder.WriteString(utils.GetInlinePrompt("app-healthy", ""))
		return
	}

	builder.WriteString(utils.GetInlinePrompt("app-conditions", ""))
	for _, condition := range app.Status.Conditions {
		logger.Info("include app Conditions", "application name", app.Name)
		builder.WriteString(fmt.Sprintf("Condition Message: %s, Status: %s, LastTransitionTime: %s\n", condition.Type, condition.Message, condition.LastTransitionTime))
	}

	for _, res := range app.Status.Resources {
		builder.WriteString(utils.GetInlinePrompt("non-healthy-res", ""))
		if res.Health != nil && res.Health.Status != common.HealthStatusHealthy {
			builder.WriteString(fmt.Sprintf("Resource Name: %s Resource Health: %s  and kubernetes Message: %s", res.Name, res.Health.Status, res.Health.Message))
		}
	}
}

func (g *GenAIOperator) processRollout(builder *strings.Builder, r *rolloutv1alpha1.Rollout, o metav1.Object, logger logr.Logger) {

	builder.WriteString(utils.GetInlinePrompt("rollout", r.Name))
	if rollout, ok := utils.StripTheKeys(r).(*rolloutv1alpha1.Rollout); ok {
		pods, _ := getPodsWithLabel(g.k8sClient, r.Status.CurrentPodHash)
		g.processPods(builder, pods, r.Namespace, logger)
		g.processAnalysisRuns(builder, o, r, logger)
		builder.WriteString(rollout.Status.String())
	}
}

func (g *GenAIOperator) processPods(builder *strings.Builder, podList []string, namespace string, logger logr.Logger) {
	if len(podList) == 0 {
		builder.WriteString(utils.GetInlinePrompt("no-pod-log", ""))
		return
	}

	logs, err := getLogsForPod(podList[0], namespace, g.kubeClient)
	if err != nil {
		logger.Info("no pod logs to process")
	} else {
		builder.WriteString(utils.GetInlinePrompt("logs-with-error", ""))
		builder.WriteString(logs)
	}

	podStatus, err := getPodStatus(podList[0], namespace, g.kubeClient)
	if err != nil {
		logger.Info("no pod status to process")
		return
	}

	builder.WriteString(utils.GetInlinePrompt("podContainerStatus", ""))
	for _, containerStatus := range podStatus.ContainerStatuses {
		builder.WriteString(fmt.Sprintf("Container Name: %s, started: %t, State: %s, Ready: %t, Restart Count: %d\n",
			containerStatus.Name, containerStatus.Started, containerStatus.State, containerStatus.Ready, containerStatus.RestartCount))
	}

	builder.WriteString(utils.GetInlinePrompt("podInitContainerStatus", ""))
	for _, containerStatus := range podStatus.InitContainerStatuses {
		builder.WriteString(fmt.Sprintf("Container Name: %s, started: %t, State: %s, Ready: %t, Restart Count: %d\n",
			containerStatus.Name, containerStatus.Started, containerStatus.State, containerStatus.Ready, containerStatus.RestartCount))
	}
}

func (g *GenAIOperator) processAnalysisRuns(builder *strings.Builder, o metav1.Object, r *rolloutv1alpha1.Rollout, logger logr.Logger) {
	analysisLister := analysisListFromClient(g.dynamicClient)
	aRuns, err := analysisLister(o.GetNamespace(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "failed to list analysis runs")
		return
	}

	for i, ar := range aRuns {
		if o.GetAnnotations()[rolloutRevision] != ar.Annotations[rolloutRevision] {
			aRuns = append(aRuns[:i], aRuns[i+1:]...)
		}
	}

	if len(aRuns) > 0 {
		builder.WriteString(utils.GetInlinePrompt("analysis-runs", ""))
		builder.WriteString(aRuns[0].Status.String())
	} else {
		logger.Info("no analysisrun for given rollout", "rollout revision", o.GetAnnotations()[rolloutRevision])
	}
}

func (g *GenAIOperator) collectEventData(ctx context.Context, builder *strings.Builder, o metav1.Object, logger logr.Logger) {
	logger.Info("start collecting event data")
	var eventList v1.EventList
	err := g.k8sClient.List(ctx, &eventList, client.InNamespace(o.GetNamespace()))
	if err != nil {
		logger.Error(err, "np events to fetch for namespace %s", o.GetNamespace())
		return
	}
	if len(eventList.Items) > 0 {
		builder.WriteString(utils.GetInlinePrompt("events", ""))
		for _, event := range eventList.Items {
			if event.Message == "Warning" || strings.Contains(event.Message, "Failed") {
				builder.WriteString(event.String())
			}
		}
	}
}

func genericListFromClient(c dynamic.DynamicClient, gvr schema.GroupVersionResource) func(string, metav1.ListOptions) ([]*unstructured.Unstructured, error) {
	return func(namespace string, options metav1.ListOptions) ([]*unstructured.Unstructured, error) {
		res, err := c.Resource(gvr).Namespace(namespace).List(context.Background(), options)
		if err != nil {
			return nil, err
		}
		var resourceList []*unstructured.Unstructured
		for i := range res.Items {
			resourceList = append(resourceList, &res.Items[i])
		}
		return resourceList, nil
	}
}

type rolloutListFunc func(namespace string, options metav1.ListOptions) ([]*rolloutv1alpha1.Rollout, error)
type analysisListFunc func(namespace string, options metav1.ListOptions) ([]*rolloutv1alpha1.AnalysisRun, error)

func rolloutListFromClient(c dynamic.DynamicClient) rolloutListFunc {
	genericLister := genericListFromClient(c, v1alpha1.SchemeGroupVersion.WithResource("rollouts"))
	return func(namespace string, options metav1.ListOptions) ([]*rolloutv1alpha1.Rollout, error) {
		unstructuredList, err := genericLister(namespace, options)
		if err != nil {
			return nil, err
		}
		var rolloutList []*rolloutv1alpha1.Rollout
		for _, unstructuredRollout := range unstructuredList {
			rollout := &rolloutv1alpha1.Rollout{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredRollout.Object, rollout)
			if err != nil {
				return nil, err
			}
			rolloutList = append(rolloutList, rollout)
		}
		return rolloutList, nil
	}
}

func analysisListFromClient(c dynamic.DynamicClient) analysisListFunc {
	genericLister := genericListFromClient(c, v1alpha1.SchemeGroupVersion.WithResource("analysisruns"))
	return func(namespace string, options metav1.ListOptions) ([]*rolloutv1alpha1.AnalysisRun, error) {
		unstructuredList, err := genericLister(namespace, options)
		if err != nil {
			return nil, err
		}
		var analysisRunList []*rolloutv1alpha1.AnalysisRun
		for _, unstructuredAnalysisRun := range unstructuredList {
			analysisRun := &rolloutv1alpha1.AnalysisRun{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredAnalysisRun.Object, analysisRun)
			if err != nil {
				return nil, err
			}
			analysisRunList = append(analysisRunList, analysisRun)
		}
		return analysisRunList, nil
	}
}

func getPodsWithLabel(K8sClient client.Client, label string) ([]string, error) {
	podList := &v1.PodList{}

	// Create a ListOption with LabelSelector
	listOpts := []client.ListOption{
		client.MatchingLabels{"rollouts-pod-template-hash": label},
	}
	if err := K8sClient.List(context.TODO(), podList, listOpts...); err != nil {
		return nil, err
	}
	var podNames []string
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}
	return podNames, nil
}
func getLogsForPod(podName, namespace string, kubeClient kubernetes.Interface) (string, error) {
	// Fetch logs from the pod
	podLogOpts := v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)

	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("could not fetch logs: %v", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, podLogs); err != nil {
		return "", fmt.Errorf("could not read logs: %v", err)
	}

	logs := buf.String()
	lines := strings.Split(logs, "\n")
	for i, line := range lines {
		if strings.Contains(line, "error") {
			start := maxLine(0, i-5)
			end := minLine(len(lines), i+6)
			return strings.Join(lines[start:end], "\n"), nil
		}
	}
	return "", nil
}
func getPodStatus(podName, namespace string, kubeClient kubernetes.Interface) (*v1.PodStatus, error) {
	pod, err := kubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get pod: %v", err)
	}

	return &pod.Status, nil
}

func minLine(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxLine(a, b int) int {
	if a > b {
		return a
	}
	return b
}
