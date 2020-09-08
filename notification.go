// Golang HTML5 Server Side Events Example
//
// Run this code like:
//  > go run server.go
//
// Then open up your browser to http://localhost:8000
// Your browser must support HTML5 SSE, of course.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/julienschmidt/httprouter"
)

//UserDetails hold all details of a person
type UserDetails struct {
	Username          string
	Email             string
	Name              string
	NotificationCount int
}

var username = "komalbhalge"
var redisConn redis.Conn
var broker *Broker

var templates *template.Template

//GetRedisConn return a single stable connection
func GetRedisConn() redis.Conn {
	fmt.Println("Redis demo started!")
	pool := newPool()
	conn := pool.Get()
	//defer conn.Close()
	err := ping(conn)
	if err != nil {
		fmt.Println(err)
	}
	return conn
}

// A single Broker will be created in this program. It is responsible
// for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (messages) to those clients.
//
type Broker2 struct {

	// Create a map of clients, the keys of the map are the channels
	// over which we can push messages to attached clients.  (The values
	// are just booleans and are meaningless.)
	//
	clients map[chan string]bool

	// Channel into which new clients can be pushed
	//
	newClients chan chan string

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan chan string

	// Channel into which messages are pushed to be broadcast out
	// to attahed clients.
	//
	messages chan string
}

// This Broker method starts a new goroutine.  It handles
// the addition & removal of clients, as well as the broadcasting
// of messages out to clients that are currently attached.
//
func (b *Broker) Start2() {

	// Start a goroutine
	//
	go func() {

		// Loop endlessly
		//
		for {

			// Block until we receive from one of the
			// three following channels.
			select {

			case s := <-b.newClients:

				// There is a new client attached and we
				// want to start sending them messages.
				b.clients[s] = true
				log.Println("Added new client")

			case s := <-b.defunctClients:

				// A client has dettached and we want to
				// stop sending them messages.
				delete(b.clients, s)
				close(s)

				log.Println("Removed client")

			case msg := <-b.messages:

				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				for s := range b.clients {
					s <- msg
				}
				log.Printf("Broadcast message to %d clients", len(b.clients))
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
//
func (b *Broker) ServeHTTP2(w http.ResponseWriter, r *http.Request) {

	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel, over which the broker can
	// send this client messages.
	messageChan := make(chan string)

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- messageChan

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.
		b.defunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Don't close the connection, instead loop endlessly.
	for {

		// Read from our messageChan.
		msg, open := <-messageChan

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			break
		}

		// Write to the ResponseWriter, `w`.
		fmt.Fprintf(w, "data: Received Message: %s\n\n", msg)

		// Flush the response.  This is only possible if
		// the repsonse supports streaming.
		f.Flush()
	}

	// Done.
	log.Println("Finished HTTP request at ", r.URL.Path)
}

// Handler for the main page, which we wire up to the
// route at "/" below in `main`.
//

// Main routine
func initNotifications() {
	templates = template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	// Make b the HTTP handler for "/events/".  It can do
	// this because it has a ServeHTTP method.  That method
	// is called in a separate goroutine for each
	// request to "/events/".
	//http.Handle("/events/", broker)

	// Generate a constant stream of events that get pushed
	// into the Broker's messages channel and are then broadcast
	// out to any clients that are attached.
	/*go func() {
		for i := 0; ; i++ {

			// Create a little message to send to clients,
			// including the current time.
			broker.messages <- fmt.Sprintf("%d - the time is %v", i, time.Now())

			// Print a nice log message and sleep for 5s.
			log.Printf("Sent message %d ", i)
			time.Sleep(5e9)

		}
	}() */

	// When we get a request at "/", call `handler`
	// in a new goroutine.
	//http.Handle("/", http.HandlerFunc(handler))

	// Start the server and listen forever on port 8000.
	router := setupRoutes2()
	http.ListenAndServe(":8000", router)
}
func startReceiver(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("startReceiver: ")

	// Make a new Broker instance
	broker = &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
	}

	// Start processing events
	broker.Start()
}
func setupRoutes2() *httprouter.Router {
	router := httprouter.New()
	router.GET("/", home)
	router.GET("/events/", startReceiver)
	router.GET("/addnotification", addNotification)
	return router
}
func home(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	templates.ExecuteTemplate(w, "notification.gohtml", nil)
}

func addNotification(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("addNotification: ")
	conn := GetRedisConn()
	nCount := 0
	oldUser, err := getUserDetails(conn, username)
	if err != nil {
		log.Println("addNotification Error: ", err.Error())
	}
	if oldUser.Name != "" {
		nCount = oldUser.NotificationCount + 1
	}
	fmt.Println("Old User: ", oldUser.NotificationCount)
	//Incvreate notification count by one
	usr := UserDetails{Username: username, Email: "komal.bhalge@yahoo.com", Name: "Komal Bhalge", NotificationCount: nCount}
	setUserDetails(conn, usr)

	broker.messages <- strconv.Itoa(usr.NotificationCount)

	// Print a nice log message and sleep for 5s.
	log.Printf("Sent message %d ", usr.NotificationCount)
	templates.ExecuteTemplate(w, "notification.gohtml", nil)

}

func setUserDetails(c redis.Conn, usr UserDetails) error {

	// serialize User object to JSON
	json, err := json.Marshal(usr)
	if err != nil {
		return err
	}
	// SET object
	_, err = c.Do("SET", usr.Username, json)
	if err != nil {
		return err
	}

	return nil
}
func getUserDetails(c redis.Conn, username string) (UserDetails, error) {

	usr := UserDetails{}
	s, err := redis.String(c.Do("GET", username))
	if err == redis.ErrNil {
		fmt.Println("User does not exist")
	} else if err != nil {

		return usr, err
	}

	err = json.Unmarshal([]byte(s), &usr)

	fmt.Printf("%+v\n", usr)

	return usr, nil
}

func initRedisDB() {
	fmt.Println("Redis demo started!")
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	err := ping(conn)
	if err != nil {
		fmt.Println(err)
	}
}
