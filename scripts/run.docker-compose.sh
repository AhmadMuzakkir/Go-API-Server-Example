#!/bin/bash
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

cd "${SCRIPT_DIR}"/..
go mod vendor
cd -

docker-compose up --build