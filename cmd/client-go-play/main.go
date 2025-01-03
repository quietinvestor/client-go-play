package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/quietinvestor/client-go-play/internal/client"
	"github.com/quietinvestor/client-go-play/internal/pods"
)

func errCheck(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	home, err := client.HomePathGet()
	errCheck(err)

	kubeconfig, err := client.KubeconfigGet(filepath.Join(home, ".kube", "config"))
	errCheck(err)

	clientset, err := client.ClientsetCreate(kubeconfig)
	errCheck(err)

	podsList, err := pods.PodsList(clientset)
	errCheck(err)

	for _, pod := range podsList {
		fmt.Println(pod.Name)
	}
}
