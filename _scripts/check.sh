#!/usr/bin/env bash

echo "### go vet ###"
go vet $(go list ./... | grep -v /vendor/)

echo "### golint ###"
golint $(go list ./... | grep -v /vendor/)