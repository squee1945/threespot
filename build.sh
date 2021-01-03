#!/bin/bash

image="${1:-kaiser}"

echo "Building image '${image}'"

pack build "${image}" --path server --builder gcr.io/buildpacks/builder:v1 -e GOOGLE_BUILDABLE=./cmd

echo "Built image '${image}'"
