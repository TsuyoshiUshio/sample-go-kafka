# KEDA Kafka Scaled Job Samples

Sample program for sending/receiving message to Kafka Broker and KEDA v2 Kafka Trigger

# How to run the sample

## Prerequisite
* Install Go lang (1.16+)
* Docker

## Set Enviornment Variables for the container

The current sample is set the broker as EventHubs. EventHubs using `username` and `password` as the authentication.

| Key | Description |
| ---- | ------ |
| SASL_PASSWORD | The Password of Kafka Broker (In case of EventHub Kafka API, it is connection string) |
| BROKER_LIST | The url of the broker |

## Run queue receiver/sender

### Prerequiste 

Set the environment described in the Docker section. `SASL_PASSWORD` and `BROKER_LIST`. 

### receiver

Receive queue messages.

```bash
$ cd cmd/receive
$ go run receive.go
```

### sender

Send 100 messages to the Azure Storage Queue.

```bash
$ cd cmd/send
$ go run send.go -n 100 -m "*** hello world ***" // -n the number of sending the message. -m a message to send
```

# How to debug

To see the behavior, you can debug it. This is the sample of the VSCode `.vscode/launch.json`. Then Start Debugging. 
It requires, VSCode [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go).

```json
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}",
            "env": {
                "SASL_PASSWORD": "YOUR_SASL_PASSWORD_HERE",
                "BROKER_LIST":"YOUR_BROKER_URL_HERE"
            },
            "args": []
        }
    ]
}
```

## Run with KEDA

### Prerequisite
KEDA v2 is deployed already. 
You have a kubernetes cluster and configured with kubectl. 

### Start the ScaledJob

#### Copy secret.yaml.example to secret.yaml
Modify the following section. 

```
  sasl: "plaintext as base64"
  username: "$ConnectionString as base64"
  password: "YOUR_EVENT_HUBS_CONNECTION_STRING_BASE64"
  tls: "enable as base64"
```

Then Apply it. This will create a secret for the ScaledJob.

```
$ kubectl apply -f secret.yaml
$ kubectl apply -f scaledjob.yaml
```

Send queue to the target queue. You can do it with the `send` command. 

