#!/usr/bin/env bash

set -e

go vet $(go list ./... | egrep -v "platform")
go vet -v $(go list ./... | egrep -v "platform")
