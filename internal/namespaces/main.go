package namespaces

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type QuotaResult struct {
	Quotas []v1.ResourceQuota
	Error  error
}

func GetQuotas(ctx context.Context, clientset *kubernetes.Clientset, filterNamespaces []string) ([]QuotaResult, error) {

	var namespaces []string

	if len(filterNamespaces) == 0 {
		// no namespaces provided
		kns, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("no namespaces provided. Errore getting namespaces: %w", err)
		}
		for i := range kns.Items {
			namespaces = append(namespaces, kns.Items[i].Name)
		}
	} else {
		namespaces = append(namespaces, filterNamespaces...)
	}

	slices.Sort(namespaces)

	quotas := make([]QuotaResult, 0)

	for _, n := range namespaces {
		n = strings.TrimSpace(n)

		qs, err := clientset.CoreV1().ResourceQuotas(n).List(ctx, metav1.ListOptions{})
		if err != nil {
			quotas = append(quotas, QuotaResult{Error: err})
			continue
		}

		quotas = append(quotas, QuotaResult{Quotas: qs.Items})
	}

	return quotas, nil
}
