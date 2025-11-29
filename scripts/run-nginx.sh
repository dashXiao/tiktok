#!/bin/bash
# Start nginx to serve docs/demo and reverse-proxy /douyin to tiktok-api.

set -e

# Ensure network exists (ignore if already present)
docker network create tiktok_tiktok >/dev/null 2>&1 || true

docker rm -f tiktok-nginx >/dev/null 2>&1 || true

docker run -d --name tiktok-nginx --network=tiktok_tiktok \
  -p 80:80 \
  -v "$PWD/docs/demo":/usr/share/nginx/html:ro \
  -v "$PWD/nginx.conf":/etc/nginx/conf.d/default.conf:ro \
  nginx:latest

echo "nginx started on port 80 (tiktok-nginx)."
