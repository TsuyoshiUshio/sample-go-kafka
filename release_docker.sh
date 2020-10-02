#!/bin/bash

docker build . -t tsuyoshiushio/azure-servicebus-queue-client:dev

docker push tsuyoshiushio/azure-servicebus-queue-client:dev