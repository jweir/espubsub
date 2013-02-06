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

	// close all Redis channels and EventSource connections
	Close()

	// list all channels subscribed too
	Channels() []string
}

type chanCollection struct {
	channels    map[string]channel
	redisClient *redis.Client
}

// binds a Redis subsciptions to an EventSource
type channel struct {
	pubsub    *redis.PubSubClient
	es        eventsource.EventSource
	redisCh   chan *redis.Message
	id        string // channel to listen on
	activated bool   // has this channel received a message yet
}

// the number of consumers subscibed to the channel
func (s channel) consumers() int {
	return s.es.ConsumersCount()
}

func (s channel) close() {
	log.Printf("closing %s", s.id)
	s.pubsub.Close()
	s.es.Close()
}

// listen for published events and send to the EventSource
func (sc *chanCollection) open(s channel) {
	for {
		msg, ok := <-s.redisCh
		if s.activated && s.consumers() == 0 {
			log.Printf("zero consumers on %s", s.id)
			sc.remove(s)
			return
		}
		if ok {
			s.activated = true
			s.es.SendMessage(msg.Message, "", "")
			log.Printf(">> %s (consumers: %d)", s.id, s.consumers())
		}
	}
}

// remove the channel from the collection and close it
func (sc *chanCollection) remove(s channel) {
	s.close()
	delete(sc.channels, s.id)
}

func (sc *chanCollection) newChannel(id string) {
	log.Printf("creating channel %s", id)

	pubsub, err := sc.redisClient.PubSubClient()
	if err != nil {
		panic(err)
	}

	redisCh, err := pubsub.PSubscribe(id)
	if err != nil {
		panic(err)
	}

	es := eventsource.New(nil)
	s := channel{pubsub, es, redisCh, id, false}
	sc.channels[id] = s
}

func (sc *chanCollection) Close() {
	for _, s := range sc.channels {
		sc.remove(s)
	}
}

func (sc *chanCollection) Channels() (channels []string) {
	for _, s := range sc.channels {
		channels = append(channels, s.id)
	}
	return
}

func (sc *chanCollection) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[0:]
	_, existing := sc.channels[id]

	if !existing {
		sc.newChannel(id)
		go sc.open(sc.channels[id])
	}

	log.Printf("subscribed to %s", id)

	sc.channels[id].es.ServeHTTP(resp, req)
}

// Creates a new ESPubSub handler.
func New(redisHost, redisPassword string, redisDb int64) ESPubSub {
	s := new(chanCollection)
	s.channels = map[string]channel{}

	s.redisClient = redis.NewTCPClient(redisHost, redisPassword, redisDb)
	return s
}
