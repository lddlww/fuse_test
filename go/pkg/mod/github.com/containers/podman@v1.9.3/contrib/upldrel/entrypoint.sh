#!/bin/bash

set -e

source /usr/local/bin/lib_entrypoint.sh

req_env_var GCPJSON_FILEPATH GCPNAME GCPPROJECT BUCKET FROM_FILEPATH TO_FILENAME

[[ -r "$FROM_FILEPATH" ]] || \
    die 2 ERROR Cannot read release archive file: "$FROM_FILEPATH"

[[ -r "$GCPJSON_FILEPATH" ]] || \
    die 3 ERROR Cannot read GCP credentials file: "$GCPJSON_FILEPATH"

echo "Authenticating to google cloud for upload"
gcloud_init "$GCPJSON_FILEPATH"

echo "Uploading archive as $TO_FILENAME"
gsutil cp "$FROM_FILEPATH" "gs://$BUCKET/$TO_FILENAME"
[[ -z "$ALSO_FILENAME" ]] || \
    gsutil cp "$FROM_FILEPATH" "gs://$BUCKET/$ALSO_FILENAME"

echo "."
echo "Release now available for download at:"
echo "    https://storage.googleapis.com/$BUCKET/$TO_FILENAME"
[[ -z "$ALSO_FILENAME" ]] || \
    echo "    https://storage.googleapis.com/$BUCKET/$ALSO_FILENAME"
