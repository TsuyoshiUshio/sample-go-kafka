package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

func main() {
	fmt.Println("Azure ServiceBus Queue Sender")
	connectionString := os.Getenv("ConnectionString")
	queueName := os.Getenv("queueName")
	if len(os.Args) != 2 {
		log.Fatalf("Specify the counter parameter. e.g. send 100 Parameter length: %d\n", len(os.Args))
	}
	count, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("count should be integer : %s", os.Args[1])
	}
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
	for i := 0; i < count; i++ {
		err = q.Send(ctx, servicebus.NewMessageFromString("Hello!"))
		fmt.Printf("Send Hello %d \n", i)
		if err != nil {
			log.Fatal(err)
		}
	}
}
