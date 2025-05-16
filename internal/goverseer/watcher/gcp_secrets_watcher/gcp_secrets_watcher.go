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
	// Projects is a list of GCP project IDs to monitor
	Projects []string

	// SecretName is the name of the secret to watch in each project
	SecretName string

	// CredentialsFile is the path to the GCP credentials file
	// If not set, the default credentials will be used
	CredentialsFile string

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

	// Must have at least one project in the form of a string
	if projectsRaw, ok := config["projects"]; ok {
        projectsSlice, isSlice := projectsRaw.([]interface{})
        if !isSlice {
            return nil, fmt.Errorf("projects must be a list")
        }
        if len(projectsSlice) == 0 {
            return nil, fmt.Errorf("at least one project must be specified")
        }
        var projects []string
        for _, p := range projectsSlice {
            project, isString := p.(string)
            if !isString {
                return nil, fmt.Errorf("project name must be a string")
            }
            projects = append(projects, project)
        }
        cfg.Projects = projects
    } else {
        return nil, fmt.Errorf("projects is required")
    }

	// Secret name is required and must be a string
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

	// Credentials file is optional and must be a string
	// If not set, the default ADC will be used
	if credentialsFileRaw, ok := config["credentials_file"].(string); ok {
		if credentialsFileRaw == "" {
			return nil, fmt.Errorf("credentials_file cannot be empty")
		}
		cfg.CredentialsFile = credentialsFileRaw
	} else if config["credentials_file"] != nil {
		return nil, fmt.Errorf("credentials_file must be a string")
	}

	// Check interval is optional and must be a positive integer
	// If not set, the default is 60 seconds
	if checkIntervalSecondsRaw, ok := config["check_interval_seconds"].(int); ok {
		if checkIntervalSecondsRaw <= 0 {
			return nil, fmt.Errorf("check_interval_seconds must be a positive integer")
		}
		cfg.CheckIntervalSeconds = checkIntervalSecondsRaw
	} else if config["check_interval_seconds"] != nil {
		return nil, fmt.Errorf("check_interval_seconds must be an integer")
	}

	// Secret error wait seconds is optional and must be a positive integer
	// If not set, the default is 5 seconds
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

	// lastKnownETags stores the last known ETag for each project's secret
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
    var client *secretmanager.Client

    // Creates a new GCP Secrets Manager client
    // If a credentials file is provided, uses it
    // Otherwise, uses the default credentials
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
        Config:         *cfg,
        lastKnownETags: make(map[string]string),
        client:         client,
        ctx:            derivedCtx,
        cancel:         cancel,
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

// Retrieves the latest value and ETag of the secret
// from GCP Secrets Manager
func (w *GcpSecretsWatcher) getSecretValueAndEtag(project string) (string, string, error) {
    name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", project, w.SecretName)
    req := &secretmanagerpb.AccessSecretVersionRequest{
        Name: name,
    }

    resp, err := w.client.AccessSecretVersion(w.ctx, req)
    if err != nil {
        return "", "", fmt.Errorf("failed to access secret %s in %s: %v", w.SecretName, project, err)
    }

	// TODO: Check if this is the best approach for getting the ETag, which is not part of the AccessSecretVersionResponse
    etag := resp.ProtoReflect().Get(resp.ProtoReflect().Descriptor().Fields().ByName("Etag")).String()
    payload := string(resp.GetPayload().GetData())

    return payload, etag, nil
}

// Watches the GCP Secrets Manager for changes in ETag and sends
// the project ID and the secret payload to the changes channel
func (w *GcpSecretsWatcher) Watch(change chan interface{}) {
    log.Println("starting GCP Secrets Manager watcher (using ETags for change detection)")

    w.lastKnownETags = make(map[string]string)

    for {
        select {
        case <-w.ctx.Done():
            log.Println("GCP Secrets Manager watcher (using NewClient) stopped")
            return
        default:
            etagChanged := false
            updatedProjects := make(map[string]string)

            for _, project := range w.Projects {
                secretValue, etag, err := w.getSecretValueAndEtag(project)
                if err != nil {
                    log.Printf("error getting secret and ETag for %s: %v", project, err)
                    time.Sleep(time.Duration(w.SecretErrorWaitSeconds) * time.Second)
                    continue
                }

                lastEtag, known := w.lastKnownETags[project]
                if !known || etag != lastEtag {
                    log.Printf("ETag changed for secret %s in project %s", w.SecretName, project)
                    etagChanged = true
                    updatedProjects[project] = secretValue
                    w.lastKnownETags[project] = etag
                }
            }

            if etagChanged {
                change <- updatedProjects
            }

            time.Sleep(time.Duration(w.CheckIntervalSeconds) * time.Second)
        }
    }
}

// Called when the watcher is no longer needed
// This function signals to the goroutine 
// to close the Secret Manager client
func (w *GcpSecretsWatcher) Stop() {
	log.Println("shutting down GCP Secrets Manager watcher (using NewClient)")
	w.cancel()
}