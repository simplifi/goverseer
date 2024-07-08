package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/lmittmann/tint"
)

var (
	// Ensure this implements the Watcher interface
	_ Watcher = (*GCEWatcher)(nil)
)

const (
	// DefaultMetadataUrl is the default URL for GCE metadata
	// it can be overridden by setting the metadata_url in the config
	// this can be useful for testing
	// e.g. http://localhost:8888/computeMetadata/v1/
	DefaultMetadataUrl     = "http://metadata.google.internal/computeMetadata/v1/"
	metadataSourceInstance = "instance"
	metadataSourceProject  = "project"
)

// GCEWatcher watches a GCE metadata key for changes
type GCEWatcher struct {
	// Key is the key to watch in the GCE metadata
	Key string

	// Recursive is whether to recurse the metadata keys
	Recursive bool

	// client is the resty client used to make requests
	client *resty.Client

	// lastValue is the last value used to compare changes, it is set to the ETag
	lastValue string

	// log is the logger
	log *slog.Logger

	// ctx is the context
	ctx context.Context

	// cancel is the cancel function used to stop the watcher
	cancel context.CancelFunc
}

// NewGCEWatcher creates a new GCEWatcher
// Source must be either 'instance' or 'project'
// Key is the key to watch in the GCE metadata
// MetadataUrl is the URL to the GCE metadata server
// Recursive is whether to recurse the metadata keys
// log is the logger
func NewGCEWatcher(Source, Key, MetadataUrl string, Recursive bool, log *slog.Logger) (*GCEWatcher, error) {
	// Check that Source is either 'instance' or 'project'
	if Source != metadataSourceInstance && Source != metadataSourceProject {
		return nil, fmt.Errorf("source must be either 'instance' or 'project'")
	}

	// Set default MetadataUrl if not provided
	if MetadataUrl == "" {
		MetadataUrl = DefaultMetadataUrl
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := resty.New().
		SetBaseURL(fmt.Sprintf("%s/%s/attributes", MetadataUrl, Source)).
		SetHeader("Metadata-Flavor", "Google").
		SetTimeout(0) // No timeout (infinite)

	w := &GCEWatcher{
		Key:       Key,
		Recursive: Recursive,
		client:    client,
		lastValue: "",
		log:       log,
		ctx:       ctx,
		cancel:    cancel,
	}
	return w, nil
}

// SetLastValue sets the last value to the current ETag
func (w *GCEWatcher) SetLastValue() error {
	resp, err := w.client.R().
		SetContext(w.ctx).
		Get(w.Key)

	// If we are shutting down, return early to avoid other errors firing
	if w.ctx.Err() == context.Canceled {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to get etag: %w", err)
	}

	w.lastValue = resp.Header().Get("ETag")
	return nil
}

// awaitChange waits for a change in the GCE metadata
func (w *GCEWatcher) awaitChange() (string, error) {
	response, err := w.client.R().
		SetContext(w.ctx).
		SetQueryParam("wait_for_change", "true").
		SetQueryParam("last_etag", w.lastValue).
		SetQueryParam("recursive", fmt.Sprintf("%v", w.Recursive)).
		Get(w.Key)

	// If we are shutting down, return early to avoid other errors firing
	if w.ctx.Err() == context.Canceled {
		return "", nil
	}

	// Check for other errors.
	if err != nil {
		return "", err
	}

	return response.String(), nil
}

// Watch watches the GCE metadata for changes and sends value to changes channel
// The changes channel is where the value is sent when it changes
func (w *GCEWatcher) Watch(changes chan interface{}, wg *sync.WaitGroup) {
	defer w.cancel()
	defer wg.Done()

	w.log.Info("starting watcher")

	// TODO: When this errors, we might want to have a backoff
	// because it spams the shit out of the logs, and depending on the error it
	// could hammer the metadata server
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.log.Info("waiting for change", slog.String("key", w.Key))

			value, err := w.awaitChange()
			if err != nil {
				w.log.Error("error watching changes", slog.String("key", w.Key), tint.Err(err))
				continue
			}

			changes <- value

			if err := w.SetLastValue(); err != nil {
				w.log.Error("error updating last value", slog.String("key", w.Key), tint.Err(err))
				continue
			}
		}
	}
}

// Stop stops the watcher
func (w *GCEWatcher) Stop() {
	w.log.Info("shutting down watcher")
	w.cancel()
}
