# Goverseer

![Goverseer](assets/goverseer.png)

Goverseer is a tool for monitoring some kind of upstream, and taking actions
when a change is detected. In some ways it is similar to consul-template, except
Goverseer does not handle templating. Instead it is focused on watching for
changes and taking some kind of action based on those changes.

## Usage

Goverseer is configured using a yaml config file. The `goverseer start` command
takes a path to the config file. If a config path is not provided,
`/etc/goverseer.yaml` will be used.

```yaml
# Example goverseer config
---
name: example

watcher:
  type: time
  config:
    poll_seconds: 1

executioner:
  type: log
  config:
    tag: example
```

The available values for `watcher.type` are:

- `file`: [File Watcher](docs/watchers/file_watcher.md)
- `gce_metadata`: [GCE Metadata Watcher](docs/watchers/gce_metadata_watcher.md)
- `time`: [Time Watcher](docs/watchers/time_watcher.md)

The available values for `executioner.type` are:

- `log`: [Log Executioner](docs/executioners/log_executioner.md)
- `shell`: [Shell Executioner](docs/executioners/shell_executioner.md)

The configuration options for `watcher.config` and `executioner.config` are
determined by the selected type. See the documentation for the specific watcher
and executioner type for more details on available configuration options.

## Building

To build Goverseer, simply run `make build`. Once complete, you should see a
binary in the root of your checkout.

## Development

To run locally during development, run `go run ./cmd/goverseer --help`.

To run all tests, run `make test`.
