#!/bin/bash

VERSION=0.3
docker build . -t tsuyoshiushio/kafka-consumer:$VERSION

docker push tsuyoshiushio/kafka-consumer:$VERSION
docker push tsuyoshiushio/kafka-consumer:latest