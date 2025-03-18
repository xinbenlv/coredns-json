#!/bin/bash
set -e

# Get Google Cloud project ID
PROJECT_ID=$(gcloud config get-value project)
echo "Using Google Cloud Project ID: $PROJECT_ID"

# Check command line arguments
REBUILD=false
if [ "$1" == "--rebuild" ]; then
  REBUILD=true
  echo "--rebuild parameter detected, will rebuild base image"
fi

# Check if base image already exists
if [ "$REBUILD" == "true" ] || ! gcloud container images describe gcr.io/$PROJECT_ID/coredns-base:latest > /dev/null 2>&1; then
  echo "Building base image..."
  gcloud builds submit --config=cloudbuild-base.yaml
else
  echo "Base image already exists, automatically skipping build"
fi

# Build JSON plugin image
echo "Building JSON plugin image..."
gcloud builds submit --config=cloudbuild-json.yaml

echo "Build complete!"
echo "Base image: gcr.io/$PROJECT_ID/coredns-base:latest"
echo "JSON plugin image: gcr.io/$PROJECT_ID/coredns-json:latest" 