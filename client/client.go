package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ClientsetCreate(kubeconfig *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(kubeconfig)

	if err != nil {
		return nil, err
	}

	return clientset, nil
}
