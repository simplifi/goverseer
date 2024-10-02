package gce_metadata_watcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

const (
	// ValidSourceInstance is the string value for an instance metadata source
	ValidSourceInstance = "instance"

	// ValidSourceProject is the string value for a project metadata source
	ValidSourceProject = "project"

	// DefaultSource is the default metadata source
	DefaultSource = ValidSourceInstance

	// DefaultRecursive is the default value for the recursive flag
	// it can be overridden by setting the recursive flag in the config
	DefaultRecursive = false

	// DefaultMetadataUrl is the default URL for GCE metadata
	DefaultMetadataUrl = "http://metadata.google.internal/computeMetadata/v1"

	// DefaultMetadataErrorWaitSeconds is the default number of seconds to wait
	// before retrying a failed metadata request
	DefaultMetadataErrorWaitSeconds = 1
)

// Config is the configuration for a GCE metadata watcher
type Config struct {
	// Source is the metadata source to watch
	// Valid values are 'instance' and 'project'
	// Default is 'instance'
	Source string

	// Key is the key to watch in the GCE metadata
	// This is required config value
	Key string

	// Recursive is whether to recurse the metadata keys
	// Default is false
	Recursive bool

	// MetadataUrl is the URL this watcher will use when reading from the GCE
	// metadata server
	// It can be useful to override during testing
	// e.g. http://localhost:8888/computeMetadata/v1
	MetadataUrl string

	// MetadataErrorWaitSeconds is the number of seconds to wait before retrying
	// a failed metadata request. This prevents hammering the metadata server.
	// Default is 1 second
	MetadataErrorWaitSeconds int
}

// ParseConfig parses the config for the watcher
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(config interface{}) (*Config, error) {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	cfg := &Config{
		Source:                   DefaultSource,
		Recursive:                DefaultRecursive,
		MetadataUrl:              DefaultMetadataUrl,
		MetadataErrorWaitSeconds: DefaultMetadataErrorWaitSeconds,
	}

	// If source is set, it should be one of the valid sources
	if cfgMap["source"] != nil {
		if source, ok := cfgMap["source"].(string); ok {
			if source != ValidSourceInstance && source != ValidSourceProject {
				return nil, fmt.Errorf("source must be one of %s or %s", ValidSourceInstance, ValidSourceProject)
			}
			cfg.Source = source
		} else if cfgMap["source"] != nil {
			return nil, fmt.Errorf("source must be a string")
		}
	}

	// If recursive is set, it should be a boolean
	if cfgMap["recursive"] != nil {
		if recursive, ok := cfgMap["recursive"].(bool); ok {
			cfg.Recursive = recursive
		} else if cfgMap["recursive"] != nil {
			return nil, fmt.Errorf("recursive must be a boolean")
		}
	}

	// Key is required and must be a string
	if key, ok := cfgMap["key"].(string); ok {
		if key == "" {
			return nil, fmt.Errorf("key must not be empty")
		}
		cfg.Key = key
	} else if cfgMap["key"] != nil {
		return nil, fmt.Errorf("key must be a string")
	} else {
		return nil, fmt.Errorf("key is required")
	}

	// If metadata_url is set, it should be a string
	if cfgMap["metadata_url"] != nil {
		if metadataUrl, ok := cfgMap["metadata_url"].(string); ok {
			if metadataUrl == "" {
				return nil, fmt.Errorf("metadata_url must not be empty")
			}
			cfg.MetadataUrl = metadataUrl
		} else if cfgMap["metadata_url"] != nil {
			return nil, fmt.Errorf("metadata_url must be a string")
		}
	}

	// If metadata_error_wait_seconds is set, it should be an integer
	if cfgMap["metadata_error_wait_seconds"] != nil {
		if metadataErrorWaitSeconds, ok := cfgMap["metadata_error_wait_seconds"].(int); ok {
			cfg.MetadataErrorWaitSeconds = metadataErrorWaitSeconds
		} else if cfgMap["metadata_error_wait_seconds"] != nil {
			return nil, fmt.Errorf("metadata_error_wait_seconds must be an integer")
		}
	}

	return cfg, nil
}

type GceMetadataWatcher struct {
	Config

	// lastETag is the last etag, used to compare changes
	lastETag string

	// ctx is the context
	ctx context.Context

	// cancel is the cancel function used to stop the watcher
	cancel context.CancelFunc
}

// New creates a new GceMetadataWatcher based on the passed config
func New(cfg config.Config) (*GceMetadataWatcher, error) {
	pcfg, err := ParseConfig(cfg.Watcher.Config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &GceMetadataWatcher{
		Config: Config{
			Key:                      pcfg.Key,
			Recursive:                pcfg.Recursive,
			MetadataUrl:              pcfg.MetadataUrl,
			MetadataErrorWaitSeconds: pcfg.MetadataErrorWaitSeconds,
		},
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// gceMetadataResponse is the response from the GCE metadata server
type gceMetadataResponse struct {
	// etag is the etag of the metadata
	// used to compare changes
	etag string

	// body is the body of the metadata
	body string
}

// getMetadata gets the metadata from the GCE metadata server
// It returns the metadata response or an error
func (w *GceMetadataWatcher) getMetadata() (*gceMetadataResponse, error) {
	client := http.Client{
		Timeout: 0, // No timeout (infinite)
	}

	urlWithKey := fmt.Sprintf("%s/%s", w.MetadataUrl, w.Key)
	req, err := http.NewRequestWithContext(w.ctx, "GET", urlWithKey, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Metadata-Flavor", "Google")
	q := req.URL.Query()
	q.Add("wait_for_change", "true")
	q.Add("recursive", fmt.Sprintf("%v", w.Recursive))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return &gceMetadataResponse{
		etag: resp.Header.Get("ETag"),
		body: string(body),
	}, nil
}

// Watch watches the GCE metadata for changes and sends value to changes channel
// The changes channel is where the value is sent when it changes
func (w *GceMetadataWatcher) Watch(change chan interface{}) {
	log.Info("starting watcher")

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			gceMetadata, err := w.getMetadata()
			if err != nil {
				// Avoid logging errors if the context was canceled mid-request
				// This will happen when the watcher is stopped
				if w.ctx.Err() == context.Canceled {
					continue
				}

				log.Error("error getting metadata", "err", err)

				// Usually getMetadata opens up a connection to the metadata server
				// and waits for a change. If there is an error we want to wait for a
				// bit before trying again to prevent hammering the metadata server.
				// Since we're in a for loop here the retrys will come VERY fast without
				// this sleep.
				time.Sleep(time.Duration(w.MetadataErrorWaitSeconds) * time.Second)
				continue
			}

			// Only send a change if it has actually changed by comparing etags
			if w.lastETag != gceMetadata.etag {
				log.Info("change detected",
					"key", w.Key,
					"etag", gceMetadata.etag,
					"previous_etag", w.lastETag)

				change <- gceMetadata.body

				w.lastETag = gceMetadata.etag
			}
		}
	}
}

// Stop signals the watcher to stop
func (w *GceMetadataWatcher) Stop() {
	log.Info("shutting down watcher")
	w.cancel()
}
