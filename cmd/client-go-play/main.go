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
	defaultPodListTimeout = 30 * time.Second
)

func setupClient(kubeconfig string) (kubeclient.Interface, error, []interface{}) {
	clientset, err := kubeclient.New(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err),
			[]interface{}{"kubeconfig", kubeconfig}
	}

	klog.V(2).InfoS("Successfully created clientset",
		"kubeconfig", kubeconfig)

	return clientset, nil, nil
}

func listPods(ctx context.Context, clientset kubeclient.Interface, namespace string, opts metav1.ListOptions) (error, []interface{}) {
	podsClient := pods.New(clientset.ClientSet(), namespace)
	podList, err := podsClient.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err),
			[]interface{}{
				"namespace", namespace,
				"opts", opts,
				"podsCount", len(podList.Items),
			}
	}

	klog.V(2).InfoS("Successfully listed pods",
		"namespace", namespace,
		"opts", opts,
		"podsCount", len(podList.Items))

	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}

	return nil, nil
}

func run(ctx context.Context, kubeconfig, namespace string, opts metav1.ListOptions) (error, []interface{}) {
	clientset, err, kv := setupClient(kubeconfig)
	if err != nil {
		return err, kv
	}

	return listPods(ctx, clientset, namespace, opts)
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	kubeconfig := ""
	namespace := ""
	opts := metav1.ListOptions{}

	ctx, cancel := context.WithTimeout(context.Background(), defaultPodListTimeout)
	defer cancel()

	if err, kv := run(ctx, kubeconfig, namespace, opts); err != nil {
		klog.ErrorS(err, "Failed to run application", kv...)
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
