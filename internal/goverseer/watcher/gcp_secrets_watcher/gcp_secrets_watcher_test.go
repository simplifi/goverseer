package gcp_secrets_watcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Tests the ParseConfig function
func TestParseConfig(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    map[string]interface{}
		expectedConfig *Config
		expectedError  string
	}{
		{
			name: "Valid - Full config",
			inputConfig: map[string]interface{}{
				"project_id":                "test-project-full",
				"secret_name":               "test-secret-full",
				"credentials_file":          "/path/to/full-creds.json",
				"check_interval_seconds":    120,
				"secret_error_wait_seconds": 10,
				"secrets_file_path":         "/tmp/full-secrets.txt",
			},
			expectedConfig: &Config{
				ProjectID:              "test-project-full",
				SecretName:             "test-secret-full",
				CredentialsFile:        "/path/to/full-creds.json",
				CheckIntervalSeconds:   120,
				SecretErrorWaitSeconds: 10,
				SecretsFilePath:        "/tmp/full-secrets.txt",
			},
			expectedError: "",
		},
		{
			name: "Valid - Minimal config",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project-req",
				"secret_name":       "test-secret-req",
				"secrets_file_path": "/tmp/test-secrets-req.txt",
			},
			expectedConfig: &Config{
				ProjectID:              "test-project-req",
				SecretName:             "test-secret-req",
				CheckIntervalSeconds:   DefaultCheckIntervalSeconds,
				SecretErrorWaitSeconds: DefaultSecretErrorWaitSeconds,
				SecretsFilePath:        "/tmp/test-secrets-req.txt",
			},
			expectedError: "",
		},
		{
			name: "Invalid - Missing project_id",
			inputConfig: map[string]interface{}{
				"secret_name":       "test-secret",
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "project_id is required",
		},
		{
			name: "Invalid - Empty project_id",
			inputConfig: map[string]interface{}{
				"project_id":        "",
				"secret_name":       "test-secret",
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "project_id must not be empty",
		},
		{
			name: "Invalid - project_id wrong type",
			inputConfig: map[string]interface{}{
				"project_id":        123,
				"secret_name":       "test-secret",
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "project_id must be a string",
		},
		{
			name: "Invalid - Missing secret_name",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "secret_name is required",
		},
		{
			name: "Invalid - Empty secret_name",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secret_name":       "",
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "secret_name must not be empty",
		},
		{
			name: "Invalid - secret_name wrong type",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secret_name":       123,
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "secret_name must be a string",
		},
		{
			name: "Invalid - Missing secrets_file_path",
			inputConfig: map[string]interface{}{
				"project_id":  "test-project",
				"secret_name": "test-secret",
			},
			expectedConfig: nil,
			expectedError:  "secrets_file_path is required",
		},
		{
			name: "Invalid - Empty secrets_file_path",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secret_name":       "test-secret",
				"secrets_file_path": "",
			},
			expectedConfig: nil,
			expectedError:  "secrets_file_path must not be empty",
		},
		{
			name: "Invalid - secrets_file_path wrong type",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secret_name":       "test-secret",
				"secrets_file_path": 123,
			},
			expectedConfig: nil,
			expectedError:  "secrets_file_path must be a string",
		},
		{
			name: "Invalid - check_interval_seconds non-positive",
			inputConfig: map[string]interface{}{
				"project_id":             "test-project",
				"secret_name":            "test-secret",
				"check_interval_seconds": 0,
				"secrets_file_path":      "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "check_interval_seconds must be a positive integer",
		},
		{
			name: "Invalid - check_interval_seconds wrong type",
			inputConfig: map[string]interface{}{
				"project_id":             "test-project",
				"secret_name":            "test-secret",
				"check_interval_seconds": "abc",
				"secrets_file_path":      "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "check_interval_seconds must be an integer",
		},
		{
			name: "Invalid - secret_error_wait_seconds non-positive",
			inputConfig: map[string]interface{}{
				"project_id":                "test-project",
				"secret_name":               "test-secret",
				"secret_error_wait_seconds": 0,
				"secrets_file_path":         "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "secret_error_wait_seconds must be a positive integer",
		},
		{
			name: "Invalid - secret_error_wait_seconds wrong type",
			inputConfig: map[string]interface{}{
				"project_id":                "test-project",
				"secret_name":               "test-secret",
				"secret_error_wait_seconds": "abc",
				"secrets_file_path":         "/tmp/test-secrets.txt",
			},
			expectedConfig: nil,
			expectedError:  "secret_error_wait_seconds must be an integer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := ParseConfig(tt.inputConfig)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedConfig, gotConfig)
			}
		})
	}
}

