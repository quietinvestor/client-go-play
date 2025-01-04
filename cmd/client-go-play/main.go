package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/quietinvestor/client-go-play/internal/client"
	"github.com/quietinvestor/client-go-play/internal/pods"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func errCheck(msg string, err error) {
	if err != nil {
		log.Fatalf("Failed to %s: %v", msg, err)
	}
}

func main() {
	home, err := client.HomePathGet()
	errCheck("get home directory", err)

	kubeconfig, err := client.KubeconfigGet(filepath.Join(home, ".kube", "config"))
	errCheck("get kubeconfig", err)

	clientset, err := client.ClientsetCreate(kubeconfig)
	errCheck("create clientset", err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	opts := metav1.ListOptions{}

	podsList, err := pods.PodsList(ctx, clientset, "", opts)
	errCheck("list pods", err)

	for _, pod := range podsList {
		fmt.Println(pod.Name)
	}
}
