#!/bin/bash

echo "Building all images"

sudo docker build -t autoscaler ./autoscaler
sudo docker build -t txserver ./txserver
sudo docker build -t webserver ./webserver

echo "Success!"
