# Goverseer

Goverseer is a tool for monitoring some kind of upstream, and taking actions
when a change is detected. In some ways it is similar to consul-template, execpt
Goverseer does not handle templating. Instead it is focused on watching for
changes and taking some kind of action based on those changes.

## Supported Watchers & Executors

Overseer currently supports:

* Dummy Watcher
* Dummy Executor

Planned future support includes:

* GCE Metadata Watcher
* Consul Watcher
* Command Executor

## Building

To build Goverseer, simply run `go build ./cmd/goverseer`. Once complete, you
should see a binary in the root of your checkout.

## Development

To run locally during development run `go run ./cmd/goverseer --help`.

To run all tests run `go test -count 1 -v ./...` (the `-count 1` avoids caching
results.

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
