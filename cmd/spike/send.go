package main

import (
	"bufio"
	"context"
	"errors"
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

	// qm := ns.NewQueueManager()
	// qe, err := qm.Put(ctx, "abc", servicebus.QueueEntityWithDeadLetteringOnMessageExpiration()) // Doesn't work
	// if err != nil {
	// 	fmt.Println("error :", err.Error())
	// } else {
	// 	fmt.Println("endity anme: ", qe.Name)
	// }

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

	go executeCountLoop(ns, queueName)
	go executeReceiveLoop(ns, queueName)
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	time.Sleep(time.Minute * 30) // This is for debugger
}

func executeCountLoop(ns *servicebus.Namespace, queueName string) {
	q, err := ns.NewQueue(queueName)
	if err != nil {
		log.Fatal("Cannot create a Count client for the queue", err)
	}
	m := ns.NewQueueManager()
	ctx2 := context.Background()

	for {
		entity, _ := m.Get(ctx2, queueName)
		fmt.Println("ActiveMessageCount: ", *entity.CountDetails.ActiveMessageCount)
		ctx := context.Background()
		iterator, _ := q.Peek(ctx)
		var c int
		for {
			item, _ := iterator.Next(ctx)
			if item != nil && item.LockToken != nil {
				fmt.Printf("lockToken: %s \n", item.LockToken.String())
			} else {
				if item != nil {
					//	fmt.Println("lockToken: nil")
					c++
				} else {
					fmt.Println("item is nil")
				}
			}
			if item != nil {
				// body, _ := json.Marshal(item)
				// fmt.Println(string(body))
			}

			if item == nil {
				iterator.Done()
				break
			}
		}
		fmt.Println("Count: ", c)
		time.Sleep(time.Second * 2)
	}
}

func executeReceiveLoop(ns *servicebus.Namespace, queueName string) {
	// q, err := ns.NewQueue(queueName, servicebus.QueueWithReceiveAndDelete())
	q, err := ns.NewQueue(queueName)
	if err != nil {
		log.Fatal("Cannot create a Receive client for the queue", err)
	}
	ctx := context.Background()

	for {
		err = q.ReceiveOne(
			ctx,
			servicebus.HandlerFunc(func(ctx context.Context, message *servicebus.Message) error {

				//				fmt.Println(message.LockToken.String())  // It will be null when the ReceiveAndDelete mode
				fmt.Println(string(message.Data))
				fmt.Println("Waiting...")
				time.Sleep(time.Second * 10)
				fmt.Println("Done.")
				m := make(map[string]string)
				m["rootCause"] = "hagechabin"
				err := message.DeadLetterWithInfo(ctx, errors.New("go to deadletter"), servicebus.ErrorInternalError, m) // Doesn't work for ReceiveAndDelete

				return err
				//return message.Abandon(ctx) // test of dead letter queue
				// return message.Complete(ctx)
			}))
		if err != nil {
			fmt.Printf("Receive Error: %v", err)

			break
		}
	}
}
