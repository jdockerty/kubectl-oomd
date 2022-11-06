package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jdockerty/kubectl-oomlie/pkg/logger"
	"github.com/jdockerty/kubectl-oomlie/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
    noHeaders bool
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubectl oomlie",
		Short:         "Show pods which have recently been OOMKilled",
		Long:          `Show pods which have recently been terminated by Kubernetes due to an 'Out Of Memory' error`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()

			namespaceFlag := *KubernetesConfigFlags.Namespace
			oomPods, err := plugin.RunPlugin(KubernetesConfigFlags, namespaceFlag, log)
			if err != nil {
				return errors.Unwrap(err)
			}

			t := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			formatting := "%s\t%s\t%s\n"

			if !noHeaders {
				fmt.Fprintf(t, formatting, "POD", "CONTAINER", "TERMINATION TIME")
			}

			for _, v := range oomPods {
				fmt.Fprintf(t, formatting, v.Pod.Name, v.ContainerName, v.TerminatedTime)
			}

			t.Flush()

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

    cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "Don't print headers")
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
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
