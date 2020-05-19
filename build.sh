#!/usr/bin/env bash

# start minikube. e.g.
# minikube start --driver=hyperkit

# if any command returns non-zero, quit
set -e

export NAMESPACE=tcp-pool-proxy

echo "About to clean..."
go clean

echo "About to build..."
export GOOS=linux GOARCH=amd64
go build -i

echo "About to package as Docker container..."
eval $(minikube docker-env)
docker build . -t nextmetaphor/tcp-pool-proxy:latest

# happy to continue if the undeploy fails - it may not have been deployed yet...
set +e

echo "Deleting from k8s..."
kubectl delete -f _k8s/tcp-pool-proxy/service.yaml --namespace NAMESPACE 2>/dev/null
kubectl delete -f _k8s/tcp-pool-proxy/deployment.yaml --namespace $NAMESPACE 2>/dev/null

echo "Creating in k8s..."
kubectl create -f _k8s/tcp-pool-proxy/deployment.yaml --namespace $NAMESPACE
kubectl create -f _k8s/tcp-pool-proxy/service.yaml --namespace $NAMESPACE