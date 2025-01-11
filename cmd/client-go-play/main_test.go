package main

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2/ktesting"
)

type testCase struct {
	name     string
	setup    func()
	wantErr  bool
	wantLogs []ktesting.LogEntry
}

type configTestCase struct {
	testCase
	path string
}

func TestLoadConfig(t *testing.T) {
	configTestCases := []configTestCase{
		{
			testCase: testCase{
				name: "empty kubeconfig with no home dir",
				setup: func() {
					t.Setenv("HOME", "")
				},
				wantErr: true,
				wantLogs: []ktesting.LogEntry{
					{
						Type:    ktesting.LogError,
						Message: "Failed to create client",
						Prefix:  "test",
						Err:     errors.New("home directory not found"),
					},
				},
			},
			path: "",
		},
		{
			testCase: testCase{
				name:    "invalid kubeconfig path",
				setup:   func() {},
				wantErr: true,
				wantLogs: []ktesting.LogEntry{
					{
						Type:    ktesting.LogError,
						Message: "Failed to load kubeconfig",
						Prefix:  "test",
						WithKVList: []interface{}{
							"path", "/non/existent/path/config",
						},
						Err: errors.New("stat /non/existent/path/config: no such file or directory"),
					},
				},
			},
			path: "/non/existent/path/config",
		},
	}

	for _, tt := range configTestCases {
		t.Run(tt.name, func(t *testing.T) {
			config := ktesting.NewConfig(ktesting.BufferLogs(true))
			logger := ktesting.NewLogger(t, config).WithName("test")

			ctx := logr.NewContext(context.Background(), logger)

			tt.setup()

			_, err := loadConfig(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			assertLogFields(ctx, t, tt.testCase)
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
