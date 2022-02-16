#!/bin/bash

echo "Removing all containers and images"

docker rm $(docker ps -a -q)
docker-compose down --remove-orphans
docker rmi $(docker image ls -a)

echo "Done."