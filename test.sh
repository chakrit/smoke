#!/bin/sh

set -e

go install -v .
$(go env GOPATH)/bin/smoke "$@" tests.yml
