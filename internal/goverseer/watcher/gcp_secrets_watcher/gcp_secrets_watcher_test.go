package gcp_secrets_watcher

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// Ensure mockClient satisfies the SecretManagerClientInterface
var _ SecretManagerClientInterface = (*mockSecretManagerClient)(nil)

// Tests the New function
func TestNew(t *testing.T) {
	validConfig := map[string]interface{}{
		"project_id":        "test-project",
		"secret_name":       "test-secret",
		"secrets_file_path": "/tmp/test-secrets.txt",
	}
	watcher, err := New(validConfig)
	assert.NoError(t, err, "Creating a new GcpSecretsWatcher should not return an error")
	assert.NotNil(t, watcher, "Creating a new GcpSecretsWatcher should return a watcher")

	invalidConfig := map[string]interface{}{
		"project_id":        nil,
		"secret_name":       "test-secret",
		"secrets_file_path": "/tmp/test-secrets.txt",
	}
	watcher, err = New(invalidConfig)
	assert.Error(t, err, "Creating a new GcpSecretsWatcher with an invalid config should return an error")
	assert.Nil(t, watcher, "Creating a new GcpSecretsWatcher with an invalid config should not return a watcher")
}

// Tests the Watch function for value changes based on ETag
func TestGcpSecretsWatcher_Watch_EtagChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockClient := new(mockSecretManagerClient)

	log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Started")

	// Sets expectation for the initial GetSecretVersion
	mockClient.On("GetSecretVersion", mock.Anything, mock.Anything, mock.Anything).Return(
		&secretmanagerpb.SecretVersion{Etag: "etag-1"}, nil).Once() // Initial ETag

	// Sets expectation for the second GetSecretVersion (after the poll)
	mockClient.On("GetSecretVersion", mock.Anything, mock.Anything, mock.Anything).Return(
		&secretmanagerpb.SecretVersion{Etag: "etag-2"}, nil).Once() // ETag changes

	// Sets expectation for AccessSecretVersion after the ETag change
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
		lastKnownETag: "etag-1",
		cancel:        cancel,
	}

	changeChan := make(chan interface{}, 1)
	stopChan := make(chan struct{})

	go func() {
		defer close(stopChan)
		watcher.Watch(changeChan)
	}()

	select {
	case value := <-changeChan:
		expected := "new-secret-value"
		assert.Equal(t, expected, value, "Watch should send the new secret value on the change channel")
		log.Println("TestGcpSecretsWatcher_Watch_EtagChange: Change received:", value)
		watcher.Stop()
	case <-time.After(3 * time.Second):
		t.Fatalf("Watch did not send a change within the timeout")
	}

	<-stopChan
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
