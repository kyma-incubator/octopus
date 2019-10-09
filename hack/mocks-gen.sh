#!/usr/bin/env bash

cd ./pkg/scheduler
mockery -name=StatusProvider -output=automock -outpkg=automock -case=underscore