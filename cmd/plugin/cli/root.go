package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jdockerty/kubectl-oomd/pkg/plugin"
	"github.com/jdockerty/kubectl-oomd/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
	noHeaders             bool
	allNamespaces         bool
	showVersion           bool

	// When using the namespace provided by the `--namespace/-n` flag or current context.
	// This represents: Pod, Container, Request, Limit, and Termination Time
	singleNamespaceFormatting = "%s\t%s\t%s\t%s\t%s\n"

	// When using the `all-namespaces` flag, we must show which namespace the pod was in, this becomes an extra column.
	// This represents: Namespace, Pod, Container, Request, Limit, and Termination Time
	allNamespacesFormatting = "%s\t%s\t%s\t%s\t%s\t%s\n"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubectl oomd",
		Short:         "Show pods which have recently been OOMKilled",
		Long:          `Show pods which have recently been terminated by Kubernetes due to an 'Out Of Memory' error`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			if showVersion {
				versionInfo := version.GetVersion()
				fmt.Printf("%s", versionInfo.ToString())
				return nil
			}

			namespace := *KubernetesConfigFlags.Namespace
			t := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0) // Formatting for table output, similar to other kubectl commands.

			oomPods, err := plugin.Run(KubernetesConfigFlags, allNamespaces, namespace)
			if err != nil {
				return errors.Unwrap(err)
			}

			// All namespaces flag requires an extra 'NAMESPACE' heading
			if allNamespaces {
				if !noHeaders {
					fmt.Fprintf(t, allNamespacesFormatting, "NAMESPACE", "POD", "CONTAINER", "REQUEST", "LIMIT", "TERMINATION TIME")
				}

				for _, p := range oomPods {
					fmt.Fprintf(t, allNamespacesFormatting, p.Pod.Namespace, p.Pod.Name, p.ContainerName, p.Memory.Request, p.Memory.Limit, p.TerminatedTime)
				}

				t.Flush()
				return nil
			}

			if !noHeaders {
				fmt.Fprintf(t, singleNamespaceFormatting, "POD", "CONTAINER", "REQUEST", "LIMIT", "TERMINATION TIME")
			}

			for _, p := range oomPods {
				fmt.Fprintf(t, singleNamespaceFormatting, p.Pod.Name, p.ContainerName, p.Memory.Request, p.Memory.Limit, p.TerminatedTime)
			}

			t.Flush()

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "Don't print headers")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Show OOMKilled containers across all namespaces")
	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Display version information")
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(true)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
