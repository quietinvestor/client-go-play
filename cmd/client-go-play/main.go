package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/quietinvestor/client-go-play/internal/kubeclient"
	"github.com/quietinvestor/client-go-play/internal/pods"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	defaultTimeout = 10 * time.Second
)

func setupClient(ctx context.Context, kubeconfig string) (kubeclient.Interface, error) {
	logger := klog.FromContext(ctx)

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
	logger := klog.FromContext(ctx)

	podsClient := pods.New(clientset.ClientSet(), namespace)
	podList, err := podsClient.List(ctx, opts)
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
	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	logger := klog.Background().WithName("pod-client")

	ctx := klog.NewContext(context.Background(), logger)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	kubeconfig := ""
	namespace := ""
	opts := metav1.ListOptions{}

	if err := run(ctx, kubeconfig, namespace, opts); err != nil {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
