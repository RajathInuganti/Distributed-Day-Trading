#!/bin/bash

echo "Building all images"

sudo docker build -t autoscaler ./autoscaler

echo "Success!"