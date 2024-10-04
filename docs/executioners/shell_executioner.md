# Shell Executioner

The Shell Executioner allows you to execute arbitrary shell commands when
Goverseer detects a change. This is useful for triggering actions like
restarting services, sending notifications, or running custom scripts. The data
from the watcher that triggered the execution is stored in a file in the
executioner's work directory. The path to this data file is passed to the shell
command via an environment variable named `GOVERSEER_DATA`.

## Configuration

To use the Shell Executioner, you need to configure it in your Goverseer config
file. The following configuration options are available:

- `command`: This is the shell command you want to execute.For example,
  `echo "Data received: $GOVERSEER_DATA"`.
- `shell`: (Optional) This specifies the shell to use for executing the command.
  Defaults to `/bin/sh -ec` if not provided.
- `work_dir`: (Optional) This specifies the directory where the executioner
  stores the data file. Defaults to the `/tmp` if not provided.
- `persist_data`: (Optional) This determines whether the command and data will
  persist after completion. This can be useful to enable when troubleshooting
  configured commands but should generally remain disabled otherwise. Defaults
  to `false` if not provided.

**Example Configuration:**

```yaml
name: shell_executioner_example
watcher:
  type: time
  config:
    poll_seconds: 60
executioner:
  type: shell
  config:
    shell: /bin/bash -euo pipefail -c
    command: |
      echo "Data received: $GOVERSEER_DATA"
```

**Note:**

- Ensure that the specified shell is available on your system.
- The Shell Executioner writes the `GOVERSEER_DATA` content to a temporary file
  and sets the `GOVERSEER_DATA` environment variable to the path of this file.

This executioner is particularly useful for automating tasks that require shell
command execution based on dynamic data changes. For example, you can use it to
restart a service whenever a configuration file is updated, or to send
notifications when specific events occur.
