package plugin

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/scheme"
)

var KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

const (

	// The namespace to use with tests that do not require access to a working cluster.
	testNamespace = "test-namespace"

	// A manifest which contains the `oomkilled` namespace and `oomer` deployment for testing purposes.
	forceOOMKilledManifest = "https://raw.githubusercontent.com/jdockerty/oomer/main/oomer.yaml"
)

type RequiresClusterTests struct {
	suite.Suite

	// The namespace defined for use with the `oomer` manifest.
	IntegrationTestNamespace string
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

// getMemoryRequestAndLimitFromDeploymentManifest is a helper function for retrieving the
// string representation of a container's memory limit and request, such as "128Mi", from
// its own manifest.
func getMemoryRequestAndLimitFromDeploymentManifest(r io.Reader, containerIndex int) (string, string, error) {

	multir := k8syaml.NewYAMLReader(bufio.NewReader(r))

	for {

		buf, err := multir.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", "", err
		}

		obj, gvk, err := scheme.Codecs.UniversalDeserializer().Decode(buf, nil, nil)
		if err != nil {
			return "", "", err
		}

		switch gvk.Kind {
		case "Deployment":
			d := obj.(*appsv1.Deployment)
			request := d.Spec.Template.Spec.Containers[containerIndex].Resources.Requests["memory"]
			limit := d.Spec.Template.Spec.Containers[containerIndex].Resources.Limits["memory"]
			return request.String(), limit.String(), nil

		}
	}

	return "", "", fmt.Errorf("unable to get Kind=Deployment from manifest")
}

func (rc *RequiresClusterTests) SetupTest() {
	rc.IntegrationTestNamespace = "oomkilled"
}

func (rc *RequiresClusterTests) SetupSuite() {
	err := applyOrDeleteOOMKilledManifest(false)
	if err != nil {
		rc.T().Fatalf("unable to apply OOMKilled manifest: %s", err)
	}

	rc.T().Log("Waiting 30 seconds for pods to start being OOMKilled...")
	time.Sleep(30 * time.Second)
}

func (rc *RequiresClusterTests) TearDownSuite() {
	applyOrDeleteOOMKilledManifest(true)
}

// TestRunPlugin tests against an initialised cluster with OOMKilled pods that
// the plugin's functionality works as expected.
func (rc *RequiresClusterTests) TestRunPlugin() {
	pods, err := Run(KubernetesConfigFlags, rc.IntegrationTestNamespace)
	assert.Nil(rc.T(), err)

	assert.Greater(rc.T(), len(pods), 0, "expected number of failed pods to be greater than 0, got %d", len(pods))
}

func (rc *RequiresClusterTests) TestCorrectResources() {
	res, err := http.Get(forceOOMKilledManifest)
	if err != nil {
		// TODO Wrap these functions in a retry as they have reliance on external factors
		// which could cause a failure, e.g. here an issue with GitHub would cause an
		// error with getting the manifest.
		rc.T().Skipf("Skipping %s as could not perform GET request for the manifest", rc.T().Name())
	}
	defer res.Body.Close()

	// We can pass containerIndex 0 here as I control the manifest we are using, so it is
	// okay to hardcode it.
	manifestReq, manifestLim, err := getMemoryRequestAndLimitFromDeploymentManifest(res.Body, 0)
	assert.Nil(rc.T(), err) // We don't skip this on failure, as if we got the manifest it should be a Deployment.

	pods, _ := Run(KubernetesConfigFlags, rc.IntegrationTestNamespace)

	fmt.Println(manifestReq, manifestLim)
	podMemoryRequest := pods[0].Pod.Spec.Containers[0].Resources.Requests["memory"]
	podMemoryLimit := pods[0].Pod.Spec.Containers[0].Resources.Limits["memory"]

	assert.Equal(rc.T(), podMemoryRequest.String(), manifestReq)
	assert.Equal(rc.T(), podMemoryLimit.String(), manifestLim)
	rc.T().Logf(
		"\nMemory:\n\tManifest Request: %s\n\tManifest Limit: %s\n\tPod Request: %s\n\tPod Limit: %s\n",
		manifestReq,
		manifestLim,
		&podMemoryRequest,
		&podMemoryLimit,
	)

}

func TestRequiresClusterSuite(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s which requires running cluster", t.Name())
	}

	suite.Run(t, new(RequiresClusterTests))
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

func TestSortByTimestamp(t *testing.T) {

	now := time.Now()

	times := map[string]time.Time{
		"now": now,
		"1d":  now.AddDate(0, 0, 1),
		"2d":  now.AddDate(0, 0, 2),
		"1mo": now.AddDate(0, 1, 0),
	}

	// These are not in the descending order
	tests := TerminatedPods{
		TerminatedPodInfo{ContainerName: "1 month", terminatedTime: times["1mo"]},
		TerminatedPodInfo{ContainerName: "now", terminatedTime: times["now"]},
		TerminatedPodInfo{ContainerName: "2 days", terminatedTime: times["2d"]},
		TerminatedPodInfo{ContainerName: "1 day", terminatedTime: times["1d"]},
	}

	tests.SortByTimestamp()

	assert.Equal(t, tests[0].ContainerName, "now")
	assert.Equal(t, tests[1].ContainerName, "1 day")
	assert.Equal(t, tests[2].ContainerName, "2 days")
	assert.Equal(t, tests[3].ContainerName, "1 month")

	t.Log("Pods are in descending order.")

}
