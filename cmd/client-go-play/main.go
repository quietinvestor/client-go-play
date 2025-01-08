package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/quietinvestor/client-go-play/internal/kubeclient"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2/textlogger"
)

const (
	defaultTimeout = 10 * time.Second
)

func setupClient(ctx context.Context, kubeconfig string) (kubeclient.Interface, error) {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		config := textlogger.NewConfig()
		logger = textlogger.NewLogger(config).WithName("setup-client")
		logger.V(2).Info("No logger found in context, created new logger")
	}

	clientset, err := kubeclient.New(kubeconfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset",
			"kubeconfig", kubeconfig)
		return nil, fmt.Errorf("Failed to create clientset: %w", err)
	}

	logger.V(2).Info("Successfully created clientset",
		"kubeconfig", kubeconfig)

	return clientset, nil
}

func listPods(ctx context.Context, clientset kubeclient.Interface, namespace string, opts metav1.ListOptions) error {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		config := textlogger.NewConfig()
		logger = textlogger.NewLogger(config).WithName("list-pods")
		logger.V(2).Info("No logger found in context, created new logger")
	}

	podList, err := clientset.ClientSet().CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		logger.Error(err, "Failed to list pods",
			"namespace", namespace,
			"opts", opts,
			"podsCount", len(podList.Items))
		return fmt.Errorf("Failed to list pods: %w", err)
	}

	logger.V(2).Info("Successfully listed pods",
		"namespace", namespace,
		"opts", opts,
		"podsCount", len(podList.Items))

	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}

	return nil
}

func run(ctx context.Context, kubeconfig, namespace string, opts metav1.ListOptions) error {
	clientset, err := setupClient(ctx, kubeconfig)
	if err != nil {
		return err
	}

	return listPods(ctx, clientset, namespace, opts)
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

	if err := run(ctx, kubeconfig, namespace, opts); err != nil {
		os.Exit(1)
	}
}
