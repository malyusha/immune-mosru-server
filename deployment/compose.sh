#!/usr/bin/env bash

env=local
if [ -n "$IMMUNE_ENV" ]; then
  env="$IMMUNE_ENV"
fi

docker-compose -f docker-compose.yml -f "docker-compose-${env}.yml" "$@"
