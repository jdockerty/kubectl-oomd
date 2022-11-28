package plugin

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
)

var KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

const (

	// The namespace to use with tests that do not require access to a working cluster.
	testNamespace = "test-namespace"

	// The namespace defined for use with the `oomer` manifest.
	integrationTestNamespace = "oomkilled"

	// A manifest which contains the `oomkilled` namespace and `oomer` deployment for testing purposes.
	forceOOMKilledManifest = "https://raw.githubusercontent.com/jdockerty/oomer/main/oomer.yaml"
)

func setupIntegrationTestDependencies(t *testing.T) {

	err := applyOrDeleteOOMKilledManifest(false)
	if err != nil {
		t.Fatalf("unable to apply OOMKilled manifest: %s", err)
	}
	defer applyOrDeleteOOMKilledManifest(true)

	t.Log("Waiting 20 seconds for pods to start being OOMKilled...")
	time.Sleep(20 * time.Second)
}

// applyOrDeleteOOMKilledManifest is a rather hacky way to utilise `kubectl` within
// our test to apply the oomer manifest, this means we do not have to use a large
// setup function to parse the give manifest and apply it using a kubernetes.clientset.
func applyOrDeleteOOMKilledManifest(runDelete bool) error {

	if runDelete {
		os.Args = []string{"kubectl", "delete", "-f", forceOOMKilledManifest}
	} else {
		os.Args = []string{"kubectl", "apply", "-f", forceOOMKilledManifest}
	}

	err := cmd.NewDefaultKubectlCommand().Execute()
	if err != nil {
		return err
	}
	return nil
}

// TestRunPlugin tests against an initialised cluster with OOMKilled pods that
// the plugin's functionality works as expected.
func TestRunPlugin(t *testing.T) {

	if testing.Short() {
		t.Skipf("skipping %s which requires running cluster", t.Name())
	}

	setupIntegrationTestDependencies(t)

	pods, err := Run(KubernetesConfigFlags, integrationTestNamespace)
	assert.Nil(t, err)

	assert.Greater(t, len(pods), 0, "expected number of failed pods to be greater than 0, got %d", len(pods))

}

func TestGetNamespace(t *testing.T) {

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
