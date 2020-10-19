#!/bin/bash

set -eo pipefail

tgt="../RethinkRAW.app/Contents/MacOS/utils/exiftool"
exiftool="https://exiftool.org/Image-ExifTool-12.00.tar.gz"


cd $(dirname "${BASH_SOURCE[0]}")

# Setup
rm -rf tmp/
mkdir -p tmp/

# Download Exiftool
curl "$exiftool" | tar xz -C tmp/
mv tmp/* tmp/exiftool

# Cleanup and test
pushd tmp/exiftool
rm -rf html *_files
prove -l lib -b t 
rm -rf t
./exiftool -ver -v
popd

# Move to destination
rm -rf "$tgt"
mv tmp/exiftool "$tgt"

# Cleanup
rm -rf tmp/