# Infrastructure Structure

- `docker/` contains local development compose files and observability provisioning
- `docker/grafana/provisioning` contains datasource and dashboard provisioning
- `docker/grafana/dashboards` stores versioned dashboard JSON files
- `postgres/init` is reserved for optional bootstrap SQL scripts
- `redis/` is reserved for local Redis configuration when needed