// A mock implementation of SecretManagerClientFactory
type mockSecretManagerClientFactory struct {
	mock.Mock
}

func (m *mockSecretManagerClientFactory) CreateClient(ctx context.Context, credentialFile string) (SecretManagerClientInterface, error) {
	args := m.Called(ctx, credentialFile)
	if client, ok := args.Get(0).(SecretManagerClientInterface); ok {
		return client, args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock Secret Manager client for testing
type mockSecretManagerClient struct {
	mock.Mock
}

func (m *mockSecretManagerClient) GetSecretVersion(ctx context.Context, req *secretmanagerpb.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	args := m.Called(ctx, req, opts)
	if a := args.Get(0); a != nil {
		return a.(*secretmanagerpb.SecretVersion), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	args := m.Called(ctx, req, opts)
	if a := args.Get(0); a != nil {
		return a.(*secretmanagerpb.AccessSecretVersionResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSecretManagerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Tests the New function with valid and invalid configs
func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		inputConfig   map[string]interface{}
		mockFactory   *mockSecretManagerClientFactory
		mockClient    *mockSecretManagerClient
		expectedError string
	}{
		{
			name: "Valid config",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secret_name":       "test-secret",
				"secrets_file_path": "/tmp/test-secrets.txt",
				"credentials_file":  "/path/to/mock-creds.json",
			},
			mockFactory:   new(mockSecretManagerClientFactory),
			mockClient:    new(mockSecretManagerClient),
			expectedError: "",
		},
		{
			name: "Valid config - ADC (no creds file)",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project-adc",
				"secret_name":       "test-secret-adc",
				"secrets_file_path": "/tmp/test-secrets-adc.txt",
			},
			mockFactory:   new(mockSecretManagerClientFactory),
			mockClient:    new(mockSecretManagerClient),
			expectedError: "",
		},
		{
			name: "Invalid config - Missing project_id",
			inputConfig: map[string]interface{}{
				"secret_name":       "test-secret",
				"secrets_file_path": "/tmp/test-secrets.txt",
			},
			mockFactory:   nil,
			mockClient:    nil,
			expectedError: "project_id is required",
		},
		{
			name: "Invalid config - Client creation error",
			inputConfig: map[string]interface{}{
				"project_id":        "test-project",
				"secret_name":       "test-secret",
				"secrets_file_path": "/tmp/test-secrets.txt",
				"credentials_file":  "/path/to/bad-creds.json",
			},
			mockFactory:   new(mockSecretManagerClientFactory),
			mockClient:    nil,
			expectedError: "failed to create Secrets Manager client: mock client creation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up factory expectations (only if mockFactory is provided)
			if tt.mockFactory != nil {
				tt.mockFactory.On("CreateClient", mock.Anything, mock.AnythingOfType("string")).Return(tt.mockClient, fmt.Errorf("mock client creation error")).Maybe()
				if tt.expectedError == "" {
					tt.mockFactory.ExpectedCalls = []*mock.Call{}
					tt.mockFactory.On("CreateClient", mock.Anything, mock.AnythingOfType("string")).Return(tt.mockClient, nil).Once()
				}
			}

			// Pass the mock factory to New
			var watcher *GcpSecretsWatcher
			var err error
			if tt.mockFactory != nil {
				watcher, err = New(tt.inputConfig, tt.mockFactory)
			} else {
				watcher, err = New(tt.inputConfig)
			}

			if tt.expectedError != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tt.expectedError, "Error message should contain expected text")
				assert.Nil(t, watcher, "Watcher should be nil on error")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
				assert.NotNil(t, watcher, "Watcher should not be nil")
				assert.NotNil(t, watcher.client, "Watcher client should not be nil")
				assert.Equal(t, tt.mockClient, watcher.client, "Watcher should use the injected mock client")
			}

			if tt.mockFactory != nil {
				tt.mockFactory.AssertExpectations(t)
			}
		})
	}
}

