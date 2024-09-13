# File Watcher

The File Watcher allows you to trigger an action when a file is changed. When a
change is detected, Goverseer can trigger an executioner to take action. The
path to the changed file is passed to the executioner for processing.

## Configuration

To use the File Watcher, you need to configure it in your Goverseer config file.
The following configuration options are available:

- `path`: This is the path to the file that should be monitored for changes
- `poll_seconds`: (Optional) This specifies the frequency in seconds for
  checking if the file has been modified. Defaults to `5` if not provided.

**Example Configuration:**

```yaml
watcher:
  type: file
  config:
    path: /path/to/file
    poll_seconds: 10
executioner:
  type: log
```

This configuration would check the file located at `/path/to/file` for a changed
timestamp every 10 seconds. If the file has been modified, it would trigger the
log executioner.

**Note:**

- The File Watcher will only trigger the executioner if the modification
  timestamp of the file being monitored has been updated since the last check.
- Consider the resource consumption when choosing a polling interval, as
  frequent checks can impact performance.
