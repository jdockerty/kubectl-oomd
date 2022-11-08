package plugin

import (
	"fmt"

	"github.com/jdockerty/kubectl-oomd/pkg/logger"
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
func GetNamespace(configFlags *genericclioptions.ConfigFlags, givenNamespace string) (string, error) {

	if givenNamespace == "" {

		// Retrieve the current namespace from the raw kubeconfig struct
		currentNamespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return "", fmt.Errorf("failed to during creating raw kubeconfig: %w", err)
		}
		return currentNamespace, nil
	}

	return givenNamespace, nil
}

// RunPlugin returns the pod information for those that have been OOMKilled, this provides the plugins' functionality.
func RunPlugin(configFlags *genericclioptions.ConfigFlags, namespace string, logger *logger.Logger) ([]TerminatedPodInfo, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	if namespace == "" {
		// Retrieve the current namespace from the raw kubeconfig struct
		currentNamespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return nil, fmt.Errorf("failed to during creating raw kubeconfig: %w", err)
		}
		namespace = currentNamespace
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var terminatedPodsInfo []TerminatedPodInfo

	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {

			// The terminated state may be nil, i.e. not terminated, we must check this first.
			if terminated := containerStatus.LastTerminationState.Terminated; terminated != nil {
				if terminated.ExitCode == 137 {

					var containerIndex int // The container which was OOMKilled, it may not always be the 0th index.

					for i, c := range pod.Spec.Containers {
						// Found OOMKilled container
						if containerStatus.Name == c.Name {
							containerIndex = i
							break
						}
					}

					info := TerminatedPodInfo{
						Pod:            pod,
						ContainerName:  containerStatus.Name,
						StartTime:      terminated.StartedAt.String(),
						TerminatedTime: terminated.FinishedAt.String(),
						Memory: MemoryInfo{
							Limit:   pod.Spec.Containers[containerIndex].Resources.Limits.Memory().String(),
							Request: pod.Spec.Containers[containerIndex].Resources.Requests.Memory().String(),
						},
					}
					terminatedPodsInfo = append(terminatedPodsInfo, info)
				}
			}
		}
	}

	return terminatedPodsInfo, nil
}
