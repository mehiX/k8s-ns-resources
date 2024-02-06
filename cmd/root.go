package cmd

import (
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var kCtx *string
var kconf *string
var ns *string

var cmdRoot = cobra.Command{
	Use:   "k8s-ns",
	Short: "Inspect namespace(s)",
}

func init() {

	if home := homedir.HomeDir(); home != "" {
		kconf = cmdRoot.PersistentFlags().String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kconf = cmdRoot.PersistentFlags().String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	kCtx = cmdRoot.PersistentFlags().String("context", "", "choose a Kubernetes context other than the default")

	ns = cmdRoot.PersistentFlags().String("namespaces", "", "comma-separated list of namespaces")
}

func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
	}
}
