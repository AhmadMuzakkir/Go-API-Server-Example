#!/bin/bash

TEST=false

while getopts ":t" opt; do
  case $opt in
  t)
    TEST=true
    ;;
  \?) ;;
  esac
done

if [ ${TEST} = true ]; then
  # Start mysql database for the mysql unit tests
  docker run --rm -it \
    --name "go-sample-api-server-structure" \
    --init \
    -p 3307:3306 \
    --env="MYSQL_ROOT_PASSWORD=test_password" \
    --env="MYSQL_DATABASE=test_database" \
    mysql:8 \
    bash -c "/entrypoint.sh mysqld & bash"
else
  docker run --rm -it \
    --name "go-sample-api-server-structure" \
    --init \
    --mount source=go_sample_api_server_structure_mysql,target=/var/lib/mysql \
    -p 3306:3306 \
    --env="MYSQL_ROOT_PASSWORD=12345" \
    --env="MYSQL_DATABASE=go_sample_api_server_structure" \
    --security-opt="seccomp=unconfined" \
    mysql:8 \
    bash -c "/entrypoint.sh mysqld & bash"
fi
