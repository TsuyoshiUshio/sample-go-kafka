package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

var (
	brokers = ""
	version = ""
	topics  = ""
	count   = 0
	message = ""
	verbose = false
)

const (
	saslPassword = "SASL_PASSWORD"
	brokerList   = "BROKER_LIST"
	topicsEnv    = "TOPICS"
)

func init() {
	flag.StringVar(&brokers, "brokers", "", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&topics, "topics", "workitems", "Kafka topics to be consumed, as a comma separated list")
	flag.BoolVar(&verbose, "verbose", true, "Sarama logging")
	flag.IntVar(&count, "n", 1, "Number of the messages you want to send. default 1")
	flag.StringVar(&message, "m", "message: hello", "Message that you want to send")
	fmt.Println("Init end*****")
}

func main() {
	log.Println("Starting a new Sarama producer")
	fmt.Println("Parse*****")
	flag.Parse()
	flag.PrintDefaults()
	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	//	version, err := sarama.ParseKafkaVersion(version)
	//	if err != nil {
	//		log.Panicf("Error parsing Kafka version: %v", err)
	//	}

	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0

	config.Net.SASL.Enable = true
	config.Net.SASL.User = "$ConnectionString"
	config.Net.SASL.Password = os.Getenv(saslPassword)

	// config.Producer.Retry.Max = 1   // default 3
	// config.Producer.RequiredAcks = sarama.WaitForAll
	// config.Producer.Return.Successes = true

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

	log.Println("broker:" + brokers)
	log.Println("SASL_PASS:" + config.Net.SASL.Password)
	log.Println("Topic:" + topics)
	log.Printf("Count: %d\n", count)
	log.Printf("Message: %s\n", message)

	config.Net.TLS.Enable = true

	tlsConfig := &tls.Config{}

	caCert := "cacert body" // If it necessary
	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig.RootCAs = caCertPool
		tlsConfig.InsecureSkipVerify = true
	}
	brokerList := strings.Split(brokers, ",")
	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Panicf("Error creating producer client: %v", err)
	}
	for i := 0; i < count; i++ {
		producer.Input() <- &sarama.ProducerMessage{Topic: topics, Value: sarama.StringEncoder(message)}
		log.Printf("Message: '%s' sent", message)
	}

	if err := producer.Close(); err != nil {
		log.Panicf("Error closing producer client: %v", err)

	}
}
