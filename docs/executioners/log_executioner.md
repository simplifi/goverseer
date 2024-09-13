# Log Executioner

The Log Executioner provides a simple way to log messages when Goverseer detects
a change. This is useful for debugging and monitoring purposes, allowing you to
track changes and their associated actions.

## Configuration

To use the Log Executioner, configure it in your Goverseer config file. The
following configuration option is available:

- `tag`: (Optional) This allows you to specify a custom tag that will be added
  to the log message. This can be helpful for filtering and searching logs.
  Defaults to no tag.

**Example Configuration:**

```yaml
executioner:
  type: log
  config:
    tag: my-application
```

With this configuration, every time the watcher detects a change, the log
executioner will output a message like:

```log
Sep 12 16:12:58.095 INF received data data=<content of the change>
```

**Note:**

- The Log Executioner simply logs the change detected by the watcher; it does
  not perform any other actions.
- For more complex actions, consider using the `shell_executioner` or creating a
  custom executioner.

This executioner is particularly useful during development and testing, allowing
you to quickly verify that Goverseer is functioning as expected and that changes
are being detected. It can also be helpful for debugging issues and
understanding the behavior of your application in response to changes.