// Tests the Watch function
// Simulates a change in the secret value,
// checks if the watcher detects it,
// and sends the new value to the change channel
func TestGcpSecretsWatcher_Watch_EtagChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockClient := new(mockSecretManagerClient)

	log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Started")

	// Establishes expectation for the initial GetSecretVersion
	mockClient.On("GetSecretVersion", mock.Anything, mock.Anything, mock.Anything).Return(
		&secretmanagerpb.SecretVersion{Etag: "etag-1"}, nil).Once()

	// Establishes expectation for AccessSecretVersion
	mockClient.On("AccessSecretVersion", mock.Anything, mock.Anything, mock.Anything).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte("new-secret-value"),
			},
		}, nil).Once()

	watcher := GcpSecretsWatcher{
		Config: Config{
			ProjectID:              "test-project",
			SecretName:             "test-secret",
			CheckIntervalSeconds:   1,
			SecretErrorWaitSeconds: 1,
			SecretsFilePath:        "/tmp/test-secrets.txt",
		},
		client:        mockClient,
		ctx:           ctx,
		lastKnownETag: "",
		cancel:        cancel,
	}

	changeChan := make(chan interface{}, 1)
	stopTestGoroutine := make(chan struct{})

	go func() {
		defer close(stopTestGoroutine)
		log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Watch goroutine started")
		watcher.Watch(changeChan)
		log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Watch goroutine finished")
	}()

	select {
	case value := <-changeChan:
		expected := "new-secret-value"
		assert.Equal(t, expected, value, "Watch should send the new secret value on the change channel when ETag changes")
		log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Change received:", value)
		watcher.Stop()
	case <-time.After(time.Duration(watcher.Config.CheckIntervalSeconds) * 2 * time.Second):
		t.Fatalf("Watch did not send a change within the timeout")
	}

	<-stopTestGoroutine
	log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Test finished")
	mockClient.AssertExpectations(t)
}

// Tests the Stop function
func TestGcpSecretsWatcher_Stop(t *testing.T) {
	log.Println("TestGcpSecretsWatcher_Stop: Started")
	ctx, cancel := context.WithCancel(context.Background())
	closed := make(chan bool, 1)
	mockClient := new(mockSecretManagerClient)
	mockClient.On("GetSecretVersion", mock.Anything, mock.Anything, mock.Anything).Return(
		&secretmanagerpb.SecretVersion{Etag: "etag-1"}, nil).Maybe()
	mockClient.On("AccessSecretVersion", mock.Anything, mock.Anything, mock.Anything).Return(
		&secretmanagerpb.AccessSecretVersionResponse{
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte("test-secret-value"),
			},
		}, nil).Maybe()
	mockClient.On("Close").Return(nil).Once().Run(func(args mock.Arguments) {
		closed <- true
	})

	watcher := GcpSecretsWatcher{
		Config: Config{
			ProjectID:              "test-project",
			SecretName:             "test-secret",
			CheckIntervalSeconds:   1,
			SecretErrorWaitSeconds: 1,
			SecretsFilePath:        "/tmp/test-secrets.txt",
		},
		client: mockClient,
		ctx:    ctx,
		cancel: cancel,
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	// Starts goroutine to simulate the Watch function
	go func() {
		defer wg.Done()
		log.Println("TestGcpSecretsWatcher_Stop: Watch goroutine started")
		watcher.Watch(make(chan interface{}, 1))
		log.Println("TestGcpSecretsWatcher_Stop: Watch goroutine finished")
		close(stopChan)
	}()

	// Gives the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Calls Stop, waits for the goroutine to finish
	log.Println("TestGcpSecretsWatcher_Stop: Calling watcher.Stop()")
	watcher.Stop()
	log.Println("TestGcpSecretsWatcher_Stop: Waiting for Watch goroutine")
	wg.Wait()

	// Explicitly calls Close on the mock client
	watcher.client.Close()

	// Asserts that the client's Close method was called
	log.Println("TestGcpSecretsWatcher_Stop: Asserting mock expectations")
	mockClient.AssertExpectations(t)
	log.Println("TestGcpSecretsWatcher_Stop: Finished")

	select {
	case <-closed:
		log.Println("TestGcpSecretsWatcher_Stop: Client closed successfully")
	case <-stopChan:
		log.Println("TestGcpSecretsWatcher_Stop: Stop channel closed")
	case <-time.After(1 * time.Second):
		t.Fatalf("Stop did not trigger client Close")
	}
}
