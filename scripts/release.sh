#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)"

cd "${DIR}/.."
git pull && docker-compose -f deployments/docker-compose.prod.yml up -d --build
cd -
