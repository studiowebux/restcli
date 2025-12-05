#!/bin/bash

mkdir -p lgtm/{grafana,loki,prometheus}

podman run --name lgtm -p 3000:3000 -p 4317:4317 -p 4318:4318 -d \
    -v ./lgtm/grafana:/data/grafana \
    -v ./lgtm/prometheus:/data/prometheus \
    -v ./lgtm/loki:/data/loki \
    -e GF_PATHS_DATA=/data/grafana \
    docker.io/grafana/otel-lgtm:0.8.1

OTEL_DENO=true deno run -A streaming.ts
