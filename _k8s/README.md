```bash

# create namespace for testing
kubectl create -f _k8s/_namespace/namespace.yaml

# add sample backend API
kubectl create -f _k8s/sample-api/sample-api-deployment.yaml
kubectl create -f _k8s/sample-api/sample-api-service.yaml

# add redis in-memory cache for backend service pool
kubectl create -f _k8s/redis/redis-deployment.yaml
kubectl create -f _k8s/redis/redis-service.yaml

# add influx for monitoring
kubectl create configmap influxdb-conf --namespace=aws-container-factory --from-file=_k8s/influxdb/influxdb.conf
kubectl create -f _k8s/influxdb/influxdb-deployment.yaml
kubectl create -f _k8s/influxdb/influxdb-service.yaml

# create a database
curl -X POST -u admin:admin "http://`minikube ip`:30100/query" --data-urlencode "q=CREATE DATABASE \"aws-container-factory\""

# add grafana for monitoring dashboards
kubectl create -f _k8s/grafana/grafana-deployment.yaml
kubectl create -f _k8s/grafana/grafana-service.yaml

# add the data sources
curl -X POST -u admin:admin "http://`minikube ip`:30103/api/datasources" -d @_k8s/grafana/data-source/aws-container-factory.json --header "Content-Type: application/json"
curl -X POST -u admin:admin "http://`minikube ip`:30103/api/datasources" -d @_k8s/grafana/data-source/gatling.json --header "Content-Type: application/json"

# log into the grafana dashboard (admin:admin)
open http://`minikube ip`:30103
```
