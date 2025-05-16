# GCP Secrets Manager Watcher

The GCP Secrets Manager Watcher allows you to monitor a secret stored in Google Cloud Secrets Manager for changes. When a change in the secret's payload is detected, Goverseer can trigger an executioner to take action based on the updated secret value. The changed secret value (as a map of project ID to the secret payload) is passed to the executioner for processing.

## Configuration

To use the GCP Secrets Manager Watcher, you need to configure it in your Goverseer config file. The following configuration options are available under the `config` section of your watcher definition:

- `projects`: (Required) A list of GCP project IDs where the secret is stored (e.g., `["nomad-dev", "nomad-prd"]`). The watcher will monitor the secret in all listed projects.
- `secret_name`: (Required) The name of the secret to watch within each of the specified projects (e.g., `nomad-license-key`).
- `credential_file`: (Optional) The path to a Google Cloud service account credential file. If provided, the watcher will use these credentials to access Secret Manager. If not provided, it will rely on Google Application Default Credentials (ADC).
- `check_interval_seconds`: (Optional) The interval in seconds at which the watcher will poll the Secret Manager for changes. Defaults to `60` seconds.
- `secret_error_wait_seconds`: (Optional) The number of seconds to wait before retrying after a failed attempt to access the secret. Defaults to `5` seconds.
- `state_file_path`: (Optional) The path to a file where the watcher can store the last known secret values. This helps the watcher detect changes across restarts. Defaults to `last_known_licenses.json` in the current working directory.

**Example Configuration:**

```yaml
name: gcp_secrets_watcher_example
watcher:
  type: gcp_secrets
  config:
    projects: ["nomad-dev", "nomad-prd"]
    secret_name: nomad-license-key
    credential_file: /path/to/your/gcp-credentials.json
    check_interval_seconds: 300
    state_file_path: /var/goverseer/state/nomad_license_state.json
executioner:
  type: shell # Or your desired executioner type
```

## Development

To test the Goverseer setup for this watcher locally, 