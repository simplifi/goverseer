package gcp_secrets_watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/logger"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

const (
	// Default interval to check for secret changes
	DefaultCheckIntervalSeconds = 60

	// Default number of seconds to wait
	// before retrying a failed secret access.
	DefaultSecretErrorWaitSeconds = 5
)

type Config struct {
	// GCP project ID where the secret is located
	ProjectID string

	// Name of the secret to watch in the specified project
	SecretName string

	// Path to the GCP credentials file
	// If not set, the default ADC will be used
	CredentialsFile string

	// Interval in seconds to poll the secret
	// Default is 60 seconds
	CheckIntervalSeconds int

	// Number of seconds to wait
	// before retrying a failed secret access
	// Default is 5 seconds
	SecretErrorWaitSeconds int

	// Path to the file to update with the secrets' value
	SecretsFilePath string
}

// Defines an interface for creating Secret Manager clients
// Helpful for testing purposes, allowing us to mock the client creation
type SecretManagerClientFactory interface {
	CreateClient(ctx context.Context, credentialsFile string) (SecretManagerClientInterface, error)
}

// Creates a real Secret Manager client
// Can be replaced with a mock implementation for testing
type defaultSecretManagerClientFactory struct{}

func (f *defaultSecretManagerClientFactory) CreateClient(ctx context.Context, credentialsFile string) (SecretManagerClientInterface, error) {
	var client *secretmanager.Client
	var err error
	if credentialsFile != "" {
		client, err = secretmanager.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		client, err = secretmanager.NewClient(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Secrets Manager client: %w", err)
	}
	return client, nil
}

// Defines the methods from the Secret Manager
// client that GcpSecretsWatcher uses.
type SecretManagerClientInterface interface {
	GetSecretVersion(ctx context.Context, req *secretmanagerpb.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

type GcpSecretsWatcher struct {
	Config
	lastKnownETag string
	client        SecretManagerClientInterface
	ctx           context.Context
	cancel        context.CancelFunc
	clientFactory SecretManagerClientFactory
}

// Parses a required string field from config
// Returns an error if the field is missing or not a string
// Also checks if the string is empty
// (Used for project_id, secret_name, and secrets_file_path)
func parseRequiredString(cfgMap map[string]interface{}, fieldName string) (string, error) {
	if raw, ok := cfgMap[fieldName]; ok {
		if val, isString := raw.(string); isString {
			if val == "" {
				return "", fmt.Errorf("%s must not be empty", fieldName)
			}
			return val, nil
		}
		return "", fmt.Errorf("%s must be a string", fieldName)
	}
	return "", fmt.Errorf("%s is required", fieldName)
}

// Parses an optional string field from config
// Returns an error if the field is not a string
// Also checks if the string is empty
// (Used for credentials_file)
func parseOptionalString(cfgMap map[string]interface{}, fieldName string) (string, error) {
	if raw, ok := cfgMap[fieldName]; ok {
		if val, isString := raw.(string); isString {
			return val, nil
		}
		return "", fmt.Errorf("%s must be a string", fieldName)
	}
	return "", nil
}

// Parses an optional positive integer field from config
// Returns an error if the field is not an integer
// Also checks if the integer is positive
// (Used for check_interval_seconds and secret_error_wait_seconds)
func parseOptionalPositiveInt(cfgMap map[string]interface{}, fieldName string) (int, error) {
	if raw, ok := cfgMap[fieldName]; ok {
		if val, isInt := raw.(int); isInt {
			if val <= 0 {
				return 0, fmt.Errorf("%s must be a positive integer", fieldName)
			}
			return val, nil
		}
		return 0, fmt.Errorf("%s must be an integer", fieldName)
	}
	return 0, nil
}

// Parses and validates the config for the watcher,
// sets defaults if missing, and returns the config
func ParseConfig(config map[string]interface{}) (*Config, error) {
	cfg := &Config{
		CheckIntervalSeconds:   DefaultCheckIntervalSeconds,
		SecretErrorWaitSeconds: DefaultSecretErrorWaitSeconds,
	}
	var err error

	var val int

	cfg.ProjectID, err = parseRequiredString(config, "project_id")
	if err != nil {
		return nil, err
	}

	cfg.SecretName, err = parseRequiredString(config, "secret_name")
	if err != nil {
		return nil, err
	}

	cfg.CredentialsFile, err = parseOptionalString(config, "credentials_file")
	if err != nil {
		return nil, err
	}

	if val, err = parseOptionalPositiveInt(config, "check_interval_seconds"); err != nil {
		return nil, err
	} else if val != 0 {
		cfg.CheckIntervalSeconds = val
	}

	if val, err = parseOptionalPositiveInt(config, "secret_error_wait_seconds"); err != nil {
		return nil, err
	} else if val != 0 {
		cfg.SecretErrorWaitSeconds = val
	}

	cfg.SecretsFilePath, err = parseRequiredString(config, "secrets_file_path")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Creates a new GcpSecretsWatcher based on the passed config
func New(config map[string]interface{}, factory ...SecretManagerClientFactory) (*GcpSecretsWatcher, error) { // MODIFIED SIGNATURE
	cfg, err := ParseConfig(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var clientFactory SecretManagerClientFactory

	// Checks if factory was provided
	if len(factory) > 0 && factory[0] != nil {
		clientFactory = factory[0]
	} else {
		// Uses the default factory if none was provided
		clientFactory = &defaultSecretManagerClientFactory{}
	}

	// Uses the factory to create client
	client, err := clientFactory.CreateClient(ctx, cfg.CredentialsFile)
	if err != nil {

		return nil, fmt.Errorf("failed to create Secrets Manager client: %w", err)
	}

	derivedCtx, cancel := context.WithCancel(ctx)

	watcher := &GcpSecretsWatcher{
		Config:        *cfg,
		lastKnownETag: "",
		client:        client,
		ctx:           derivedCtx,
		cancel:        cancel,
		clientFactory: clientFactory,
	}

	go func() {
		<-watcher.ctx.Done()
		if watcher.client != nil {
			if err := watcher.client.Close(); err != nil {
				logger.Log.Error("error closing Secrets Manager client", "err", err)
			}
		}
	}()

	return watcher, nil
}

// Retrieves the latest ETag of the secret from GCP Secrets Manager
func (w *GcpSecretsWatcher) getSecretEtag(projectID string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, w.SecretName)
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: name,
	}

	resp, err := w.client.GetSecretVersion(w.ctx, req)
	if err != nil {

		return "", fmt.Errorf("failed to access secret version %s in %s: %w", w.SecretName, projectID, err)
	}
	return resp.Etag, nil
}

// Retrieves the latest value of the secret from GCP Secrets Manager
func (w *GcpSecretsWatcher) getSecretValue(projectID string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, w.SecretName)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	resp, err := w.client.AccessSecretVersion(w.ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s in %s: %v", w.SecretName, projectID, err)
	}
	return string(resp.Payload.Data), nil
}

// Watches the GCP Secrets Manager for changes in ETag
// and sends the new value to the changes channel
func (w *GcpSecretsWatcher) Watch(change chan interface{}) {
	logger.Log.Info("starting GCP Secrets Manager watcher for project: %s, secret: %s", w.ProjectID, w.SecretName)

	for {
		select {
		case <-w.ctx.Done():
			logger.Log.Info("GCP Secrets Manager watcher stopped")
			return
		default:
			// Gets ETag
			etag, err := w.getSecretEtag(w.ProjectID)
			if err != nil {
				logger.Log.Error("Failed to get ETag", "secret", w.SecretName, "project", w.ProjectID, "err", err)
				time.Sleep(time.Duration(w.SecretErrorWaitSeconds) * time.Second)
				continue
			}

			// Checks for ETag change
			if etag != w.lastKnownETag {
				logger.Log.Info("ETag changed", "secret", w.SecretName, "project", w.ProjectID, "err", err)

				// Gets Secret Value (only if ETag changed)
				secretValue, err := w.getSecretValue(w.ProjectID)
				if err != nil {
					logger.Log.Error("Failed to get secret value after ETag change", "secret", w.SecretName, "project", w.ProjectID, "old_etag", w.lastKnownETag, "error", err)
					time.Sleep(time.Duration(w.SecretErrorWaitSeconds) * time.Second)
					continue
				}

				change <- secretValue
				w.lastKnownETag = etag
			}

			time.Sleep(time.Duration(w.CheckIntervalSeconds) * time.Second)
		}
	}
}

// Signals the watcher to stop
func (w *GcpSecretsWatcher) Stop() {
	logger.Log.Info("shutting down watcher")
	w.cancel()
}
