#!/usr/bin/env bash

for d in ./pkg/* ; do
    cd "$d"
    mockery -all -output=automock -outpkg=automock -case=underscore
    cd ../..
done