package mylar

import "net/http"

// Why: a private pool; httptest.Server.Close and other callers close idle conns on http.DefaultTransport.
var sharedTransport = newTransport() //nolint:gochecknoglobals // one pool for every Client of the package, never exported

func newTransport() http.RoundTripper {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		return t.Clone()
	}
	return &http.Transport{Proxy: http.ProxyFromEnvironment}
}
