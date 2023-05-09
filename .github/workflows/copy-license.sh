#! /bin/bash

# Usage:
# ./.github/workflows/copy-license.sh

# collect licenses in licenses and intel/licenses directories then copy them into a text file

set -e

## collect licenses in licenses directory

LICENSE_FILES=$(find licenses | grep -e LICENSE)

for lic in $LICENSE_FILES; do
    cat $lic >> licenses.txt
done

## collect licenses in intel/licenses directory

INTEL_LICENSE_FILES=$(find intel/licenses | grep -e LICENSE)

for lic in $INTEL_LICENSE_FILES; do
    cat $lic >> licenses.txt
done
