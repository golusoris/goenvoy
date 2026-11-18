package fixture

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type fprintServer struct{}

// FprintNamedHandler writes a dynamic value to the response writer in a named handler.
func FprintNamedHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	// ruleid: fprint-to-responsewriter
	if _, err := fmt.Fprintf(w, "%s", key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintLiteralHandler returns a handler that writes a concatenated value.
func FprintLiteralHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("a")
		val := r.URL.Query().Get("b")
		// ruleid: fprint-to-responsewriter
		if _, err := fmt.Fprint(w, key+val); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// FprintlnMethod writes a dynamic value in a method ServeHTTP.
func (s fprintServer) FprintlnMethod(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Path
	// ruleid: fprint-to-responsewriter
	if _, err := fmt.Fprintln(w, v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintfChecked returns a literal handler that checks the Fprintf error.
func FprintfChecked() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Path
		// ruleid: fprint-to-responsewriter
		if _, err := fmt.Fprintf(w, "%s", v); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// FprintHelperSecond writes a dynamic value in a helper with the writer as a non-first parameter.
func FprintHelperSecond(tag string, w http.ResponseWriter) {
	v := tag
	// ruleid: fprint-to-responsewriter
	if _, err := fmt.Fprintf(w, "%s %s", tag, v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintConstant writes a constant string literal to the response writer.
func FprintConstant(w http.ResponseWriter, r *http.Request) {
	// ok: fprint-to-responsewriter
	if _, err := fmt.Fprint(w, "constant"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintlnRaw writes a raw string literal to the response writer.
func FprintlnRaw(w http.ResponseWriter, r *http.Request) {
	// ok: fprint-to-responsewriter
	if _, err := fmt.Fprintln(w, `{...}`); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintfBuffer writes a dynamic value to a bytes.Buffer inside a handler.
func FprintfBuffer(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	v := r.URL.Path
	// ok: fprint-to-responsewriter
	if _, err := fmt.Fprintf(buf, "%s", v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if _, err := io.Copy(w, buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintfStderr writes a dynamic value to os.Stderr inside a handler.
func FprintfStderr(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Path
	// ok: fprint-to-responsewriter
	if _, err := fmt.Fprintf(os.Stderr, "%s", v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// IocopyHandler writes using io.Copy with a strings.Reader.
func IocopyHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Path
	// ok: fprint-to-responsewriter
	if _, err := io.Copy(w, strings.NewReader(v)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// FprintfWriterParam writes a dynamic value to an io.Writer parameter of a non-handler function.
func FprintfWriterParam(w io.Writer, v string) {
	// ok: fprint-to-responsewriter
	if _, err := fmt.Fprintf(w, "%s", v); err != nil {
		return
	}
}
