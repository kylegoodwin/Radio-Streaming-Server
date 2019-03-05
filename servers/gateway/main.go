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
	"sync"
	"time"

	"github.com/assignments-zanewebbUW/servers/gateway/models/users"
	"github.com/assignments-zanewebbUW/servers/gateway/sessions"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"

	"github.com/assignments-zanewebbUW/servers/gateway/handlers"

	_ "github.com/go-sql-driver/mysql"
)

//Create docker  network
// docker network create __networkname__
// docker run -d --name redisServer --network __networkname__ redis

// Run docker container for mysql server ??
// sudo docker run -d --name mysqlServer --network gatewayNetwork -e MYSQL_ROOT_PASSWORD=PASS -e MYSQL_DATABASE=db zanewebb/zanemysql

//DSN will be something like username:password@protocol(address)/dbname
//							root:PASSWORD@tcp(dockerhostname)/dbname

func testHandler(w http.ResponseWriter, r *http.Request) {
	//log.Printf("Received a request and handled with testHandler")
	w.Write([]byte("Handled the test request"))
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	ADDR := os.Getenv("ADDR")
	if len(ADDR) == 0 {
		ADDR = ":443"
		//ADDR = ":8888"
	}

	//Include these when you deploy
	TLSCERT := os.Getenv("TLSCERT")
	if len(TLSCERT) == 0 {
		fmt.Println("TLSCERT env variable was not set")
		os.Exit(1)
	}

	TLSKEY := os.Getenv("TLSKEY")
	if len(TLSKEY) == 0 {
		fmt.Println("TLSKEY env variable was not set")
		os.Exit(1)
	}

	sessionkey := os.Getenv("SESSIONKEY")
	if len(sessionkey) == 0 {
		fmt.Println("SESSIONKEY env variable was not set")
		os.Exit(1)
	}

	redisaddr := os.Getenv("REDISADDR")
	if len(redisaddr) == 0 {
		//redisaddr = "172.17.0.2:6379"
		redisaddr = "redisServer:6379"
	}

	//3306
	dsn := os.Getenv("DSN")
	if len(dsn) == 0 {
		fmt.Println("DSN env variable was not set")
		os.Exit(1)
	}

	MESSAGESADDR := os.Getenv("MESSAGESADDR")
	if len(MESSAGESADDR) == 0 {
		fmt.Println("MESSAGESADDR env variable was not set")
		os.Exit(1)
	}
	//Parse comma delimited service URL strings and turn them into URL objects
	msgAddresses := strings.Split(MESSAGESADDR, ",")
	var msgURLs []*url.URL
	for _, s := range msgAddresses {
		fmt.Printf("Parsing addr of: %s", s)
		u, err := url.Parse(s)
		if err != nil {
			fmt.Printf("Error parsing message URLs: %v", err)
			os.Exit(1)
		}
		msgURLs = append(msgURLs, u)
	}

	SUMMARYADDR := os.Getenv("SUMMARYADDR")
	if len(SUMMARYADDR) == 0 {
		fmt.Println("SUMMARYADDR env variable was not set")
		os.Exit(1)
	}
	//Parse comma delimited service URL strings and turn them into URL objects
	summaryAddresses := strings.Split(SUMMARYADDR, ",")
	var summarURLs []*url.URL
	for _, s := range summaryAddresses {
		fmt.Printf("Parsing addr of: %s", s)
		u, err := url.Parse(s)
		if err != nil {
			fmt.Printf("Error parsing message URLs: %v", err)
			os.Exit(1)
		}
		summarURLs = append(summarURLs, u)
	}

	//USE STORE INITIALIZATION

	//Create DB object from SQL DB
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Error opening the database: %v", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error pinging the database: %v", err)
		os.Exit(1)
	}

	//When comeplete, close the db
	defer db.Close()

	//Create mysqlstore
	usersStore := users.NewMySQLStore(db)

	//Create redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})

	//Create redisstore
	sessionStore := sessions.NewRedisStore(redisClient, time.Hour)

	//Create context
	context := handlers.NewContext(sessionkey, sessionStore, usersStore)

	//Initialize the tree on server startup
	context.UsersStore.PopulateTrie()

	//Microservice reverse proxy setup
	//It wants a URL not a string, example from exercise does not consider this issue
	messagingRProxy := &httputil.ReverseProxy{Director: context.UserDirector(msgURLs)}
	summaryRProxy := &httputil.ReverseProxy{Director: context.UserDirector(summarURLs)}

	//RabbitMQ Setup 5672?
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"MessagingQ", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	//When main consumes from the queue, its only job is to send back to the clients that something has occured in messaging.js
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

	////Create SocketStore
	wss := &handlers.SocketStore{
		Connections: make(map[int64]*websocket.Conn),
		Lock:        &sync.Mutex{},
		Cont:        context,
	}

	//Go routine kicked off, ready to read from rabbit queue
	go wss.SendMessages(msgs)

	//Go mux setup
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ws", wss.SocketConnectionHandler)
	mux.HandleFunc("/v1/users", context.UsersHandler)
	mux.HandleFunc("/v1/users/", context.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", context.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", context.SpecificSessionHandler)
	mux.HandleFunc("/v1/test", testHandler)
	mux.Handle("/v1/summary", summaryRProxy)
	mux.Handle("/v1/channels", messagingRProxy)
	mux.Handle("/v1/channels/", messagingRProxy)
	mux.Handle("/v1/messages/", messagingRProxy)

	wrappedMux := handlers.NewCors(mux)

	log.Printf("Server running and listening on %s", ADDR)
	//log.Fatal(http.ListenAndServe(ADDR, wrappedMux))
	log.Fatal(http.ListenAndServeTLS(ADDR, TLSCERT, TLSKEY, wrappedMux))
}
