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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2/textlogger"
)

const (
	defaultTimeout = 10 * time.Second
)

func loadConfig(ctx context.Context, path string) (*rest.Config, error) {
	logger, _ := logr.FromContext(ctx)

	if path == "" {
		home := homedir.HomeDir()
		if home == "" {
			logger.Error(errors.New("home directory not found"), "Failed to create client")
			return nil, fmt.Errorf("Failed to create client: home directory not found")
		}
		path = filepath.Join(home, ".kube", "config")
	}

	logger = logger.WithValues("path", path)

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		logger.Error(err, "Failed to load kubeconfig")
		return nil, fmt.Errorf("Failed to load kubeconfig from %s: %w", path, err)
	}

	logger.V(2).Info("Successfully loaded kubeconfig")

	return config, nil
}

func newClient(ctx context.Context, config *rest.Config) (*kubernetes.Clientset, error) {
	logger, _ := logr.FromContext(ctx)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create clientset")
		return nil, fmt.Errorf("Failed to create clientset: %w", err)
	}

	logger.V(2).Info("Successfully created clientset")

	return client, nil
}

func listPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, opts metav1.ListOptions) error {
	logger, _ := logr.FromContext(ctx)
	logger = logger.WithValues("namespace", namespace, "opts", opts)

	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		logger.Error(err, "Failed to list pods")
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
	var cancel context.CancelFunc
	var client *kubernetes.Clientset
	var config *rest.Config
	var ctx context.Context
	var namespace, path string
	var opts metav1.ListOptions

	loggerConfig := textlogger.NewConfig()
	configFlags := flag.NewFlagSet("k8s", flag.ExitOnError)

	rootCmd := &cobra.Command{
		Use:   "k8s",
		Short: "Kubernetes client examples",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			logger := textlogger.NewLogger(loggerConfig).WithName("k8s-client")

			ctx = logr.NewContext(context.Background(), logger)
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)

			config, err = loadConfig(ctx, path)
			if err != nil {
				return err
			}

			client, err = newClient(ctx, config)
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
			return listPods(ctx, client, namespace, opts)
		},
	}

	loggerConfig.AddFlags(configFlags)
	rootCmd.PersistentFlags().AddGoFlagSet(configFlags)
	rootCmd.PersistentFlags().StringVar(&path, "kubeconfig", "", "kubeconfig file path")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to filter resources")

	podsCmd.AddCommand(listCmd)
	rootCmd.AddCommand(podsCmd)

	opts = metav1.ListOptions{}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
