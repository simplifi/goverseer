package gce_metadata_watcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/logger"
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
	DefaultMetadataErrorWaitSeconds = 10
)

// Config is the configuration for a GCE metadata watcher
type Config struct {
	// Source is the metadata source to watch
	// Valid values are 'instance' and 'project'
	// Default is 'instance'
	Source string `mapstructure:"source" validate:"oneof=instance project"`

	// Key is the key to watch in the GCE metadata
	// This is required config value
	Key string `mapstructure:"key" validate:"required"`

	// Recursive is whether to recurse the metadata keys
	// Default is false
	Recursive bool `mapstructure:"recursive"`

	// MetadataUrl is the URL this watcher will use when reading from the GCE
	// metadata server
	// It can be useful to override during testing
	// e.g. http://localhost:8888/computeMetadata/v1
	MetadataUrl string `mapstructure:"metadata_url" validate:"required"`

	// MetadataErrorWaitSeconds is the number of seconds to wait before retrying
	// a failed metadata request. This prevents hammering the metadata server.
	// Default is 10 seconds
	MetadataErrorWaitSeconds int `mapstructure:"metadata_error_wait_seconds" validate:"gte=1"`
}

// ParseConfig parses the config for the watcher
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(input interface{}) (*Config, error) {
	cfg := &Config{
		Source:                   DefaultSource,
		Recursive:                DefaultRecursive,
		MetadataUrl:              DefaultMetadataUrl,
		MetadataErrorWaitSeconds: DefaultMetadataErrorWaitSeconds,
	}

	if err := config.Decode(input, cfg); err != nil {
		return nil, err
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
			Source:                   pcfg.Source,
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
	logger.Log.Info("starting watcher")

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

				logger.Log.Error("error getting metadata", "err", err)

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
				logger.Log.Info("change detected",
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
	logger.Log.Info("shutting down watcher")
	w.cancel()
}
