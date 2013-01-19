Easily send messages downstream to a web browser on dedicated channels.

## Redis backed Pub/Sub EventSource server

ESPubSub is a Go HTTP Handler which allows clients to subscribe to channels via
**EventSource**. Event Source is a simple browser api for reading from a long
poll connection. Browsers(except IE and Android mobile) support it.

Channels receive data (strings) via a Redis PUBLISH command. ESPubSub
will need to connect to a running Redis server.

This has not been used in production.

Some EventSource polyfills (which I have not tried)

https://github.com/remy/polyfills/blob/master/EventSource.js

https://github.com/Yaffle/EventSource

### Demo
    git clone git@github.com:jweir/espubsub.git
    cd espubsub/examples
    go run server.go

Connect to http://localhost:8080/ with your browser

From the command line send Redis PUBLISH commands

    redis-cli PUBLISH /events/alpha "Hello Alpha"
    redis-cli PUBLISH /events/beta "`date`"

The demo shows has one channel(`/events/*`) using a glob to listen to multiple channels.

## Made Possible By

http://github.com/antage/eventsource and http://github.com/vmihailenco/redis

# LICENSE

Copyright (c) 2012 John Weir

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
