/*
espubsub is an http handler which subscribes to Redis PubSub channels
and publishes the result to a 'Server-sent event'

It accepts clients on any url and subscribes to any channel matching the url.

requires a running Redis server

http://redis.io/topics/pubsub
http://en.wikipedia.org/wiki/Server-sent_events

Note: The Redis db is not relevant, since it is not used in Pub/Sub.

Example:
Enable clients to subsribe to channels starting with '/events/'

  sub := New(":6379","",-1)  // -1 is the default Redis db
  defer sub.Close()
  http.Handle("/events/", sub)

The above would create a new handler for will clients connecting to
/events/foo or /events/bar or /events/zoo etc

These clients will receive data published from Redis on the matching channel
  redis-cli PUBLISH /events/bar "Hello"
*/
package espubsub

import (
	eventsource "github.com/antage/eventsource/http"
	redis "github.com/vmihailenco/redis"
	"log"
	"net/http"
)

type ESPubSub interface {
	// implements the ServerHTTP method
	http.Handler

	// close all Redis clients and EventSource connections
	Close()

	// list all channels subscribed too
	Channels() []string
}

type subCollection struct {
	subscriptions map[string]subscription
	redisClient   *redis.Client
}

// binds a Redis subsciptions to an EventSource
type subscription struct {
	pubsub  *redis.PubSubClient
	es      eventsource.EventSource
	redisCh chan *redis.Message
	pubChan string // channel to listen on
}

// listen for published events and send to the EventSource
func (sc *subCollection) open(s subscription) {
  firstMessage := false

	for {
		msg, ok := <-s.redisCh
    if firstMessage && s.es.ConsumersCount() == 0 {
			log.Printf("no more consumers on %s", s.pubChan)
      sc.remove(s)
      return
    }
    firstMessage = true
		if ok {
			s.es.SendMessage(msg.Message, "", "")
			log.Printf("message on %s (consumers: %d)", s.pubChan, s.es.ConsumersCount())
		}
	}
}

// remove the subscription from the collection and close it
func (sc *subCollection) remove(s subscription){
  s.close()
  delete(sc.subscriptions, s.pubChan)
}

func (s subscription) close() {
	log.Printf("closing %s", s.pubChan)
	s.pubsub.Close()
	s.es.Close()
}

func (sc *subCollection) Close() {
	for _, s := range sc.subscriptions {
		s.close()
	}
}

func (sc *subCollection) newSubscription(pubChan string) {
	log.Printf("creating channel %s", pubChan)

	pubsub, err := sc.redisClient.PubSubClient()
	if err != nil {
		panic(err)
	}

	redisCh, err := pubsub.PSubscribe(pubChan)
	if err != nil {
		panic(err)
	}

	es := eventsource.New(nil)
	s := subscription{pubsub, es, redisCh, pubChan}
	sc.subscriptions[pubChan] = s
}

func (sc *subCollection) Channels() (channels []string) {
	for _, s := range sc.subscriptions {
		channels = append(channels, s.pubChan)
	}
	return
}

func (sc *subCollection) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	pubChan := req.URL.Path[0:]
	_, existing := sc.subscriptions[pubChan]

	if !existing {
		sc.newSubscription(pubChan)
		go sc.open(sc.subscriptions[pubChan])
	}

	log.Printf("subscribed to %s", pubChan)

	sc.subscriptions[pubChan].es.ServeHTTP(resp, req)
}

// Creates a new ESPubSub handler.
func New(redisHost, redisPassword string, redisDb int64) ESPubSub {
	s := new(subCollection)
	s.subscriptions = map[string]subscription{}

	s.redisClient = redis.NewTCPClient(redisHost, redisPassword, redisDb)
	return s
}
