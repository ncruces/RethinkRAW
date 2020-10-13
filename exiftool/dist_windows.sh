#!/bin/bash

set -eo pipefail

tgt="../RethinkRAW/utils/exiftool"
exiftool="https://exiftool.org/Image-ExifTool-12.00.tar.gz"
strawberry="http://strawberryperl.com/download/5.32.0.1/strawberry-perl-5.32.0.1-64bit-portable.zip"

cd $(dirname "${BASH_SOURCE[0]}")

# Setup
rm -rf tmp/
mkdir -p tmp/

# Download Exiftool
curl "$exiftool" | tar xz -C tmp/
mv tmp/* tmp/exiftool

# Download Strawberry
curl "$strawberry" --output tmp/strawberry.zip
unzip -qd tmp/strawberry/ tmp/strawberry.zip

# Install modules
pushd tmp/strawberry/perl/bin
./cpanm.bat --notest Win32::FindFile
popd

# Cleanup Strawberry
pushd tmp/strawberry/perl
rm -rf lib/CORE lib/CPAN
find lib -name '.packlist' -delete
find lib -name '*.pod' -delete
find lib -type d -empty -delete
popd

# Copy modules
mkdir -p tmp/exiftool/lib/Win32/
cp -rl tmp/strawberry/perl/bin/perl.exe tmp/exiftool/exiftool.exe
cp -rl tmp/strawberry/perl/bin/*.dll tmp/exiftool
cp -rl tmp/strawberry/perl/bin/ tmp/exiftool
cp -rl tmp/strawberry/perl/lib/ tmp/exiftool
cp -rl tmp/strawberry/perl/vendor/lib/Archive/Zip* tmp/exiftool/lib/Archive
cp -rl tmp/strawberry/perl/vendor/lib/Win32/API* tmp/exiftool/lib/Win32
cp -rl tmp/strawberry/perl/vendor/lib/auto/Win32/API* tmp/exiftool/lib/auto/Win32
cp -rl tmp/strawberry/perl/site/lib/Win32/FindFile* tmp/exiftool/lib/Win32
cp -rl tmp/strawberry/perl/site/lib/auto/Win32/FindFile* tmp/exiftool/lib/auto/Win32

# Cleanup and test
pushd tmp/exiftool
rm -rf html *_files
bin/prove.bat -b t
rm -rf bin t
./exiftool.exe exiftool -ver -v
popd

# Copy to destination
rm -rf "$tgt"
mv tmp/exiftool "$tgt"

# Cleanup
rm -rf tmp/