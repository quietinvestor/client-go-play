package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2/textlogger"
)

const (
	defaultTimeout = 10 * time.Second
)

func setupClient(ctx context.Context, kubeconfig string) (*kubernetes.Clientset, error) {
	logger, _ := logr.FromContext(ctx)

	if kubeconfig == "" {
		home := homedir.HomeDir()
		if home == "" {
			logger.Error(errors.New("home directory not found"), "Failed to create client")
			return nil, fmt.Errorf("Failed to create client: home directory not found")
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	logger = logger.WithValues("kubeconfig", kubeconfig)

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.Error(err, "Failed to load kubeconfig")
		return nil, fmt.Errorf("Failed to load kubeconfig from %s: %w", kubeconfig, err)
	}

	logger.V(2).Info("Successfully loaded kubeconfig")

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset")
		return nil, fmt.Errorf("Failed to create clientset: %w", err)
	}

	logger.V(2).Info("Successfully created clientset")

	return clientset, nil
}

func listPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, opts metav1.ListOptions) error {
	logger, _ := logr.FromContext(ctx)
	logger = logger.WithValues("namespace", namespace, "opts", opts)

	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		logger.Error(err, "Failed to list pods",
			"podsCount", len(podList.Items))
		return fmt.Errorf("Failed to list pods: %w", err)
	}

	logger.V(2).Info("Successfully listed pods",
		"podsCount", len(podList.Items))

	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}

	return nil
}

func main() {
	config := textlogger.NewConfig()
	logger := textlogger.NewLogger(config).WithName("pod-client")

	ctx := logr.NewContext(context.Background(), logger)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	kubeconfig := ""
	namespace := ""
	opts := metav1.ListOptions{}

	clientset, err := setupClient(ctx, kubeconfig)
	if err != nil {
		os.Exit(1)
	}

	if err := listPods(ctx, clientset, namespace, opts); err != nil {
		os.Exit(1)
	}
}
