#!/bin/bash
#
# Build script to create addlicense container.
# Example of building with specific tag: TAG="something" ./build_addlicense

tag=${TAG:="latest"}
echo "Building addlicense container with tag: $tag"
docker build -t nokia/addlicense-nokia:$tag . || echo "Build failed."
