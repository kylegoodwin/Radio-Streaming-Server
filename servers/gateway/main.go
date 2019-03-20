package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"os"
	"strings"
	"time"

	"github.com/Radio-Streaming-Server/servers/gateway/models/logins"
	"github.com/streadway/amqp"

	"github.com/Radio-Streaming-Server/servers/gateway/sessions"

	"github.com/Radio-Streaming-Server/servers/gateway/models/users"

	"github.com/go-redis/redis"

	"github.com/Radio-Streaming-Server/servers/gateway/handlers"
)

//main is the main entry point for the server

func main() {

	//Microservice Addresses

	SUMMARYADDRS := strings.Split(os.Getenv("SUMMARYADDRS"), ",")
	MESSAGESADDRS := strings.Split(os.Getenv("MESSAGESADDRS"), ",")

	addr := os.Getenv("ADDR")
	key := os.Getenv("TLSKEY")
	cert := os.Getenv("TLSCERT")
	sessionkey := os.Getenv("SESSIONKEY")
	redisaddr := os.Getenv("REDISADDR")
	dbaddr := os.Getenv("DBADDR")
	dsn := fmt.Sprintf("root:%s@tcp(%s)/website", os.Getenv("MYSQL_ROOT_PASSWORD"), dbaddr)

	if sessionkey == "" {
		fmt.Errorf("Session key was not defined")
	}

	options := redis.Options{}
	options.Addr = redisaddr
	redisClient := redis.NewClient(&options)

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		fmt.Errorf("Error opening sql database")

	}

	context := handlers.HandlerContext{}
	context.Key = sessionkey
	context.User = users.NewDBConnection(db)
	context.Session = sessions.NewRedisStore(redisClient, time.Hour)
	context.Login = logins.NewDBConnection(db)
	context.Trie, err = context.User.BuildTrie()
	context.Sockets = handlers.EstablishSockets()

	if err != nil {
		fmt.Errorf("Error building search Trie")
	}

	if len(addr) == 0 {
		addr = ":443"
	}

	if cert == "" || key == "" {
		log.Fatal("Certificates not set")
	}

	//RabitMQ Creation

	conn, err := amqp.Dial("amqp://guest:guest@rabbit:5672/")
	fmt.Println(err)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"messages", // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	//forever := make(chan bool)

	go context.Sockets.SendMessages(msgs)

	/*
		func() {
			for d := range msgs {
				//log.Printf("Received a message: %s", d.Body)
				var result map[string]string
				json.Unmarshal([]byte(d.Body), &result)
				fmt.Println("a message was recieved")
				context.Sockets.SendMessageToUsers([]int64{1}, result["channel"])
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	*/

	var parsedMessageURLS []*url.URL
	var parsedSummaryURLS []*url.URL

	for _, messageURL := range MESSAGESADDRS {
		tempMessageURL, err := url.Parse(messageURL)
		if err != nil {
			fmt.Errorf("Messages URL not working")
		}
		parsedMessageURLS = append(parsedSummaryURLS, tempMessageURL)

	}

	for _, summaryURL := range SUMMARYADDRS {
		tempSummaryURL, err := url.Parse(summaryURL)
		if err != nil {
			fmt.Errorf("Summary URL not working")
		}
		parsedSummaryURLS = append(parsedSummaryURLS, tempSummaryURL)
	}
	aURL, _ := url.Parse("http://audio-api:80")
	parsedAudioURLS := []*url.URL{aURL}

	messageProxy := &httputil.ReverseProxy{Director: handlers.CustomDirector(parsedMessageURLS, context)}
	summaryProxy := &httputil.ReverseProxy{Director: handlers.CustomDirector(parsedSummaryURLS, context)}
	audioProxy := &httputil.ReverseProxy{Director: handlers.CustomDirector(parsedAudioURLS, context)}

	//New mux
	mux := http.NewServeMux()

	//Handle the v1 call
	mux.Handle("/v1/audio", audioProxy)
	mux.Handle("/v1/audio/", audioProxy)
	mux.Handle("/v1/summary", summaryProxy)
	mux.Handle("/v1/channels", audioProxy)
	mux.Handle("/v1/channels/", audioProxy)
	mux.Handle("/v1/comments", messageProxy)
	mux.Handle("/v1/comments/", messageProxy)
	mux.Handle("/v1/status/", messageProxy)
	mux.HandleFunc("/ws", context.WSUpgrade)
	mux.HandleFunc("/v1/users", context.UsersHandler)
	mux.HandleFunc("/v1/users/", context.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", context.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", context.SpecificSessionHandler)

	wrapper := &handlers.Cors{}
	wrapper.Handler = mux

	log.Fatal(http.ListenAndServeTLS(addr, cert, key, wrapper))

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
