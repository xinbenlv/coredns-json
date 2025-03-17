#!/bin/bash
set -e

# Check if Docker Hub password is provided
if [ -z "$1" ]; then
  echo "No DockerHub password provided. Will only push to Google Container Registry."
  echo "Triggering Google Cloud Build..."
  gcloud builds submit --config=cloudbuild.yaml
else
  DOCKERHUB_PASSWORD="$1"
  echo "DockerHub password provided. Will push to both Google Container Registry and DockerHub."
  echo "Triggering Google Cloud Build with DockerHub password as environment variable..."
  gcloud builds submit --config=dockerhub-build.yaml --substitutions=_DOCKERHUB_PASSWORD="$DOCKERHUB_PASSWORD"
fi

echo "Build triggered. Check Google Cloud Build console for progress." 