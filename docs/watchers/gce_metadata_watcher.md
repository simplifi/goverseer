# GCE Metadata Watcher

The GCE Metadata Watcher allows you to monitor a Google Compute Engine (GCE)
metadata key for changes. When a change is detected, Goverseer can trigger an
executioner to take action based on the updated metadata value. The changed data
is passed to the executioner for processing.

## Configuration

To use the GCE Metadata Watcher, you need to configure it in your Goverseer
config file. The following configuration options are available:

- `source`: (Optional) This is the source of the metadata value you want to
  monitor. It must be set to either `instance` or `project`. Defaults to
  `instance` if not provided.
- `key`: This is the GCE metadata key you want to monitor. For example,
  `instance/attributes/my-key`.
- `recursive`: (Optional) This determines whether to fetch metadata recursively.
  If set to `true`, all subkeys under the specified key will be monitored.
  Defaults to `false`.
- `metadata_url`: (Optional) This allows overriding the default GCE Metadata
  server URL. Useful for testing with a local server. Defaults to
  `http://metadata.google.internal/computeMetadata/v1`.
- `metadata_error_wait_seconds`: (Optional) This determines the wait time in
  seconds before retrying after a metadata fetch error. Defaults to `10`.

**Example Configuration:**

```yaml
name: gce_metadata_watcher_example
watcher:
  gce_metadata:
    key: instance/attributes/my-key
    recursive: true
executioner:
  log:
```

**Note:**

- The GCE Metadata Watcher relies on the GCE Metadata server, which is only
  accessible from within a GCE instance.
- Ensure that your GCE instance has the necessary permissions to access the
  metadata server and the specified key.

This watcher is particularly useful for dynamically updating your application's
configuration based on changes made to instance metadata. For example, you can
use it to trigger a restart or reload when a new version is deployed, or to
adjust application behavior based on metadata flags.

## Development

### Accessing GCE Metadata Locally

It can be useful to test your code locally while listening to an actual GCE
Metadata endpoint. You can achieve this by creating a tunnel to a GCE instance
using `gcloud ssh`.

The snippet below creates a tunnel to your local machine on port 8888 from your
instance on port 80:

```bash
INSTANCE_NAME=my-instance-name
PROJECT=sbx-your-sandbox
ZONE=us-central1-a
gcloud compute ssh "${INSTANCE_NAME}" \
  --project="${PROJECT}" \
  --zone="${ZONE}" \
  --internal-ip -- \
  -L 8888:metadata.google.internal:80
```

You can then make metadata changes on the instance to test goverseer:

```bash
INSTANCE_NAME=my-instance-name
PROJECT=sbx-your-sandbox
ZONE=us-central1-a
gcloud compute instances add-metadata "${INSTANCE_NAME}" \
  --project="${PROJECT}" \
  --zone="${ZONE}" \
  --metadata foo=bar
```
