package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

var (
	brokers  = ""
	version  = ""
	group    = ""
	topics   = ""
	assignor = ""
	oldest   = true
	verbose  = false
	consumed = make(chan bool)
)

const (
	saslPassword     = "SASL_PASSWORD"
	brokerList       = "BROKER_LIST"
	topicsEnv        = "TOPICS"
	consumerGroupEnv = "CONSUMER_GROUP"
)

func init() {
	flag.StringVar(&brokers, "brokers", "", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&group, "group", "$Default", "Kafka consumer group definition")
	flag.StringVar(&topics, "topics", "workitems", "Kafka topics to be consumed, as a comma separated list")
	flag.StringVar(&assignor, "assignor", "range", "Consumer group partition assignment strategy (range, roundrobin, sticky)")
	flag.BoolVar(&oldest, "oldest", true, "Kafka consumer consume initial offset from oldest")
	flag.BoolVar(&verbose, "verbose", true, "Sarama logging")
}

func main() {
	log.Println("Starting a new Sarama consumer")

	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	//	version, err := sarama.ParseKafkaVersion(version)
	//	if err != nil {
	//		log.Panicf("Error parsing Kafka version: %v", err)
	//	}

	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		log.Panicf("Unrecognized consumer group partition assginor: %s", assignor)
	}
	config.Net.SASL.Enable = true
	config.Net.SASL.User = "$ConnectionString"
	config.Net.SASL.Password = os.Getenv(saslPassword)
	config.Consumer.Group.Session.Timeout = 60 * time.Second
	if len(brokers) == 0 {
		brokers = os.Getenv(brokerList)
		if len(brokers) == 0 {
			log.Panicf("BrokerList is empty. Set brokers as environment variables: %s or -brokers option", brokerList)
		}
	}

	topicsEnvValue := os.Getenv(topicsEnv)
	if len(topicsEnvValue) > 0 {
		topics = topicsEnvValue
	}

	consumerGroupEnvValue := os.Getenv(consumerGroupEnv)
	if len(consumerGroupEnvValue) > 0 {
		group = consumerGroupEnvValue
	}

	log.Println("broker:" + brokers)
	log.Println("SASL_PASS:" + config.Net.SASL.Password)
	log.Println("ConsumerGruop:" + group)
	log.Println("Topic:" + topics)

	config.Net.TLS.Enable = true

	tlsConfig := &tls.Config{}

	caCert := "cacert body" // If it necessary
	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig.RootCAs = caCertPool
		tlsConfig.InsecureSkipVerify = true
	}

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	// Check the current offset.
	// controlClient, err := sarama.NewClient([]string{brokers}, config)
	// if err != nil {
	// 	log.Panicf("Error creating control client: %v", err)
	// }
	// latestOffset, err := controlClient.GetOffset(topics, 0, sarama.OffsetNewest)
	// if err != nil {
	// 	log.Panicf("Error getting newest offset: %v", err)
	// }
	// fmt.Printf("LatestOffset: %d\n", latestOffset)

	// admin, err := sarama.NewClusterAdminFromClient(controlClient)

	// if err != nil {
	// 	if !controlClient.Closed() {
	// 		controlClient.Close()
	// 	}
	// 	log.Panicf("error creating kafka admin: %s", err)
	// }

	// offsets, err := admin.ListConsumerGroupOffsets(group, map[string][]int32{
	// 	topics: {0},
	// })
	// if err != nil {
	// 	log.Panicf("error listing consumer group offsets: %s", err)
	// }

	// block := offsets.GetBlock(topics, 0)
	// if block == nil {
	// 	log.Panicf("error finding offset block for topic %s and partition %d", topics, 0)
	// }
	// fmt.Printf("***Block: %v\n", block)

	//time.Sleep(10 * time.Second)

	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer fmt.Println("Finished wg.Done")

		//for {
		if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
			log.Panicf("Error from consumer: %v", err)
		}
		if err != nil {
			fmt.Printf("Error: on client.Consume : %v", err)
		}
		if ctx.Err() != nil {
			fmt.Printf("Error: ctx.Err() %v", ctx.Err())
			return
		}
		consumer.ready = make(chan bool)
		//}
	}()

	log.Println("Sarama consumer up and running!...")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-consumed:
		log.Println("Consumed.")
		break
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	fmt.Println("a")
	cancel()
	fmt.Println("b")
	wg.Wait()
	fmt.Println("c")
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer struct
type Consumer struct {
	ready chan bool
}

// Setup method
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

// Cleanup method
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim method
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s\n", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
		consumed <- true
		break
	}
	return nil
}
