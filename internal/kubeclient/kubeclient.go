package kubeclient

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Interface interface {
	ClientSet() kubernetes.Interface
}

type Client struct {
	clientset kubernetes.Interface
}

func New(kubeconfig string) (Interface, error) {
	if kubeconfig == "" {
		home := homedir.HomeDir()
		if home == "" {
			return nil, fmt.Errorf("failed to create client: home directory not found")
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset from restConfig: %w", err)
	}

	return &Client{
		clientset: clientset,
	}, nil
}

func (c Client) ClientSet() kubernetes.Interface {
	return c.clientset
}
