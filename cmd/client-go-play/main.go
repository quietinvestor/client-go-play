package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/quietinvestor/client-go-play/internal/kubeclient"
	"github.com/quietinvestor/client-go-play/internal/pods"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2/textlogger"
)

const (
	defaultTimeout = 10 * time.Second
)

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

			config, err = kubeclient.NewConfig(ctx, path)
			if err != nil {
				return err
			}

			client, err = kubeclient.NewClient(ctx, config)
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
			return pods.List(ctx, client, namespace, opts)
		},
	}

	loggerConfig.AddFlags(configFlags)
	rootCmd.PersistentFlags().AddGoFlagSet(configFlags)
	rootCmd.PersistentFlags().StringVarP(&path, "kubeconfig", "k", "", "kubeconfig file path")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to filter resources")

	podsCmd.AddCommand(listCmd)
	rootCmd.AddCommand(podsCmd)

	opts = metav1.ListOptions{}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
