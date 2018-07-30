#!/usr/bin/env bash

go test -cover $(go list ./... | grep -v /vendor/)
go test ./cntrpool -coverprofile=coverage.out
go tool cover -html=coverage.out