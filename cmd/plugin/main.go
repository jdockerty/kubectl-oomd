package main

import (
	"github.com/jdockerty/kubectl-oomlie/cmd/plugin/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // auth plugins
)

func main() {
	cli.InitAndExecute()
}
