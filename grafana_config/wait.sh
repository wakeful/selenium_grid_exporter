#!/bin/sh

while ! curl --output /dev/null --silent --head --fail http://grafana:3000; do sleep 1 && echo -n .; done;

curl -v -H "Content-Type: application/json" -X POST -d @prometheus_datasource.json http://grafana:3000/api/datasources

curl -v -H "Content-Type: application/json" -X POST -d @selenium_grid_dashboard.json http://grafana:3000/api/dashboards/db
