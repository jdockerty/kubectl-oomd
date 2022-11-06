package plugin

import (
	"fmt"

	"github.com/jdockerty/kubectl-oomlie/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type TerminatedPodInfo struct {
	Pod            v1.Pod
	ContainerName  string // Name of the container within the pod that was terminated, in the case of multi-container pods.
	TerminatedTime string // When the pod was terminated
	StartTime      string // When the pod was started during the termination period.
}

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
		for _, cStatus := range pod.Status.ContainerStatuses {

			// The terminated state may be nil, i.e. not terminated, we must check this first.
			if terminated := cStatus.LastTerminationState.Terminated; terminated != nil {
				if terminated.ExitCode == 137 {

					info := TerminatedPodInfo{
						Pod:            pod,
						ContainerName:  cStatus.Name,
						StartTime:      terminated.StartedAt.String(),
						TerminatedTime: terminated.FinishedAt.String(),
					}

					terminatedPodsInfo = append(terminatedPodsInfo, info)
				}
			}
		}
	}

	return terminatedPodsInfo, nil
}
