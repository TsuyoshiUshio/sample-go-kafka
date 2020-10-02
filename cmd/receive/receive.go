package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

func main() {
	fmt.Println("Azure ServiceBus Queue Sender")
	connectionString := os.Getenv("ConnectionString")
	queueName := os.Getenv("queueName")
	fmt.Printf("ConnectionString: '%s'\n", connectionString)
	fmt.Printf("QueueName: '%s'\n", queueName)

	// Create a client to communicate with a Service Bus Namespace
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connectionString))
	if err != nil {
		log.Fatal("Cannot create a client for the Service Bus Namespace", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Create a client to communicate with the queue
	q, err := ns.NewQueue(queueName)
	if err != nil {
		log.Fatal("Cannot create a client for the queue", err)
	}
	err = q.ReceiveOne(
		ctx,
		servicebus.HandlerFunc(func(ctx context.Context, message *servicebus.Message) error {
			fmt.Println(string(message.Data))
			return message.Complete(ctx)
		}))

	if err != nil {
		log.Fatal("Message Receive Failed", err)
	}
}
