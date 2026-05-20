#!/bin/bash

# Start Prometheus binding to all network interfaces
/etc/prometheus/prometheus --config.file=/etc/prometheus/prometheus.yml --storage.tsdb.path=/etc/prometheus/data

# Wait to keep the container running
wait -n