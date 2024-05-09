#!/bin/bash
# Git submodule update
if ! git submodule update --init --recursive; then
  echo "Git submodule update failed"
  exit 1
fi

rm -f unreal_nibi_sdk.dylib

# Build project
if ! go build -o unreal_nibi_sdk.dylib -buildmode=c-shared ./api.go; then
  echo "Project build failed"
  exit 1
fi

# Check build success and test
if ! gcc -o sdk_test sdk_test.c unreal_nibi_sdk.dylib -lpthread; then
  echo "Build check/test executable failed"
  exit 1
fi

./sdk_test

# Setup library for Unreal
if ! install_name_tool -id @rpath/unreal_nibi_sdk.dylib unreal_nibi_sdk.dylib; then
  echo "Library setup for Unreal failed"
  exit 1
fi

echo "All steps completed successfully"
