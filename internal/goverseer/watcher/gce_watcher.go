package watcher

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Watcher interface
	_ Watcher = (*GCEWatcher)(nil)
)

func init() {
	RegisterWatcher("gce", func() Watcher { return &GCEWatcher{} })
}

const (
	// metadataErrWait is the time to wait before trying a failed metadata request
	// again, this prevents hammering the metadata server
	metadataErrWait = 1 * time.Second
)

// GCEWatcher watches a GCE metadata key for changes
type GCEWatcher struct {
	// Key is the key to watch in the GCE metadata
	Key string

	// Recursive is whether to recurse the metadata keys
	Recursive bool

	// metadataUrl is the URL this watcher will use when reading from the GCE
	// metadata server
	MetadataUrl string

	// lastETag is the last etag, used to compare changes
	lastETag string

	// log is the logger
	log *slog.Logger

	// ctx is the context
	ctx context.Context

	// cancel is the cancel function used to stop the watcher
	cancel context.CancelFunc
}

func (w *GCEWatcher) Create(cfg config.WatcherConfig, log *slog.Logger) error {
	v, ok := cfg.(*config.GCEWatcherConfig)
	if !ok {
		return fmt.Errorf("invalid config for dummy watcher")
	}

	ctx, cancel := context.WithCancel(context.Background())

	w.Key = v.Key
	w.MetadataUrl = fmt.Sprintf("%s/%s/attributes", v.MetadataUrl, v.Source)
	w.Recursive = v.Recursive
	w.lastETag = ""
	w.log = log
	w.ctx = ctx
	w.cancel = cancel

	return nil
}

type gceMetadataResponse struct {
	etag string
	body string
}

func (w *GCEWatcher) getMetadata() (*gceMetadataResponse, error) {
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

	// Check for a non-200 status code
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
func (w *GCEWatcher) Watch(changes chan interface{}) {
	w.log.Info("starting watcher")

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			// Get the metadata
			gceMetadata, err := w.getMetadata()
			if err != nil {
				// Avoid logging errors if the context was canceled mid-request
				if w.ctx.Err() == context.Canceled {
					continue
				}

				w.log.Error("error getting metadata",
					tint.Err(err))
				time.Sleep(metadataErrWait)
				continue
			}

			// We only send a change if it has actually changed
			if w.lastETag != gceMetadata.etag {
				w.log.Info("change detected",
					"key", w.Key,
					"etag", gceMetadata.etag,
					"previous_etag", w.lastETag)

				// Send the new value to the changes channel
				changes <- gceMetadata.body

				// Update the last value with the current etag
				w.lastETag = gceMetadata.etag
			}
		}
	}
}

// Stop stops the watcher
func (w *GCEWatcher) Stop() {
	w.log.Info("shutting down watcher")
	w.cancel()
}
