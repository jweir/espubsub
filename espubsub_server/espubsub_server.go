// create a simple PubSub server
//  go run example.go
//
// in another terminal
//  curl http://localhost:8080/events/foo
//  # or open a web browser to http://localhost:8080
//
// and finally in another terminal
//  redis-cli PUBLISH /events/foo 'hello world'
package main

import (
	espubsub "github.com/jweir/espubsub"
	"log"
	"net/http"
)

func main() {
	sub := espubsub.New(":6379", "", -1)
	defer sub.Close()
	http.Handle("/events/", sub)
	http.Handle("/", http.FileServer(http.Dir("./pub")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
