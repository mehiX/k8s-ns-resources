package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var kconf *string
	if home := homedir.HomeDir(); home != "" {
		kconf = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kconf = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	var kCtx = flag.String("context", "", "choose a Kubernetes context other than the default")

	ns := flag.String("namespaces", "", "comma-separated list of namespaces")

	flag.Parse()

	if *ns == "" {
		fmt.Println("use --namespaces to provide the namespaces")
		os.Exit(1)
	}

	namespaces := strings.Split(*ns, ",")

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

	ctx := context.TODO()

	quotas := make([]v1.ResourceQuota, 0)

	for _, n := range namespaces {
		n = strings.TrimSpace(n)

		qs, err := clientset.CoreV1().ResourceQuotas(n).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("Getting limitRanges: %s\n", err.Error())
			continue
		}

		quotas = append(quotas, qs.Items...)

	}

	printQuotas(os.Stdout, quotas)
}

func printQuotas(out io.Writer, quotas []v1.ResourceQuota) {

	if len(quotas) == 0 {
		fmt.Println("no data")
		return
	}

	w := tabwriter.NewWriter(out, 10, 0, 2, ' ', tabwriter.AlignRight|tabwriter.Debug)
	defer w.Flush()

	delim := strings.Repeat("-----------\t", 7)
	lf := "%s\t%v\t%v\t%v\t%v\t%v\t%v\t\n"

	fmt.Fprintln(w, "namespace\tcpuR\tcpuL\tmemoryR\tmemoryL\tstorage\tstorage (ephem)\t")
	fmt.Fprintln(w, delim)

	var cpuRT, cpuLT, memRT, memLT, storageT, ephemStorageT = new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity),
		new(resource.Quantity)

	for _, q := range quotas {
		h := q.Spec.Hard

		cpuRT.Add(*h.Name(v1.ResourceRequestsCPU, resource.DecimalSI))
		cpuLT.Add(*h.Name(v1.ResourceLimitsCPU, resource.DecimalSI))
		memRT.Add(*h.Name(v1.ResourceRequestsMemory, resource.BinarySI))
		memLT.Add(*h.Name(v1.ResourceLimitsMemory, resource.BinarySI))
		storageT.Add(*h.Name(v1.ResourceStorage, resource.DecimalSI))
		ephemStorageT.Add(*h.Name(v1.ResourceEphemeralStorage, resource.DecimalSI))

		fmt.Fprintf(w, lf,
			q.Namespace,
			h.Name(v1.ResourceRequestsCPU, resource.DecimalSI),
			h.Name(v1.ResourceLimitsCPU, resource.DecimalSI),
			h.Name(v1.ResourceRequestsMemory, resource.BinarySI),
			h.Name(v1.ResourceLimitsMemory, resource.BinarySI),
			h.Name(v1.ResourceStorage, resource.DecimalSI),
			h.Name(v1.ResourceEphemeralStorage, resource.DecimalSI),
		)
	}

	fmt.Fprintln(w, delim)
	fmt.Fprintf(w, lf, "total",
		cpuRT,
		cpuLT,
		memRT,
		memLT,
		storageT,
		ephemStorageT)

	fmt.Fprintln(w)
}
