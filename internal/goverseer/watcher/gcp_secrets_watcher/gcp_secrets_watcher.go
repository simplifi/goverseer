package gcp_secrets_watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
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
	CredentialFile string

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

// Parses the config for the watcher
// Validates the config, sets defaults if missing, and returns the config
func ParseConfig(config map[string]interface{}) (*Config, error) {
	cfg := &Config{
		CheckIntervalSeconds:    DefaultCheckIntervalSeconds,
		SecretErrorWaitSeconds: DefaultSecretErrorWaitSeconds,
	}

	if projectIDRaw, ok := config["project_id"].(string); ok {
		if projectIDRaw == "" {
			return nil, fmt.Errorf("project_id must not be empty")
		}
		cfg.ProjectID = projectIDRaw
	} else {
		return nil, fmt.Errorf("project_id is required")
	}

	if secretNameRaw, ok := config["secret_name"].(string); ok {
		if secretNameRaw == "" {
			return nil, fmt.Errorf("secret_name must not be empty")
		}
		cfg.SecretName = secretNameRaw
	} else {
		return nil, fmt.Errorf("secret_name is required")
	}

	if credentialFileRaw, ok := config["credential_file"].(string); ok {
		cfg.CredentialFile = credentialFileRaw
	}

	if checkIntervalSecondsRaw, ok := config["check_interval_seconds"].(int); ok {
		if checkIntervalSecondsRaw <= 0 {
			return nil, fmt.Errorf("check_interval_seconds must be a positive integer")
		}
		cfg.CheckIntervalSeconds = checkIntervalSecondsRaw
	}

	if secretErrorWaitSecondsRaw, ok := config["secret_error_wait_seconds"].(int); ok {
		if secretErrorWaitSecondsRaw <= 0 {
			return nil, fmt.Errorf("secret_error_wait_seconds must be a positive integer")
		}
		cfg.SecretErrorWaitSeconds = secretErrorWaitSecondsRaw
	}

	if secretsFilePathRaw, ok := config["secrets_file_path"].(string); ok {
		if secretsFilePathRaw == "" {
			return nil, fmt.Errorf("secrets_file_path must not be empty")
		}
		cfg.SecretsFilePath = secretsFilePathRaw
	} else {
		return nil, fmt.Errorf("secrets_file_path is required")
	}

	return cfg, nil
}

type GcpSecretsWatcher struct {
	Config
	lastKnownETag string
	client        *secretmanager.Client
	ctx           context.Context
	cancel        context.CancelFunc
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
	if cfg.CredentialFile != "" {
		client, err = secretmanager.NewClient(ctx, option.WithCredentialsFile(cfg.CredentialFile))
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

// Watches the GCP Secrets Manager for changes in ETag and sends the new value
// to the changes channel
func (w *GcpSecretsWatcher) Watch(change chan interface{}) {
	log.Printf("starting GCP Secrets Manager watcher for project: %s, secret: %s", w.ProjectID, w.SecretName)

	for {
		select {
		case <-w.ctx.Done():
			log.Println("GCP Secrets Manager watcher stopped")
			return
		default:
			etag, err := w.getSecretEtag(w.ProjectID)
			if err != nil {
				log.Printf("error getting secret ETag: %v", err)
				time.Sleep(time.Duration(w.SecretErrorWaitSeconds) * time.Second)
				continue
			}

			if etag != w.lastKnownETag {
				log.Printf("ETag changed for secret %s in project %s", w.SecretName, w.ProjectID)
				secretValue, err := w.getSecretValue(w.ProjectID)
				if err != nil {
					log.Printf("error getting secret value: %v", err)
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