package kubeclient

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewConfig(ctx context.Context, path string) (*rest.Config, error) {
	logger, _ := logr.FromContext(ctx)

	if path == "" {
		home := homedir.HomeDir()
		if home == "" {
			logger.Error(errors.New("home directory not found"), "Failed to create client")
			return nil, fmt.Errorf("Failed to create client: home directory not found")
		}
		path = filepath.Join(home, ".kube", "config")
	}

	logger = logger.WithValues("path", path)

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		logger.Error(err, "Failed to load kubeconfig")
		return nil, fmt.Errorf("Failed to load kubeconfig from %s: %w", path, err)
	}

	logger.V(2).Info("Successfully loaded kubeconfig")

	return config, nil
}

func NewClient(ctx context.Context, config *rest.Config) (*kubernetes.Clientset, error) {
	logger, _ := logr.FromContext(ctx)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create clientset")
		return nil, fmt.Errorf("Failed to create clientset: %w", err)
	}

	logger.V(2).Info("Successfully created clientset")

	return client, nil
}
