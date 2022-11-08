package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

func TestGetNamespace(t *testing.T) {

	testNamespace := "test-namespace"
	tests := map[string]struct {
		namespace string
		want      string
	}{
		"should return given namespace":   {namespace: "my-ns", want: "my-ns"},
		"should return current namespace": {namespace: "", want: testNamespace},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// Hardcode current namespace for test purposes
			if *KubernetesConfigFlags.Namespace == "" {
				*KubernetesConfigFlags.Namespace = testNamespace
			}

			ns, err := GetNamespace(KubernetesConfigFlags, tc.namespace)
			assert.Nil(t, err)

			assert.Equal(t, tc.want, ns)
		})
	}
}

func TestGetAllNamespaces(t *testing.T) {

    if testing.Short() {
        t.Skip("skipping test which requires valid kubeconfig file in short mode")
    }

    namespaces, err := GetAllNamespaces(KubernetesConfigFlags)
    assert.Nil(t, err)

    assert.True(t, len(namespaces) != 0)

}
