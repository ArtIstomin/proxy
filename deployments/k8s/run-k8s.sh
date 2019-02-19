#!/bin/bash
parent_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )

# Run Postgres Service
kubectl create -f $parent_path/postgres/activity-config.yaml
kubectl create -f $parent_path/postgres/activity-storage.yaml
kubectl create -f $parent_path/postgres/activity-postgres.yaml

# Run Activity Service
kubectl create -f $parent_path/activity/activity.yaml

#Run Proxy Service
kubectl create -f $parent_path/proxy/proxy.yaml