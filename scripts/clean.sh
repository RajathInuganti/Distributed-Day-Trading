#!/bin/bash

echo "Removing all containers and images"

sudo docker kill $(sudo docker ps -a -q)
sudo docker-compose down --remove-orphans
sudo docker rmi $(sudo docker image ls -a)

echo "Done."