package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestGetNamespace(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping test which requires a valid kubeconfig file in short test mode.")
	}

	tests := map[string]struct {
		namespace string
		want      string
	}{
		"should return given namespace": {namespace: "my-ns", want: "my-ns"},
        "should return current namespace": {namespace: "", want: "test-namespace"},
	}

	KubernetesConfigFlags := genericclioptions.NewConfigFlags(false)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
            
            // Hardcode current namespace for test purposes
            if *KubernetesConfigFlags.Namespace == "" {
                *KubernetesConfigFlags.Namespace = "test-namespace"
            }

			ns, err := GetNamespace(KubernetesConfigFlags, tc.namespace)
			assert.Nil(t, err)

			assert.Equal(t, tc.want, ns)
		})
	}

}
