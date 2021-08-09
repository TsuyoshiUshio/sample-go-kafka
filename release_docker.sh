#!/bin/bash

VERSION=0.6
docker build . -t tsuyoshiushio/kafka-consumer:$VERSION -t tsuyoshiushio/kafka-consumer:latest

docker push tsuyoshiushio/kafka-consumer:$VERSION
docker push tsuyoshiushio/kafka-consumer:latest