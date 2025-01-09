package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

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
			"podCount", len(podList.Items))
		return fmt.Errorf("Failed to list pods: %w", err)
	}

	logger.V(2).Info("Successfully listed pods",
		"podCount", len(podList.Items))

	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}

	return nil
}

func main() {
	var kubeconfig, namespace string
	var clientset *kubernetes.Clientset
	var ctx context.Context
	var cancel context.CancelFunc
	var opts metav1.ListOptions

	loggerConfig := textlogger.NewConfig()
	configFlags := flag.NewFlagSet("k8s", flag.ExitOnError)

	rootCmd := &cobra.Command{
		Use:   "k8s",
		Short: "Kubernetes client examples",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logger := textlogger.NewLogger(loggerConfig).WithName("k8s-client")

			ctx = logr.NewContext(context.Background(), logger)
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)

			var err error
			clientset, err = setupClient(ctx, kubeconfig)
			if err != nil {
				return err
			}

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			cancel()
		},
	}

	podsCmd := &cobra.Command{
		Use:   "pods",
		Short: "Pod operations",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPods(ctx, clientset, namespace, opts)
		},
	}

	loggerConfig.AddFlags(configFlags)
	rootCmd.PersistentFlags().AddGoFlagSet(configFlags)
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig file path")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace to filter resources")

	podsCmd.AddCommand(listCmd)
	rootCmd.AddCommand(podsCmd)

	opts = metav1.ListOptions{}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
