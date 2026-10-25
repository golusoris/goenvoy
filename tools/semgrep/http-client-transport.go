package fixture

import (
	"net/http"
	"time"
)

var pool = http.DefaultTransport.(*http.Transport).Clone()

func clients() []*http.Client {
	// ruleid: http-client-without-transport
	a := &http.Client{Timeout: time.Second}
	// ruleid: http-client-without-transport
	b := &http.Client{}
	// ruleid: http-client-without-transport
	v := http.Client{Timeout: time.Second}
	// ok: http-client-without-transport
	c := &http.Client{Timeout: time.Second, Transport: pool}
	// ok: http-client-without-transport
	d := &http.Client{Transport: pool}
	// ok: http-client-without-transport
	w := http.Client{Transport: pool, Timeout: time.Second}
	return []*http.Client{a, b, &v, c, d, &w}
}
