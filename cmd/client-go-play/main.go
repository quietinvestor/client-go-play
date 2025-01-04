package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/quietinvestor/client-go-play/internal/client"
	"github.com/quietinvestor/client-go-play/internal/pods"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func handleError(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		klog.ErrorS(err, "Failed to "+msg, keysAndValues...)
		os.Exit(1)
	}
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	home, err := client.HomePathGet()
	handleError(err, "get home directory", "os", runtime.GOOS)

	kubeconfigPath := filepath.Join(home, ".kube", "config")

	kubeconfig, err := client.KubeconfigGet(kubeconfigPath)
	handleError(err, "get kubeconfig", "kubeconfigPath", kubeconfigPath)

	clientset, err := client.ClientsetCreate(kubeconfig)
	handleError(err, "create clientset")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	namespace := ""
	opts := metav1.ListOptions{}

	podsList, err := pods.PodsList(ctx, clientset, namespace, opts)
	handleError(err, "list pods", "namespace", namespace, "opts", opts)

	for _, pod := range podsList {
		fmt.Println(pod.Name)
	}
}
