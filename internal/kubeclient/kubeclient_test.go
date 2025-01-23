package kubeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2/ktesting"
)

type testCase struct {
	fileContent string
	fs          afero.Fs
	kubeConfig  *clientcmdapi.Config
	name        string
	path        string
	restConfig  *rest.Config
	setup       func()
	wantErr     bool
	wantLogs    []ktesting.LogEntry
}

func (tc *testCase) writeMockFile() error {
	dir := filepath.Dir(tc.path)

	if err := tc.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories for path %s: %w", tc.path, err)
	}

	if err := afero.WriteFile(tc.fs, tc.path, []byte(tc.fileContent), 0644); err != nil {
		return fmt.Errorf("failed to write mock file: %w", err)
	}

	return nil
}

func TestNewPath(t *testing.T) {
	testCases := []testCase{
		{
			name: "kubeconfig with home dir",
			path: "/home/testuser/.kube/config",
			setup: func() {
				t.Setenv("HOME", "/home/testuser")
			},
			wantErr: false,
			wantLogs: []ktesting.LogEntry{
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully created kubeconfig path",
					Prefix:    "test",
					Verbosity: 2,
				},
			},
		},
		{
			name: "empty kubeconfig with no home dir",
			path: "",
			setup: func() {
				t.Setenv("HOME", "")
			},
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to create client",
					Prefix:  "test",
					Err:     errors.New("home directory not found"),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig := ktesting.NewConfig(ktesting.BufferLogs(true))
			logger := ktesting.NewLogger(t, loggerConfig).WithName("test")

			ctx := logr.NewContext(context.Background(), logger)

			if tt.setup != nil {
				tt.setup()
			}

			_, err := NewPath(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertLogFields(ctx, t, tt)
		})
	}
}

type mockFs struct {
	afero.MemMapFs
	readFileErr error
	statErr     error
}

func (m *mockFs) Open(name string) (afero.File, error) {
	if m.readFileErr != nil {
		return nil, m.readFileErr
	}
	return m.MemMapFs.Open(name)
}

func (m *mockFs) Stat(name string) (os.FileInfo, error) {
	if m.statErr != nil {
		return nil, m.statErr
	}
	return m.MemMapFs.Stat(name)
}

func TestNewKubeConfig(t *testing.T) {
	testCases := []testCase{
		{
			fileContent: "proper: yaml",
			fs:          &mockFs{},
			name:        "kubeconfig with home directory",
			path:        "/home/testuser/.kube/config",
			wantErr:     false,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogInfo,
					Message: "successfully loaded kubeconfig",
					Prefix:  "test",
					WithKVList: []interface{}{
						"path", "/home/testuser/.kube/config",
					},
					Verbosity: 2,
				},
			},
		},
		{
			fs: &mockFs{
				statErr: fmt.Errorf("simulated stat error"),
			},
			name:    "error checking kubeconfig existence",
			path:    "/home/testuser/.kube/config",
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to check kubeconfig existence",
					Prefix:  "test",
					Err:     fmt.Errorf("simulated stat error"),
					WithKVList: []interface{}{
						"path", "/home/testuser/.kube/config",
					},
				},
			},
		},
		{
			fs:      &mockFs{},
			name:    "kubeconfig does not exist",
			path:    "/non/existent/path/config",
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to find kubeconfig file",
					Prefix:  "test",
					WithKVList: []interface{}{
						"path", "/non/existent/path/config",
					},
					Err: errors.New("kubeconfig file does not exist"),
				},
			},
		},
		{
			fileContent: "kubeconfig definition",
			fs: &mockFs{
				readFileErr: fmt.Errorf("simulated read error"),
			},
			name:    "Error reading kubeconfig file",
			path:    "/home/testuser/.kube/config",
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to read kubeconfig file",
					Prefix:  "test",
					Err:     fmt.Errorf("simulated read error"),
					WithKVList: []interface{}{
						"path", "/home/testuser/.kube/config",
					},
				},
			},
		},
		{
			fileContent: `{"kind": 123}`,
			fs:          &mockFs{},
			name:        "Error parsing kubeconfig content",
			path:        "/home/testuser/.kube/config",
			wantErr:     true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to parse kubeconfig content",
					Prefix:  "test",
					Err: func() error {
						var x struct {
							Kind string `json:"kind,omitempty"`
						}
						err := json.Unmarshal([]byte(`{"kind":123}`), &x)
						if err != nil {
							return fmt.Errorf(
								"couldn't get version/kind; "+
									"json parse error: %v", err)
						}
						return nil
					}(),
					WithKVList: []interface{}{
						"path", "/home/testuser/.kube/config",
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig := ktesting.NewConfig(ktesting.BufferLogs(true))
			logger := ktesting.NewLogger(t, loggerConfig).WithName("test")

			ctx := logr.NewContext(context.Background(), logger)

			if tt.fs != nil && tt.fileContent != "" {
				if err := tt.writeMockFile(); err != nil {
					t.Error(err)
				}
			}

			_, err := NewKubeConfig(ctx, tt.fs, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertLogFields(ctx, t, tt)
		})
	}
}

