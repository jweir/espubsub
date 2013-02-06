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
  "fmt"
	"net/http"
)

func main() {
	sub := espubsub.New(":6379", "", -1)
	defer sub.Close()
	http.Handle("/events/", sub)
	http.HandleFunc("/", index)
  log.Print("ESPubSub now running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, template())
}

func template() string {
  return `
<!DOCTYPE html>

<html>
<head>
<title>ESPubSub Demo</title>
<link href='http://fonts.googleapis.com/css?family=Cabin:400,700,400italic,700italic' rel='stylesheet' type='text/css'>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/1.9.0/jquery.min.js"></script>
<style type="text/css">
  body {
    margin: 20px;
    font-family: 'Cabin', san-serif;
    font-size: 14px/18px
  }

  th, td {
    border: 1px #EEE solid;
    margin: 0 -1px -1px 0;
    text-align: left;
    padding: 9px;
    vertical-align: top;
    width: 33%;
  }

  code {
    background: #D0D0D0;
    text-shadow: 0px 1px 1px #FFF;
    padding: 3px;
    border-radius: 4px;
  }
</style>
</head>

<body>
<h1>ESPubSub Demo</h1>
  <p>
    This client is listening on 3 channels: /events/alpha, /events/beta/ and /events/*
  </p>

  <p>
    Send a message from the command line <br/> <code>redis-cli PUBLISH /events/alpha "my message"</code>
  </p>

  <table border="0" style="width:50%">
    <tr>
      <th>alpha</th><th>beta</th><th>*</th>
    </tr>
    <tr>
      <td id="alpha"></td><td id="beta"></td><td id="both"></td>
    </tr>
  </table>

<script>
function listen(channel, target){
  var source = new EventSource('/events/'+channel);
  source.onmessage = function(e){
    target.prepend($("<div>"+e.data+"</div>"))
  }
}

listen("alpha", $("#alpha"))
listen("beta", $("#beta"))
listen("*", $("#both"))
</script>
</body>

</html>

  `
}
