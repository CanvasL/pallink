FROM grafana/grafana

COPY deploy/monitoring/grafana/provisioning /etc/grafana/provisioning
COPY deploy/monitoring/grafana/dashboards /etc/grafana/dashboards
