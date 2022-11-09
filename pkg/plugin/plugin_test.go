package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

func TestGetNamespace(t *testing.T) {

	testNamespace := "test-namespace"
	tests := map[string]struct {
		namespace string
		all       bool
		want      string
	}{
		"should return given namespace":   {namespace: "my-ns", all: false, want: "my-ns"},
		"should return current namespace": {namespace: "", all: false, want: testNamespace},
		"should return all namespaces":    {namespace: "", all: true, want: metav1.NamespaceAll},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// Hardcode current namespace for test purposes
			if *KubernetesConfigFlags.Namespace == "" {
				*KubernetesConfigFlags.Namespace = testNamespace
			}

			ns, err := GetNamespace(KubernetesConfigFlags, tc.all, tc.namespace)
			assert.Nil(t, err)

			assert.Equal(t, tc.want, ns)
		})
	}
}

func TestFilterTerminatedPods(t *testing.T) {
	testPods := []v1.Pod{
		v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "oomedPod",
			},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					v1.ContainerStatus{
						LastTerminationState: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								ExitCode: 137,
								Reason:   "OOMKilled",
							},
						},
					},
				},
			},
		},
		v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "okayPod",
			},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					v1.ContainerStatus{
						LastTerminationState: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								ExitCode: 0,
								Reason:   "Completed",
							},
						},
					},
				},
			},
		},
	}

	oomed := TerminatedPodsFilter(testPods)

	assert.Equal(t, 1, len(oomed))
	assert.Equal(t, "oomedPod", oomed[0].Name)

}
