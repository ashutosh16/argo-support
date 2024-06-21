package wf_operations

import (
	"context"
	"github.com/argoproj-labs/argo-support/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Executor interface {
	RunWorkflow(ctx context.Context, obj metav1.Object) error
	UpdateWorkflow(ctx context.Context, obj metav1.Object) (*v1alpha1.Feedback, error)
}
