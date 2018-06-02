```bash

export NAMESPACE=tcp-proxy-pool

# create namespace for testing
kubectl create namespace $NAMESPACE

# add sample backend API
kubectl create -f _k8s/sample-api/deployment.yaml --namespace=$NAMESPACE
kubectl create -f _k8s/sample-api/service.yaml --namespace=$NAMESPACE

# add influx for monitoring
kubectl create configmap influxdb-conf --from-file=_k8s/influxdb/influxdb.conf --namespace=$NAMESPACE
kubectl create -f _k8s/influxdb/deployment.yaml --namespace=$NAMESPACE
kubectl create -f _k8s/influxdb/service.yaml --namespace=$NAMESPACE

# create a database
curl -X POST -u admin:admin "http://`minikube ip`:30100/query" --data-urlencode "q=CREATE DATABASE \"tcp-proxy-pool\""

# add grafana for monitoring dashboards
kubectl create -f _k8s/grafana/deployment.yaml --namespace=$NAMESPACE
kubectl create -f _k8s/grafana/service.yaml --namespace=$NAMESPACE

# add the data sources
curl -X POST -u admin:admin "http://`minikube ip`:30103/api/datasources" -d @_k8s/grafana/data-source/tcp-proxy-pool.json --header "Content-Type: application/json"
curl -X POST -u admin:admin "http://`minikube ip`:30103/api/datasources" -d @_k8s/grafana/data-source/gatling.json --header "Content-Type: application/json"

# log into the grafana dashboard (admin:admin)
open http://`minikube ip`:30103
```
