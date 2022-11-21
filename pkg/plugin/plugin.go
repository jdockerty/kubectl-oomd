package plugin

import (
	"fmt"

	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// TerminatedPodInfo is a wrapper struct around an OOMKilled Pod's information.
type TerminatedPodInfo struct {
	Pod            v1.Pod
	Memory         MemoryInfo
	ContainerName  string // Name of the container within the pod that was terminated, in the case of multi-container pods.
	TerminatedTime string // When the pod was terminated
	StartTime      string // When the pod was started during the termination period.
}

// MemoryInfo is the container resource requests, specific to the memory limit and requests.
type MemoryInfo struct {
	Request string
	Limit   string
}

// GetNamespace will retrieve the current namespace from the provided namespace or kubeconfig file of the caller.
func GetNamespace(configFlags *genericclioptions.ConfigFlags, all bool, givenNamespace string) (string, error) {

	if givenNamespace == "" && all {
		return metav1.NamespaceAll, nil
	} else if givenNamespace == "" {
		// Retrieve the current namespace from the raw kubeconfig struct
		currentNamespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return "", fmt.Errorf("failed to during creating raw kubeconfig: %w", err)
		}
		return currentNamespace, nil
	}

	return givenNamespace, nil
}

// TerminatedPodsFilter is used to filter for pods that contain a terminated container, with an exit code of 137 (OOMKilled).
func TerminatedPodsFilter(pods []v1.Pod) []v1.Pod {

	var terminatedPods []v1.Pod

	for _, pod := range pods {
		for _, containerStatus := range pod.Status.ContainerStatuses {

			// The terminated state may be nil, i.e. not terminated, we must check this first.
			if terminated := containerStatus.LastTerminationState.Terminated; terminated != nil {
				if terminated.ExitCode == 137 {
					terminatedPods = append(terminatedPods, pod)
				}
			}
		}
	}

	return terminatedPods
}

// BuildTerminatedPodsInfo retrieves the terminated pod information, bundled into a slice of the informational struct.
func BuildTerminatedPodsInfo(client *kubernetes.Clientset, namespace string) ([]TerminatedPodInfo, error) {

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	terminatedPods := TerminatedPodsFilter(pods.Items)

	var terminatedPodsInfo []TerminatedPodInfo

	for _, pod := range terminatedPods {

		for containerIndex, containerStatus := range pod.Status.ContainerStatuses {

			// Not every container within the pod will be in a terminated state, we skip these ones.
			// This also means we can use the 'containerIndex' to directly access the correct container,
			// as we know its index within the container status list.
			if containerStatus.LastTerminationState.Terminated == nil {
				continue
			}

			startTime := pod.Status.ContainerStatuses[containerIndex].LastTerminationState.Terminated.StartedAt.String()
			terminatedTime := pod.Status.ContainerStatuses[containerIndex].LastTerminationState.Terminated.FinishedAt.String()

			// Build our terminated pod info struct
			info := TerminatedPodInfo{
				Pod:            pod,
				ContainerName:  containerStatus.Name,
				StartTime:      startTime,
				TerminatedTime: terminatedTime,
				Memory: MemoryInfo{
					Limit:   pod.Spec.Containers[containerIndex].Resources.Limits.Memory().String(),
					Request: pod.Spec.Containers[containerIndex].Resources.Requests.Memory().String(),
				},
			}

			// TODO: Since we know all pods here have been in the "terminated state", can we
			// achieve this same result in an elegant way?
			terminatedPodsInfo = append(terminatedPodsInfo, info)
		}
	}

	return terminatedPodsInfo, nil
}

// Runreturns the pod information for those that have been OOMKilled, this provides the plugin functionality.
func Run(configFlags *genericclioptions.ConfigFlags, allNamespaces bool, providedNamespace string) ([]TerminatedPodInfo, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, err := GetNamespace(configFlags, allNamespaces, providedNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve namespace, got %s: %w", providedNamespace, err)
	}

	terminatedPods, err := BuildTerminatedPodsInfo(clientset, namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to build terminated pod information: %w", err)
	}

	return terminatedPods, nil
}
