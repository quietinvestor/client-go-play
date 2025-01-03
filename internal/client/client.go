package client

import (
	"errors"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func HomePathGet() (string, error) {
	if homePath := homedir.HomeDir(); homePath != "" {
		return homePath, nil
	} else {
		return "", errors.New("Environment variable not set: $HOME")
	}
}

func KubeconfigGet(kubeconfigPath string) (*rest.Config, error) {
	if _, err := os.Stat(kubeconfigPath); errors.Is(err, os.ErrNotExist) {
		return nil, os.ErrNotExist
	}

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return kubeconfig, err
}

func ClientsetCreate(kubeconfig *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
