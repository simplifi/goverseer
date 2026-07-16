# GCP Secrets Manager Watcher

The GCP Secrets Manager Watcher allows you to monitor a secret stored in Google
Cloud Secrets Manager for changes. When a change in the secret's payload is
detected, Goverseer can trigger an executioner to take action based on the
updated secret value. The changed secret value is passed to the executioner for
processing.

## Configuration

To use the GCP Secrets Manager Watcher, you need to configure it in your
Goverseer config file. The following configuration options are available under
the `config` section of your watcher definition:

- `project_id`: (Required) The GCP project that will be monitored for a change
  in its Secrets Manager. Must be a non-empty string.
- `secret_name`: (Required) The name of the secret to watch within the specified
  project (e.g., `nomad-license-key`). Must be a non-empty string.
- `secrets_file_path`: (Required) The path for the file that needs to be updated
  when a secret changes. Must be a non-empty string.
- `credentials_file`: (Optional) Path for the credentials file if needing to
  test locally or use a service account's credentials instead of the ADC
  approach assumed.
- `check_interval_seconds`: (Optional) The interval in seconds at which the
  watcher will poll the Secret Manager for changes. Must be greater than or
  equal to `1`. Defaults to `60` seconds.
- `secret_error_wait_seconds`: (Optional) The number of seconds to wait before
  retrying after a failed attempt to access the secret. Must be greater than or
  equal to `1`. Defaults to `5` seconds.

**Example Configuration:**

This is a sample configuration for watching changes to the Nomad license key in
a GCP Project's Secret Manager. With Goverseer running on the Nomad client, the
license key on that file would be updated whenever there is a change in Secret
Manager.

```yaml
name: nomad-license-watcher-dev
watcher:
  type: gcp_secrets
  config:
    project_id: "nomad-dev-2f03"
    secret_name: "nomad-license-key"
    secrets_file_path: "/etc/nomad.d/nomad.hclic"
    check_interval_seconds: 5
executioner:
  type: shell
  config:
    shell: /bin/bash -lec
    command: |
      NEW_LICENSE_KEY=$(cat "${GOVERSEER_DATA}")
      LICENSE_FILE="/tmp/test_license_file.txt"
      echo "Writing new Nomad license key to $LICENSE_FILE"
      echo "$NEW_LICENSE_KEY" | sudo tee "$LICENSE_FILE"
      echo "Restarting Nomad service..."
      sudo systemctl restart nomad
      echo "Nomad service restarted."
```