func TestNewRestConfig(t *testing.T) {
	testCases := []testCase{
		{
			kubeConfig: &clientcmdapi.Config{
				Clusters: map[string]*clientcmdapi.Cluster{
					"test-cluster": {
						Server: "https://example.com",
					},
				},
				Contexts: map[string]*clientcmdapi.Context{
					"test-context": {
						Cluster: "test-cluster",
					},
				},
				CurrentContext: "test-context",
			},
			name:    "valid config",
			wantErr: false,
			wantLogs: []ktesting.LogEntry{
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully loaded REST config",
					Prefix:    "test",
					Verbosity: 2,
				},
			},
		},
		{
			kubeConfig: &clientcmdapi.Config{},
			name:       "invalid config",
			wantErr:    true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to load REST config",
					Prefix:  "test",
					Err: fmt.Errorf(
						"invalid configuration: " +
							"no configuration has been provided, " +
							"try setting KUBERNETES_MASTER environment variable"),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig := ktesting.NewConfig(ktesting.BufferLogs(true))
			logger := ktesting.NewLogger(t, loggerConfig).WithName("test")

			ctx := logr.NewContext(context.Background(), logger)

			_, err := NewRestConfig(ctx, tt.kubeConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRestConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertLogFields(ctx, t, tt)
		})
	}
}

func TestNewClientSet(t *testing.T) {
	testCases := []testCase{
		{
			name:    "valid configuration",
			wantErr: false,
			wantLogs: []ktesting.LogEntry{
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully created clientset",
					Prefix:    "test",
					Verbosity: 2,
				},
			},
			restConfig: &rest.Config{
				Host: "https://127.0.0.1:6443",
			},
		},
		{
			name:    "invalid host",
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to create clientset",
					Prefix:  "test",
					Err:     errors.New("host must be a URL or a host:port pair: \"http//:invalid-host\""),
				},
			},
			restConfig: &rest.Config{
				Host: "http//:invalid-host",
			},
		},
		{
			name:    "invalid TLS configuration",
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to create clientset",
					Prefix:  "test",
					Err:     errors.New("open /invalid/ca.crt: no such file or directory"),
				},
			},
			restConfig: &rest.Config{
				Host: "https://127.0.0.1:6443",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/invalid/ca.crt",
				},
			},
		},
		{
			name:    "nil configuration",
			wantErr: true,
			wantLogs: []ktesting.LogEntry{
				{
					Type:    ktesting.LogError,
					Message: "failed to create clientset",
					Prefix:  "test",
					Err:     errors.New("config is nil"),
				},
			},
			restConfig: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig := ktesting.NewConfig(ktesting.BufferLogs(true))
			logger := ktesting.NewLogger(t, loggerConfig).WithName("test")

			ctx := logr.NewContext(context.Background(), logger)

			if tt.setup != nil {
				tt.setup()
			}

			_, err := NewClientSet(ctx, tt.restConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClientSet() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertLogFields(ctx, t, tt)
		})
	}
}

func TestNewClient(t *testing.T) {
	testCases := []testCase{
		{
			fileContent: `apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: https://example.com
contexts:
- name: context
  context:
    cluster: cluster
    user: user
current-context: context`,
			fs:      &mockFs{},
			name:    "valid configuration",
			path:    "/home/testuser/.kube/config",
			wantErr: false,
			wantLogs: []ktesting.LogEntry{
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully created kubeconfig path",
					Prefix:    "test",
					Verbosity: 2,
				},
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully loaded kubeconfig",
					Prefix:    "test",
					Verbosity: 2,
				},
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully loaded REST config",
					Prefix:    "test",
					Verbosity: 2,
				},
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully created clientset",
					Prefix:    "test",
					Verbosity: 2,
				},
				{
					Type:      ktesting.LogInfo,
					Message:   "successfully created client",
					Prefix:    "test",
					Verbosity: 2,
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig := ktesting.NewConfig(ktesting.BufferLogs(true))
			logger := ktesting.NewLogger(t, loggerConfig).WithName("test")

			ctx := logr.NewContext(context.Background(), logger)

			if tt.setup != nil {
				tt.setup()
			}

			if tt.fs != nil && tt.fileContent != "" {
				if err := tt.writeMockFile(); err != nil {
					t.Error(err)
				}
			}

			_, err := NewClient(ctx, tt.fs, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertLogFields(ctx, t, tt)
		})
	}
}

func assertLogFields(ctx context.Context, t *testing.T, tt testCase) {
	logger, _ := logr.FromContext(ctx)

	testLogger, ok := logger.GetSink().(ktesting.Underlier)
	if !ok {
		t.Fatal("Expected a ktesting logger")
	}

	logs := testLogger.GetBuffer().Data()
	if len(logs) != len(tt.wantLogs) {
		t.Errorf("got %d log entries, want %d", len(logs), len(tt.wantLogs))
		return
	}

	for i, want := range tt.wantLogs {
		got := logs[i]
		if got.Type != want.Type {
			t.Errorf("log[%d].Type = %v, want %v", i, got.Type, want.Type)
		}
		if got.Prefix != want.Prefix {
			t.Errorf("log[%d].Prefix = %q, want %q", i, got.Prefix, want.Prefix)
		}
		if got.Message != want.Message {
			t.Errorf("log[%d].Message = %q, want %q", i, got.Message, want.Message)
		}
		if want.Err != nil && got.Err.Error() != want.Err.Error() {
			t.Errorf("log[%d].Err = %v, want %v", i, got.Err, want.Err)
		}
		if want.Verbosity != 0 && got.Verbosity != want.Verbosity {
			t.Errorf("log[%d].Verbosity = %d, want %d", i, got.Verbosity, want.Verbosity)
		}
		if len(want.WithKVList) > 0 {
			if len(got.WithKVList) != len(want.WithKVList) {
				t.Errorf("log[%d].WithKVList length = %d, want %d", i, len(got.WithKVList), len(want.WithKVList))
				return
			}
			for j := range want.WithKVList {
				if got.WithKVList[j] != want.WithKVList[j] {
					t.Errorf("log[%d].WithKVList[%d] = %v, want %v", i, j, got.WithKVList[j], want.WithKVList[j])
				}
			}
		}
	}
}
