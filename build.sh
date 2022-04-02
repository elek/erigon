#!/usr/bin/env bash
TAG=$(git log -n 1 --pretty=format:"%h")
docker build -t ghcr.io/elek/erigon:$TAG .
docker push ghcr.io/elek/erigon:$TAG
