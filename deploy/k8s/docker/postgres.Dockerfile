FROM postgres:15

COPY deploy/docker/init.sql /docker-entrypoint-initdb.d/init.sql
