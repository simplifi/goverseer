# GOverseer

## Accessing GCE Metadata Locally

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
