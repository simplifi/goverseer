package gcp_secrets_watcher

import (
	"context"
	"sync"
	"testing"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TODO: Replace test project and secret name with valid ones
// TODO: Only replace after discussing with Charlie best project for use case


// TestParseConfig tests the ParseConfig function
func TestParseConfig(t *testing.T) {
	var parsedConfig *Config

	testKey := "valid-secret"
	testProjects := []interface{}{"test-project"}

	parsedConfig, err := ParseConfig(map[string]interface{}{
		"projects":    testProjects,
		"secret_name": testKey,
	})
	assert.NoError(t, err, "Parsing a valid config should not return an error")
	assert.Equal(t, testKey, parsedConfig.SecretName, "SecretName should be set to the value in the config")
	assert.Equal(t, []string{"test-project"}, parsedConfig.Projects, "Projects should be set to the value in the config")
	assert.Equal(t, DefaultCheckIntervalSeconds, parsedConfig.CheckIntervalSeconds, "CheckIntervalSeconds should be set to the default")
	assert.Equal(t, DefaultSecretErrorWaitSeconds, parsedConfig.SecretErrorWaitSeconds, "SecretErrorWaitSeconds should be set to the default")
	assert.Empty(t, parsedConfig.CredentialsFile, "CredentialsFile should be empty by default")

	// Test setting optional values
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"projects":                    testProjects,
		"secret_name":                 testKey,
		"credentials_file":            "/path/to/creds.json",
		"check_interval_seconds":      120,
		"secret_error_wait_seconds":   10,
	})
	assert.NoError(t, err, "Parsing a config with all options should not return an error")
	assert.Equal(t, "/path/to/creds.json", parsedConfig.CredentialsFile, "CredentialsFile should reflect the value in the config")
	assert.Equal(t, 120, parsedConfig.CheckIntervalSeconds, "CheckIntervalSeconds should be set")
	assert.Equal(t, 10, parsedConfig.SecretErrorWaitSeconds, "SecretErrorWaitSeconds should be set")

	// Test error cases for required fields
	_, err = ParseConfig(map[string]interface{}{})
	assert.Error(t, err, "Parsing a config with no projects should return an error")

	_, err = ParseConfig(map[string]interface{}{"projects": testProjects})
	assert.Error(t, err, "Parsing a config with no secret_name should return an error")

	// Test error cases for invalid types
	_, err = ParseConfig(map[string]interface{}{"projects": "not-a-list", "secret_name": testKey})
	assert.Error(t, err, "Parsing a config with invalid projects type should return an error")

	_, err = ParseConfig(map[string]interface{}{"projects": []interface{}{1}, "secret_name": testKey})
	assert.Error(t, err, "Parsing a config with invalid project name type should return an error")

	_, err = ParseConfig(map[string]interface{}{"projects": testProjects, "secret_name": 1})
	assert.Error(t, err, "Parsing a config with invalid secret_name type should return an error")

	_, err = ParseConfig(map[string]interface{}{"projects": testProjects, "secret_name": testKey, "check_interval_seconds": "not-an-int"})
	assert.Error(t, err, "Parsing a config with invalid check_interval_seconds type should return an error")

	_, err = ParseConfig(map[string]interface{}{"projects": testProjects, "secret_name": testKey, "secret_error_wait_seconds": "not-an-int"})
	assert.Error(t, err, "Parsing a config with invalid secret_error_wait_seconds type should return an error")
}

type mockSecretManagerClient struct {
	accessSecretVersionFunc func(ctx context.Context, req *secretmanager.AccessSecretVersionRequest, opts ...option.CallOption) (*secretmanager.AccessSecretVersionResponse, error)
	closeFunc               func() error
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanager.AccessSecretVersionRequest, opts ...option.CallOption) (*secretmanager.AccessSecretVersionResponse, error) {
	if m.accessSecretVersionFunc != nil {
		return m.accessSecretVersionFunc(ctx, req, opts...)
	}

	resp := &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte("new-secret-value"),
		},
		// ETag field will be set via reflection in the test
	}
	resp.ProtoReflect().Set(resp.ProtoReflect().Descriptor().Fields().ByName("Etag"), protoreflect.ValueOfString("mock-etag-2"))

	return resp, nil
}

func (m *mockSecretManagerClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// Tests the New function
func TestNew(t *testing.T) {
	var cfg config.Config
	cfg = config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "gcp_secrets",
			Config: map[string]interface{}{
				"projects":    []interface{}{"test-project"},
				"secret_name": "test-secret",
			},
		},
	}
	watcher, err := New(cfg.Watcher.Config)
	assert.NoError(t, err, "Creating a new GcpSecretsWatcher should not return an error")
	assert.NotNil(t, watcher, "Creating a new GcpSecretsWatcher should return a watcher")

	cfg = config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "gcp_secrets",
			Config: map[string]interface{}{
				"projects": nil,
			},
		},
	}
	watcher, err = New(cfg.Watcher.Config)
	assert.Error(t, err, "Creating a new GcpSecretsWatcher with an invalid config should return an error")
	assert.Nil(t, watcher, "Creating a new GcpSecretsWatcher with an invalid config should not return a watcher")
}

// Tests the Watch function
func TestGcpSecretsWatcher_Watch(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockSecretManagerClient{
		accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...grpc.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return &secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("new-secret-value"),
				},
				Etag: "mock-etag-2",
			}, nil
		},
		closeFunc: func() error {
			return nil
		},
	}

	watcher := GcpSecretsWatcher{
		Config: Config{
			Projects:               []string{"test-project"},
			SecretName:             "test-secret",
			CheckIntervalSeconds:    1,
			SecretErrorWaitSeconds: 1,
		},
		client:        (*secretmanager.Client)(mockClient),
		ctx:           ctx,
		lastKnownETags: map[string]string{"test-project": "mock-etag-1"}, // Initialize lastKnownETags
	}

	changeChan := make(chan interface{}, 1)
	stopChan := make(chan struct{})

	go func() {
		watcher.Watch(changeChan)
		close(stopChan)
	}()

	select {
	case value := <-changeChan:
		expected := map[string]string{"test-project": "new-secret-value"}
		assert.Equal(t, expected, value, "Watch should send the new secret value on the change channel when ETag changes")
	case <-time.After(3 * time.Second):
		t.Fatalf("Watch did not send a change within the timeout")
	}

	watcher.Stop()
	<-stopChan
}


// TestGcpSecretsWatcher_Stop tests the Stop function
func TestGcpSecretsWatcher_Stop(t *testing.T) {
	ctx := context.Background()
	closed := make(chan bool, 1)
	mockClient := &mockSecretManagerClient{
		accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...grpc.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			// Simulate a successful response
			return &secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("test-secret-value"),
				},
				Etag: "mock-etag",
			}, nil
		},
		closeFunc: func() error {
			closed <- true
			return nil
		},
	}

	watcher := GcpSecretsWatcher{
		Config: Config{
			Projects:               []string{"test-project"},
			SecretName:             "test-secret",
			CheckIntervalSeconds:    1,
			SecretErrorWaitSeconds: 1,
		},
		client:         (*secretmanager.Client)(mockClient),
		ctx:            ctx,
		cancel:         func() { /* Mock cancel function */ },
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Watch(make(chan interface{}))
		close(stopChan)
	}()

	// Gives the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Calls stop
	watcher.Stop()
	wg.Wait() // Wait for the Watch goroutine to exit

	// Asserts that the client's Close method was called
	select {
	case <-closed:
		// Success: The Close method was called
	case <-time.After(1 * time.Second):
		t.Fatalf("Stop did not call the client's Close method within the timeout")
	}
}