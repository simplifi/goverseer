# Time Watcher

The Time Watcher allows you to trigger at regular intervals. This is mostly
useful for testing and debugging.

## Configuration

To use the Time Watcher, configure it in your Goverseer config file. The
following configuration option is available:

- `poll_seconds`: This specifies the interval, in seconds, at which the watcher
  will trigger the executioner.

**Example Configuration:**

```yaml
watcher:
  type: time
  config:
    poll_seconds: 60
executioner:
  type: log
```

This configuration would trigger the executioner every 60 seconds.

**Note:**

- The Time Watcher will trigger the executioner regardless of whether any
  changes have occurred in the system.
- Consider the resource consumption of your executioner when choosing a polling
  interval, as frequent executions can impact performance.

This watcher is ideal for testing.
