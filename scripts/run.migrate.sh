#!/bin/bash
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
DOCKER_IMAGE_REPO="migrate/migrate:v4.8.0"
MIGRATION_DIR="${SCRIPT_DIR}/../store/mysql/migrations"

RUN_CREATE=false
RUN_MIGRATIONS=false
FORCE_MIGRATION=false
SHOW_VERSION=false
RUN_CREATE_FILE_NAME="placeholder"
RUN_DOWN=false

DB_PORT=${DB_PORT:-3306}
DB_HOST=${DB_HOST:-localhost}
DB_USER=${DB_USER:-root}
DB_NAME=${DB_NAME:-go_sample_api_server_structure}
DB_PASS=${DB_PASS:-12345}
DB_CONNECTION_URL="mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"

function show_help {
    echo "Usage: $0 [-h] [-m] [-c <string>] [-d]"
    echo "    -h            Display this help message."
    echo "    -m            Run all the migrations possible."
    echo "    -f <string>   Force a specific version."
    echo "    -c <string>   Creates new migration file in the migrations directory."
    echo "    -v            Print migration version."
    echo "    -d            Run all down migrations possible."
}

while getopts "h?mc:f:vd" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    m)  RUN_MIGRATIONS=true
        ;;
    c)  RUN_CREATE=true
        RUN_CREATE_FILE_NAME=${OPTARG}
        ;;
    f)  FORCE_MIGRATION=true
        VERSION=${OPTARG}
        ;;
    v) SHOW_VERSION=true
        ;;
    d)  RUN_DOWN=true
        ;;
    esac
done
shift $((OPTIND-1))

if [ ${RUN_MIGRATIONS} = true ]; then

  # run the migrations
  docker run --rm \
    --user $(id -u):$(id -g) \
    --mount type=bind,source="${MIGRATION_DIR}",target="/migrations",readonly \
    --network host \
    ${DOCKER_IMAGE_REPO} \
    -path=/migrations/ -database=${DB_CONNECTION_URL} -verbose up

elif [ ${SHOW_VERSION} = true ]; then

    docker run --rm \
      --user $(id -u):$(id -g) \
      --mount type=bind,source="${MIGRATION_DIR}",target="/migrations",readonly \
      --network host \
      ${DOCKER_IMAGE_REPO} \
      -path=/migrations/ -database=${DB_CONNECTION_URL} -verbose version

elif [ ${FORCE_MIGRATION} = true ]; then

    docker run --rm \
      --user $(id -u):$(id -g) \
      --mount type=bind,source="${MIGRATION_DIR}",target="/migrations",readonly \
      --network host \
      ${DOCKER_IMAGE_REPO} \
      -path=/migrations/ -database=${DB_CONNECTION_URL} -verbose force ${VERSION}

elif [ ${RUN_CREATE} = true ]; then

    docker run --rm \
      --user $(id -u):$(id -g) \
      --mount type=bind,source="${MIGRATION_DIR}",target="/migrations" \
      --network=host \
      ${DOCKER_IMAGE_REPO} \
      -path=/migrations/ -verbose create -ext=sql -dir=/migrations/ ${RUN_CREATE_FILE_NAME}

elif [ ${RUN_DOWN} = true ]; then

  docker run --rm \
    --user $(id -u):$(id -g) \
    --mount type=bind,source="${MIGRATION_DIR}",target="/migrations",readonly \
    --network host \
    ${DOCKER_IMAGE_REPO} \
    -path=/migrations/ -database=${DB_CONNECTION_URL} -verbose down

else
    show_help
fi