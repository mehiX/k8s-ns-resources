package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mehix/k8s-ns-resources/internal/namespaces"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var cmdResources = cobra.Command{
	Use:   "resources",
	Short: "Show resources for namespace(s)",
	Run: func(cmd *cobra.Command, args []string) {
		Resources()
	},
}

func init() {
	cmdRoot.AddCommand(&cmdResources)
}

func Resources() {

	filterByNS := strings.Split(*ns, ",")

	configLoadRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: *kconf}
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: *kCtx}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadRules, configOverrides).ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	u, err := url.Parse(config.Host)
	if err == nil {
		fmt.Println("Context: ", strings.Split(u.Host, ".")[0])
		fmt.Println()
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	quotas, err := namespaces.GetQuotas(context.TODO(), clientset, filterByNS)
	if err != nil {
		log.Fatalln(err.Error())
	}

	printQuotas(os.Stdout, quotas)
}

func printQuotas(out io.Writer, results []namespaces.QuotaResult) {

	if len(results) == 0 {
		fmt.Println("no data")
		return
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
	defer w.Flush()

	delim := strings.Repeat("---------\t", 6)
	lf := "%s\t%v\t%v\t%v\t%v\t%v\t\n"

	fmt.Fprintln(w, "namespace\tcpuR\tcpuL\tmemoryR\tmemoryL\tstorage\t")
	fmt.Fprintln(w, delim)

	var cpuRT, cpuLT, memRT, memLT, storageT = new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity)

	for _, r := range results {
		if r.Error != nil {
			log.Println(r.Error.Error())
			continue
		}

		for _, q := range r.Quotas {
			h := q.Spec.Hard

			cpuRT.Add(*h.Name(v1.ResourceRequestsCPU, resource.DecimalSI))
			cpuLT.Add(*h.Name(v1.ResourceLimitsCPU, resource.DecimalSI))
			memRT.Add(*h.Name(v1.ResourceRequestsMemory, resource.BinarySI))
			memLT.Add(*h.Name(v1.ResourceLimitsMemory, resource.BinarySI))
			storageT.Add(*h.Name(v1.ResourceRequestsStorage, resource.DecimalSI))

			fmt.Fprintf(w, lf,
				q.Namespace,
				h.Name(v1.ResourceRequestsCPU, resource.DecimalSI),
				h.Name(v1.ResourceLimitsCPU, resource.DecimalSI),
				h.Name(v1.ResourceRequestsMemory, resource.BinarySI),
				h.Name(v1.ResourceLimitsMemory, resource.BinarySI),
				h.Name(v1.ResourceRequestsStorage, resource.DecimalSI),
			)
		}
	}

	fmt.Fprintln(w, delim)
	fmt.Fprintf(w, lf, "total",
		cpuRT,
		cpuLT,
		memRT,
		memLT,
		storageT)

	fmt.Fprintln(w)
}
