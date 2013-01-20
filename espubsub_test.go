// test helpers taken from
// https://github.com/antage/eventsource/blob/master/http/eventsource_test.go
package espubsub

import (
	redis "github.com/vmihailenco/redis"
	"io"
	"net"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

type testEnv struct {
	redis  *redis.Client
	server *httptest.Server
	s      ESPubSub
}

func setup(t *testing.T) *testEnv {
	t.Log("creating a test env")
	e := new(testEnv)
	e.s = New(":6379", "", -1)
	e.redis = redis.NewTCPClient(":6379", "", 6)
	e.server = httptest.NewServer(e.s)
	return e
}

func teardown(t *testing.T, e *testEnv) {
	e.s.Close()
	e.redis.Close()
	e.server.Close()
}

func checkError(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func read(t *testing.T, c net.Conn) string {
	resp := make([]byte, 1024)
	_, err := c.Read(resp)
	if err != nil && err != io.EOF {
		t.Log(err)
	}
	return string(resp)
}

func expectResponse(t *testing.T, c net.Conn, expecting string) {
	time.Sleep(100 * time.Millisecond)
	resp := read(t, c)

	if !strings.Contains(resp, expecting) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, resp)
	}
}

func expectNotInResponse(t *testing.T, c net.Conn, expecting string) {
	time.Sleep(100 * time.Millisecond)
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	defer c.SetReadDeadline(time.Time{}) // clear the deadline

	resp := read(t, c)
	if strings.Contains(resp, expecting) {
		t.Errorf("not expecting:\n%s\ngot:\n%s\n", expecting, resp)
	}
}

func startEventStream(t *testing.T, e *testEnv, channel string) (net.Conn, string) {
	url := e.server.URL

	conn, err := net.Dial("tcp", strings.Replace(url, "http://", "", 1))
	checkError(t, err)

	t.Log("send GET request to the connection")
	_, err = conn.Write([]byte("GET " + channel + " HTTP/1.1\n\n"))
	checkError(t, err)

	resp := read(t, conn)
	t.Logf("got response: \n%s", resp)
	return conn, resp
}

func TestEventSourceGetsPublishedString(t *testing.T) {
	e := setup(t)
	defer teardown(t, e)

	conn0, _ := startEventStream(t, e, "/events/test")
	defer conn0.Close()

	conn2, _ := startEventStream(t, e, "/events/bar")
	defer conn2.Close()

	e.redis.Publish("/events/test", "hello world")
	expectResponse(t, conn0, "hello world")

	conn1, _ := startEventStream(t, e, "/events/test")
	defer conn1.Close()

	e.redis.Publish("/events/test", "new message")
	expectResponse(t, conn0, "new message")
	expectResponse(t, conn1, "new message")

	e.redis.Publish("/events/bar", "second channel")
	expectResponse(t, conn2, "second channel")
}

func TestPattenMatchingSubscriptions(t *testing.T) {
	e := setup(t)
	defer teardown(t, e)

	conn0, _ := startEventStream(t, e, "/events/f*")
	defer conn0.Close()

	conn1, _ := startEventStream(t, e, "/events/foo")
	defer conn1.Close()

	e.redis.Publish("/events/foo", "message 1")
	expectResponse(t, conn0, "message 1")
	expectResponse(t, conn1, "message 1")

	e.redis.Publish("/events/fee", "message 2")
	expectResponse(t, conn0, "message 2")
	expectNotInResponse(t, conn1, "message 2")

	e.redis.Publish("/events/bar", "message 3")
	expectNotInResponse(t, conn0, "message 3")
	expectNotInResponse(t, conn1, "message 3")
}

func TestConsumerCountAndChannels(t *testing.T) {
	e := setup(t)
	defer teardown(t, e)

	conn0, _ := startEventStream(t, e, "/events/bar")
	defer conn0.Close()

	conn1, _ := startEventStream(t, e, "/events/foo")
	defer conn1.Close()

	channels := e.s.Channels()
	sort.Strings(channels)
	actual := (strings.Join(channels, ","))
	expecting := ("/events/bar,/events/foo")

	if actual != expecting {
		t.Errorf("channels do not match %s, %s", expecting, actual)
	}

	time.Sleep(100 * time.Millisecond)
}
