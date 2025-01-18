package kubeclient

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func NewPath(ctx context.Context, path string) (string, error) {
	logger, _ := logr.FromContext(ctx)

	if path == "" {
		home := homedir.HomeDir()
		if home == "" {
			logger.Error(errors.New("home directory not found"), "failed to create client")
			return "", fmt.Errorf("failed to create client: home directory not found")
		}
		path = filepath.Join(home, ".kube", "config")
	}

	logger = logger.WithValues("path", path)
	logger.V(2).Info("successfully created kubeconfig path")

	return path, nil
}

func NewKubeConfig(ctx context.Context, fs afero.Fs, path string) (*clientcmdapi.Config, error) {
	logger, _ := logr.FromContext(ctx)
	logger = logger.WithValues("path", path)

	exists, err := afero.Exists(fs, path)
	if err != nil {
		logger.Error(err, "failed to check kubeconfig existence")
		return nil, fmt.Errorf("failed to check kubeconfig existence: %w", err)
	}
	if !exists {
		logger.Error(errors.New("kubeconfig file does not exist"), "failed to find kubeconfig file")
		return nil, fmt.Errorf("kubeconfig file does not exist: %s", path)
	}

	kubeConfigBytes, err := afero.ReadFile(fs, path)
	if err != nil {
		logger.Error(err, "failed to read kubeconfig file")
		return nil, fmt.Errorf("failed to read kubeconfig file: %w", err)
	}

	kubeConfig, err := clientcmd.Load(kubeConfigBytes)
	if err != nil {
		logger.Error(err, "failed to parse kubeconfig content")
		return nil, fmt.Errorf("failed to parse kubeconfig content: %w", err)
	}

	logger.V(2).Info("successfully loaded kubeconfig")

	return kubeConfig, nil
}

func NewRestConfig(ctx context.Context, kubeConfig *clientcmdapi.Config) (*rest.Config, error) {
	logger, _ := logr.FromContext(ctx)

	overrides := &clientcmd.ConfigOverrides{}

	clientConfig := clientcmd.NewDefaultClientConfig(*kubeConfig, overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to load REST config")
		return nil, fmt.Errorf("failed to load REST config: %w", err)
	}

	logger.V(2).Info("successfully loaded REST config")

	return restConfig, nil
}

func NewClientSet(ctx context.Context, config *rest.Config) (*kubernetes.Clientset, error) {
	logger, _ := logr.FromContext(ctx)

	if config == nil {
		logger.Error(fmt.Errorf("config is nil"), "failed to create clientset")
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "failed to create clientset")
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	logger.V(2).Info("successfully created clientset")

	return client, nil
}
