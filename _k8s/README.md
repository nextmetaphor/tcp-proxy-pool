```bash

# create namespace for testing
kubectl create -f _k8s/_namespace/namespace.yaml

# add sample backend API
kubectl create -f _k8s/sample-api/deployment.yaml
kubectl create -f _k8s/sample-api/service.yaml

# add redis in-memory cache for backend service pool
kubectl create -f _k8s/redis/deployment.yaml
kubectl create -f _k8s/redis/service.yaml

# add influx for monitoring
kubectl create configmap influxdb-conf --namespace=tcp-proxy-pool --from-file=_k8s/influxdb/influxdb.conf
kubectl create -f _k8s/influxdb/deployment.yaml
kubectl create -f _k8s/influxdb/service.yaml

# create a database
curl -X POST -u admin:admin "http://`minikube ip`:30100/query" --data-urlencode "q=CREATE DATABASE \"tcp-proxy-pool\""

# add grafana for monitoring dashboards
kubectl create -f _k8s/grafana/deployment.yaml
kubectl create -f _k8s/grafana/service.yaml

# add the data sources
curl -X POST -u admin:admin "http://`minikube ip`:30103/api/datasources" -d @_k8s/grafana/data-source/tcp-proxy-pool.json --header "Content-Type: application/json"
curl -X POST -u admin:admin "http://`minikube ip`:30103/api/datasources" -d @_k8s/grafana/data-source/gatling.json --header "Content-Type: application/json"

# log into the grafana dashboard (admin:admin)
open http://`minikube ip`:30103
```
