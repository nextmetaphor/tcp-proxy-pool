#!/usr/bin/env bash

export NAMESPACE=tcp-pool-proxy

# deployments
kubectl delete -f _k8s/telegraf/telegraf-deployment.yaml --namespace=${NAMESPACE}
kubectl delete -f _k8s/influxdb/deployment.yaml --namespace ${NAMESPACE}
kubectl delete -f _k8s/sample-api/deployment.yaml --namespace ${NAMESPACE}

# services
kubectl delete -f _k8s/telegraf/telegraf-service.yaml --namespace=${NAMESPACE}
kubectl delete -f _k8s/grafana/service.yaml --namespace ${NAMESPACE}
kubectl delete -f _k8s/influxdb/service.yaml --namespace ${NAMESPACE}
kubectl delete -f _k8s/sample-api/service.yaml --namespace ${NAMESPACE}

# configmaps
kubectl delete configmap telegraf-conf --namespace=${NAMESPACE}
kubectl delete configmap influxdb-conf --namespace ${NAMESPACE}

# namespaces
kubectl delete namespace ${NAMESPACE}

