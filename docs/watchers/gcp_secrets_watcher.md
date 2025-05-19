# GCP Secrets Manager Watcher

The GCP Secrets Manager Watcher allows you to monitor a secret stored in Google Cloud Secrets Manager for changes. When a change in the secret's payload is detected, Goverseer can trigger an executioner to take action based on the updated secret value. The changed secret value (as a map of project ID to the secret payload) is passed to the executioner for processing.

## Configuration

To use the GCP Secrets Manager Watcher, you need to configure it in your Goverseer config file. The following configuration options are available under the `config` section of your watcher definition:

- `projects`: (Required) A list of GCP project IDs where the secret is stored (e.g., `["nomad-dev", "nomad-prd"]`). The watcher will monitor the secrets in all listed projects.
- `secret_name`: (Required) The name of the secret to watch within each of the specified projects (e.g., `nomad-license-key`).
- `credentials_file`: (Optional) Path for the credentials file if needing to test locally or use a service account's credentials instead of the ADC approach assumed.
- `check_interval_seconds`: (Optional) The interval in seconds at which the watcher will poll the Secret Manager for changes. Defaults to `60` seconds.
- `secret_error_wait_seconds`: (Optional) The number of seconds to wait before retrying after a failed attempt to access the secret. Defaults to `5` seconds.

**Example Configuration:**

```yaml
name: gcp_secrets_watcher_example
watcher:
  type: gcp_secrets
  config:
    projects: ["nomad-dev", "nomad-prd"]
    secret_name: "nomad-license-key"
    check_interval_seconds: 60
executioner:
  type: shell
  config:
    shell: /bin/bash -lec
    command: 


```

## Development

To test the Goverseer setup for this watcher locally:
1. Install the necessary library (`go get cloud.google.com/go/secretmanager/apiv1`).
2. Use `gcloud auth application-default login` to provide authentication for Google Auth Library, or use a credential file for a service account that has permissions to access secret versions. Ensure that whatever is used has the secretmanager.secrets.get and secretmanager.versions.access for the projects in which you will be operating.
2. Replace placeholders for `projects` and `secret_name` at a minimum.
3. Run the watcher via `go run gcp_secrets_watcher.go`.