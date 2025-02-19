package cmd

import (
	"context"
	"flag"
	"io"
	"os"
	"time"

	"github.com/quietinvestor/client-go-play/internal/kubeclient"
	"github.com/quietinvestor/client-go-play/internal/pods"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2/textlogger"
)

const defaultTimeout = 10 * time.Second

type contextKey string

const stateKey = contextKey("state")

type State struct {
	Ctx       context.Context
	Cancel    context.CancelFunc
	Client    *kubernetes.Clientset
	Namespace string
	Opts      metav1.ListOptions
	Writer    io.Writer
}

func NewRootCmd() *cobra.Command {
	var namespace, path string

	loggerConfig := textlogger.NewConfig()
	configFlags := flag.NewFlagSet("k8s", flag.ExitOnError)

	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Kubernetes client examples",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			logger := textlogger.NewLogger(loggerConfig).WithName("k8s-client")

			ctx := logr.NewContext(context.Background(), logger)
			ctx, cancel := context.WithTimeout(ctx, defaultTimeout)

			fs := afero.NewOsFs()

			client, err := kubeclient.DefaultClientSet(ctx, fs, path)
			if err != nil {
				cancel()
				return err
			}

			state := &State{
				Ctx:       ctx,
				Cancel:    cancel,
				Client:    client,
				Namespace: namespace,
				Opts:      metav1.ListOptions{},
				Writer:    os.Stdout,
			}

			cmd.SetContext(context.WithValue(ctx, stateKey, state))

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			state := cmd.Context().Value(stateKey).(*State)
			state.Cancel()
		},
	}

	loggerConfig.AddFlags(configFlags)
	cmd.PersistentFlags().AddGoFlagSet(configFlags)
	cmd.PersistentFlags().StringVarP(&path, "kubeconfig", "k", "", "kubeconfig file path")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to filter resources")

	cmd.AddCommand(newPodsCmd())

	return cmd
}

func newPodsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pods",
		Short: "Pod operations",
	}

	cmd.AddCommand(newPodsListCmd())

	return cmd
}

func newPodsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			state := cmd.Context().Value(stateKey).(*State)
			return pods.List(state.Ctx, state.Writer, state.Client, state.Namespace, state.Opts)
		},
	}

	return cmd
}
