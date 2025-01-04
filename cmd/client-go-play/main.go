package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/quietinvestor/client-go-play/internal/kubeclient"
	"github.com/quietinvestor/client-go-play/internal/pods"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const defaultTimeout = 5 * time.Second

func handleError(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		klog.ErrorS(err, "Failed to "+msg, keysAndValues...)
		os.Exit(1)
	}
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	kubeconfig := ""
	clientset, err := kubeclient.New(kubeconfig)
	handleError(err, "create clientset", "kubeconfig", kubeconfig)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	namespace := ""
	opts := metav1.ListOptions{}

	podClient := pods.New(clientset.ClientSet(), namespace)
	podsList, err := podClient.List(ctx, opts)
	handleError(err, "list pods", "namespace", namespace, "opts", opts)

	for _, pod := range podsList.Items {
		fmt.Println(pod.Name)
	}
}
