# GCP Secrets Manager Watcher

The GCP Secrets Manager Watcher allows you to monitor a secret stored in Google Cloud Secrets Manager for changes. When a change in the secret's payload is detected, Goverseer can trigger an executioner to take action based on the updated secret value. The changed secret value (as a map of project ID to the secret payload) is passed to the executioner for processing.

## Configuration

To use the GCP Secrets Manager Watcher, you need to configure it in your Goverseer config file. The following configuration options are available under the `config` section of your watcher definition:

- `project_id`: (Required) The GCP project that will be monitored for a change in its secrets manager.
- `secret_name`: (Required) The name of the secret to watch within each of the specified projects (e.g., `nomad-license-key`).
- `secrets_file_path`: (Required) The path for the file that needs to be updated when a secret changes.
- `credentials_file`: (Optional) Path for the credentials file if needing to test locally or use a service account's credentials instead of the ADC approach assumed.
- `check_interval_seconds`: (Optional) The interval in seconds at which the watcher will poll the Secret Manager for changes. Defaults to `60` seconds.
- `secret_error_wait_seconds`: (Optional) The number of seconds to wait before retrying after a failed attempt to access the secret. Defaults to `5` seconds.

**Example Configuration:**

```yaml
name: gcp_secrets_watcher_example
watcher:
  type: gcp_secrets
  config:
    project_id: "nomad-dev-2f03"
    secret_name: "nomad-license-key"
    secrets_file_path: "/etc/nomad.d/nomad.hclic"
executioner:
  type: shell
  config:
    shell: /bin/bash -euo pipefail -c
    command: |
      #!/bin/bash
      NEW_LICENSE_KEY=$(cat "${GOVERSEER_DATA}")
      LICENSE_FILE="{{ .SecretsFilePath }}"
      echo "Writing new Nomad license key to $LICENSE_FILE"
      echo "$NEW_LICENSE_KEY" | sudo tee "$LICENSE_FILE"
      echo "Restarting Nomad service..."
      sudo systemctl restart nomad
      echo "Nomad service restarted."
```
