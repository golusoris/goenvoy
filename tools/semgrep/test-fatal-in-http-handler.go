package fixture

import (
	"encoding/json"
	"net/http"
	"testing"
)

type server struct {
	t *testing.T
}

var testT *testing.T

func namedHandler(w http.ResponseWriter, r *http.Request) {
	// ruleid: test-fatal-in-http-handler
	testT.Fatal("named handler fatal")
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ruleid: test-fatal-in-http-handler
	s.t.Fatal("method handler fatal")
}

func use(h http.Handler) {
	h.ServeHTTP(nil, nil)
}

func respond(t *testing.T, w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// ruleid: test-fatal-in-http-handler
		t.Fatal(err)
	}
}

func helperWriterSecond(tag string, w http.ResponseWriter, t *testing.T) {
	// ruleid: test-fatal-in-http-handler
	t.Fatal(tag)
}

func decodeRequest(t *testing.T, r *http.Request) {
	// ok: test-fatal-in-http-handler
	t.Fatal("decode request")
}

func handlers(t *testing.T) {
	// ok: test-fatal-in-http-handler
	t.Fatal("outside handler fatal")

	closure := func() {
		// ok: test-fatal-in-http-handler
		t.Fatal("non-handler closure fatal")
	}
	closure()

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.Fatal("literal fatal")
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.Fatalf("literal fatalf: %s", "err")
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.FailNow()
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.Skip("literal skip")
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.Skipf("literal skipf: %s", "skip")
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.SkipNow()
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// ruleid: test-fatal-in-http-handler
		t.Fatal("blank request param")
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// ruleid: test-fatal-in-http-handler
			t.Fatal("nested in if")
		}
	}))

	use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ok: test-fatal-in-http-handler
		t.Errorf("ok errorf: %s", "err")
		return
	}))
}
