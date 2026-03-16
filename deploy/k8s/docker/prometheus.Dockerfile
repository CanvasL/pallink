FROM prom/prometheus

COPY deploy/monitoring/prometheus/prometheus.yml /etc/prometheus/prometheus.yml
