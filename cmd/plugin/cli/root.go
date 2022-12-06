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

	// KubernetesConfigFlags provides the generic flags which are available to
	// regular `kubectl` commands, such as `--context` and `--namespace`.
	KubernetesConfigFlags *genericclioptions.ConfigFlags

	// Provides the `--no-headers` flag, this removes them from being printed to stdout.
	noHeaders bool

	// Provides the `--all-namespaces` or `-A` flag which iterates over all namespaces
	// and adds an extra 'NAMESPACE' header to the output.
	allNamespaces bool

	// Provides the `--version` or `-v` flag, displaying build/version information.
	showVersion bool

	// Provides the `--sort-field` flag, allowing sorting by field.
	// Only 'time' is supported currently.
	sortField string


	// Formatting for table output, similar to other kubectl commands.
	t = tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
)

const (
	// Do not use any sorting, this is the default and acts as a value used
	// to catch other arguments that are passed in which are unsupported.
	sortFieldDefault = "none"

	// Sort by termination timestamp in ascending order.
	sortFieldTerminationTime = "time"

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
		Short:         "Show pods/containers which have recently been OOMKilled",
		Long:          `Show pods and containers which have recently been terminated by Kubernetes due to an 'Out Of Memory' error`,
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

			// The namespace provided to the flag takes precedence.
			ns := *KubernetesConfigFlags.Namespace

			namespace, err := plugin.GetNamespace(KubernetesConfigFlags, allNamespaces, ns)
			if err != nil {
				return fmt.Errorf("unable to retrieve namespace, got %s: %w", ns, err)
			}

			oomPods, err := plugin.Run(KubernetesConfigFlags, namespace)
			if err != nil {
				return errors.Unwrap(err)
			}

			// Handle no pods/containers found in a similar fashion to `kubectl`
			if len(oomPods) == 0 {
				if allNamespaces {
					fmt.Println("No out of memory pods found.")
					return nil
				}
				fmt.Printf("No out of memory pods found in %s namespace.\n", namespace)
				return nil
			}

			// Mutate our pods slice in-place depending on the sort-field flag
			// that is used. The default is to do nothing to the slice; coincidentally
			// this does sort by container name, or namespace if `--all-namespaces`
			// flag is used.
			switch sortField {
			case sortFieldTerminationTime:
				oomPods.SortByTimestamp()
			case sortFieldDefault:
			default:
				return fmt.Errorf("%s is not a supported sortable field.", sortField)
			}

			// All namespaces flag requires the extra 'NAMESPACE' heading.
			if allNamespaces {
				if !noHeaders {
					_, err := fmt.Fprintf(t, allNamespacesFormatting, "NAMESPACE", "POD", "CONTAINER", "REQUEST", "LIMIT", "TERMINATION TIME")
					if err != nil {
						return err
					}
				}

				for _, p := range oomPods {
					_, err := fmt.Fprintf(t, allNamespacesFormatting, p.Pod.Namespace, p.Pod.Name, p.ContainerName, p.Memory.Request, p.Memory.Limit, p.TerminatedTime)
					if err != nil {
						return err
					}

				}

				t.Flush()
				return nil
			}

			if !noHeaders {
				_, err := fmt.Fprintf(t, singleNamespaceFormatting, "POD", "CONTAINER", "REQUEST", "LIMIT", "TERMINATION TIME")
				if err != nil {
					return err
				}
			}

			for _, p := range oomPods {
				_, err := fmt.Fprintf(t, singleNamespaceFormatting, p.Pod.Name, p.ContainerName, p.Memory.Request, p.Memory.Limit, p.TerminatedTime)
				if err != nil {
					return err
				}
			}

			t.Flush()

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.Flags().StringVar(&sortField, "sort-field", "none", "Sort by particular field. (Only 'time' is supported currently)")
	cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "Don't print headers")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Show OOMKilled containers across all namespaces")
	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Display version and build information")
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
