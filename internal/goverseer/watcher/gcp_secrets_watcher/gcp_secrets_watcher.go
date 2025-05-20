package gcp_secrets_watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

const (
	// DefaultCheckIntervalSeconds is the default interval to check for secret changes
	DefaultCheckIntervalSeconds = 60

	// DefaultSecretErrorWaitSeconds is the default number of seconds to wait
	// before retrying a failed secret access.
	DefaultSecretErrorWaitSeconds = 5
)

type Config struct {
	// ProjectID is the GCP project ID where the secret is located
	ProjectID string

	// SecretName is the name of the secret to watch in the specified project
	SecretName string

	// CredentialsFile is the path to the GCP credentials file
	// If not set, the default ADC will be used
	CredentialsFile string

	// The interval in seconds to poll the secret
	// Default is 60 seconds
	CheckIntervalSeconds int

	// SecretErrorWaitSeconds is the number of seconds to wait
	// before retrying a failed secret access
	// Default is 5 seconds
	SecretErrorWaitSeconds int

	// SecretsFilePath is the path to the file to update on the local VM
	SecretsFilePath string
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
}

// Parses a required string field from config
func parseRequiredString(cfgMap map[string]interface{}, fieldName string) (string, error) {
	if raw, ok := cfgMap[fieldName]; ok {
		if val, isString := raw.(string); isString {
			if val == "" {
				return "", fmt.Errorf("%s must not be empty and is required", fieldName)
			}
			return val, nil
		}
		return "", fmt.Errorf("%s must be a string", fieldName)
	}
	return "", fmt.Errorf("%s is required", fieldName)
}

// Parses an optional string field from config
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

	if val, err := parseOptionalPositiveInt(config, "check_interval_seconds"); err != nil {
		return nil, err
	} else if val != 0 {
		cfg.CheckIntervalSeconds = val
	}

	if val, err := parseOptionalPositiveInt(config, "secret_error_wait_seconds"); err != nil {
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
func New(config map[string]interface{}) (*GcpSecretsWatcher, error) {
	cfg, err := ParseConfig(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var client *secretmanager.Client

	// Creates a new GCP Secrets Manager client
	if cfg.CredentialsFile != "" {
		client, err = secretmanager.NewClient(ctx, option.WithCredentialsFile(cfg.CredentialsFile))
	} else {
		client, err = secretmanager.NewClient(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Secrets Manager client: %v", err)
	}

	derivedCtx, cancel := context.WithCancel(ctx)

	watcher := &GcpSecretsWatcher{
		Config:        *cfg,
		lastKnownETag: "",
		client:        client,
		ctx:           derivedCtx,
		cancel:        cancel,
	}

	go func() {
		<-watcher.ctx.Done()
		if watcher.client != nil {
			if err := watcher.client.Close(); err != nil {
				log.Printf("error closing Secrets Manager client: %v", err)
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
		return "", fmt.Errorf("failed to access secret version %s in %s: %v", w.SecretName, projectID, err)
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
    log.Printf("starting GCP Secrets Manager watcher for project: %s, secret: %s", w.ProjectID, w.SecretName)

    for {
        select {
        case <-w.ctx.Done():
            log.Println("GCP Secrets Manager watcher stopped")
            return
        default:
            // Gets ETag
            etag, err := w.getSecretEtag(w.ProjectID)
            if err != nil {
                log.Printf("ERROR: Failed to get ETag for secret %s in %s: %v", w.SecretName, w.ProjectID, err)
                time.Sleep(time.Duration(w.SecretErrorWaitSeconds) * time.Second)
                continue
            }

            // Checks for ETag change
            if etag != w.lastKnownETag {
                log.Printf("ETag changed for secret %s in project %s", w.SecretName, w.ProjectID)

                // Gets Secret Value (only if ETag changed)
                secretValue, err := w.getSecretValue(w.ProjectID)
                if err != nil {
                    log.Printf("ERROR: Failed to get secret value for %s in %s: %v", w.SecretName, w.ProjectID, err)
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

// Stop signals the watcher to stop
func (w *GcpSecretsWatcher) Stop() {
	log.Println("shutting down watcher")
	w.cancel()
}
