package gcp_secrets_watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const (
	// DefaultCheckIntervalSeconds is the default interval to check for secret changes
	DefaultCheckIntervalSeconds = 60

	// DefaultSecretErrorWaitSeconds is the default number of seconds to wait
	// before retrying a failed secret access.
	DefaultSecretErrorWaitSeconds = 5
)

type Config struct {
	// Projects is a list of GCP project IDs to monitor
	Projects []string

	// SecretName is the name of the secret to watch in each project
	SecretName string

	// The path to the GCP service account credential file
	// If empty, it will rely on Application Default Credentials
	CredentialFile string

	// The interval in seconds to poll the secret
	// Default is 60 seconds
	CheckIntervalSeconds int

	// SecretErrorWaitSeconds is the number of seconds to wait
	// before retrying a failed secret access
	// Default is 5 seconds
	SecretErrorWaitSeconds int
}

// Parses the config for the watcher
// Validates the config, sets defaults
// if missing, and returns the config
func ParseConfig(config map[string]interface{}) (*Config, error) {
	cfg := &Config{
		CheckIntervalSeconds:   DefaultCheckIntervalSeconds,
		SecretErrorWaitSeconds: DefaultSecretErrorWaitSeconds,	
	}

	if projectsRaw, ok := config["projects"].([]interface{}); ok {
		var projects []string
		for _, p := range projectsRaw {
			if project, ok := p.(string); ok {
				if project == "" {
					return nil, fmt.Errorf("project name cannot be empty")
				}
				projects = append(projects, project)
			} else {
				return nil, fmt.Errorf("projects must be a list of strings")
			}
		}
		if len(projects) == 0 {
			return nil, fmt.Errorf("at least one project must be specified")
		}
		cfg.Projects = projects
	} else if config["projects"] != nil {
		return nil, fmt.Errorf("projects must be a list")
	} else {
		return nil, fmt.Errorf("projects is required")
	}

	if secretNameRaw, ok := config["secret_name"].(string); ok {
		if secretNameRaw == "" {
			return nil, fmt.Errorf("secret_name cannot be empty")
		}
		cfg.SecretName = secretNameRaw
	} else if config["secret_name"] != nil {
		return nil, fmt.Errorf("secret_name must be a string")
	} else {
		return nil, fmt.Errorf("secret_name is required")
	}

	if credentialFileRaw, ok := config["credential_file"].(string); ok {
		cfg.CredentialFile = credentialFileRaw
	} else if config["credential_file"] != nil {
		return nil, fmt.Errorf("credential_file must be a string")
	}

	if checkIntervalSecondsRaw, ok := config["check_interval_seconds"].(int); ok {
		if checkIntervalSecondsRaw <= 0 {
			return nil, fmt.Errorf("check_interval_seconds must be a positive integer")
		}
		cfg.CheckIntervalSeconds = checkIntervalSecondsRaw
	} else if config["check_interval_seconds"] != nil {
		return nil, fmt.Errorf("check_interval_seconds must be an integer")
	}

	if secretErrorWaitSecondsRaw, ok := config["secret_error_wait_seconds"].(int); ok {
		if secretErrorWaitSecondsRaw <= 0 {
			return nil, fmt.Errorf("secret_error_wait_seconds must be a positive integer")
		}
		cfg.SecretErrorWaitSeconds = secretErrorWaitSecondsRaw
	} else if config["secret_error_wait_seconds"] != nil {
		return nil, fmt.Errorf("secret_error_wait_seconds must be an integer")
	}

	return cfg, nil
}

type GcpSecretsWatcher struct {
	Config

	// lastKnownValues stores the last known secret value for each project
	lastKnownValues map[string]string

	// lastknownETags stores the last known etag for each project
	lastKnownETags map[string]string

	// client is the GCP Secrets Manager client
	client *secretmanager.Client

	// ctx is the context
	ctx context.Context

	// cancel is the cancel function used to stop the watcher
	cancel context.CancelFunc
}

// Creates a new GcpSecretsWatcher based on the passed config
func New(config map[string]interface{}) (*GcpSecretsWatcher, error) {
	cfg, err := ParseConfig(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secrets Manager client: %v", err)
	}

	derivedCtx, cancel := context.WithCancel(ctx)

	watcher := &GcpSecretsWatcher{
		Config:          *cfg,
		lastKnownValues: make(map[string]string),
		lastKnownETags:  make(map[string]string),
		client:          client,
		ctx:             derivedCtx,
		cancel:          cancel,
	}
	
	// Closes the client when the watcher is stopped
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

// Retrieves the latest value of the secret 
// from GCP Secrets Manager
func (w *GcpSecretsWatcher) getSecretValue(project string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", project, w.SecretName)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	resp, err := w.client.AccessSecretVersion(w.ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s in %s: %v", w.SecretName, project, err)
	}
	return string(resp.Payload.Data), nil
}

// Watches the GCP Secrets Manager for changes and 
// sends the new value to the changes channel
func (w *GcpSecretsWatcher) Watch(change chan interface{}) {
    log.Println("starting GCP Secrets Manager watcher (using NewClient)")

    w.lastKnownValues = make(map[string]string)

    for {
        select {
        case <-w.ctx.Done():
            log.Println("GCP Secrets Manager watcher (using NewClient) stopped")
            return
        default:
            currentValues := make(map[string]string)
            changed := false

			// Iterate over each project and get the secret value
            for _, project := range w.Projects {
                secretValue, err := w.getSecretValue(project)
                if err != nil {
                    log.Printf("error getting secret for %s (using NewClient): %v", project, err)
                    time.Sleep(time.Duration(w.SecretErrorWaitSeconds) * time.Second)
                    continue
                }
                currentValues[project] = secretValue

                lastValue, known := w.lastKnownValues[project]
                if !known || secretValue != lastValue {
                    log.Printf("change detected in project %s for secret %s (using NewClient)", project, w.SecretName)
                    changed = true
                }
            }

            if changed {
                change <- currentValues
                w.lastKnownValues = currentValues
            }

            time.Sleep(time.Duration(w.CheckIntervalSeconds) * time.Second)
        }
    }
}

// Stops the watcher and closes the client
// This function is called when the watcher is stopped
// It signals to the goroutine to close the Secret Manager
// client (see the New function above)
func (w *GcpSecretsWatcher) Stop() {
	log.Println("shutting down GCP Secrets Manager watcher (using NewClient)")
	w.cancel()
}