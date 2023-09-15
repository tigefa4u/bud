package httpwrap_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livebud/bud/pkg/middleware/httpwrap"
	"github.com/matryer/is"
)

func TestHeadersNormal(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-A", "A")
		w.Write([]byte("Hello, world!"))
		w.Header().Add("X-B", "B")
	})
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 200)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "")
}

func TestHeadersWrapped(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := httpwrap.New().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-A", "A")
		w.Write([]byte("Hello, world!"))
		w.Header().Add("X-B", "B")
	}))
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 200)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "B")
}

func TestWriteStatusNormal(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-A", "A")
		w.WriteHeader(201)
		w.Write([]byte("Hello, world!"))
		w.Header().Add("X-B", "B")
	})
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 201)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "Hello, world!")
}

func TestWriteStatusWrapped(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := httpwrap.New().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-A", "A")
		w.WriteHeader(201)
		w.Write([]byte("Hello, world!"))
		w.Header().Add("X-B", "B")
	}))
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 201)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "B")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "Hello, world!")
}

func TestFlushNormal(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-A", "A")
		w.WriteHeader(201)
		w.Write([]byte("Hello, world!"))
		flush, ok := w.(http.Flusher)
		if ok {
			flush.Flush()
			flush.Flush()
		}
		w.Header().Add("X-B", "B")
	})
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 201)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "Hello, world!")
}

func TestFlushWrapped(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := httpwrap.New().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
		w.Header().Add("X-A", "A")
		flush, ok := w.(http.Flusher)
		if ok {
			flush.Flush()
			w.Write([]byte("yoyo"))
			flush.Flush()
		}
		w.Header().Add("X-B", "B")
	}))
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 200)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "Hello, world!yoyo")
}

func TestFlushStatusWrapped(t *testing.T) {
	is := is.New(t)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	is.NoErr(err)
	h := httpwrap.New().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
		w.WriteHeader(201)
		w.Header().Add("X-A", "A")
		flush, ok := w.(http.Flusher)
		if ok {
			flush.Flush()
			w.Write([]byte("yoyo"))
			flush.Flush()
		}
		w.Header().Add("X-B", "B")
	}))
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, 201)
	is.Equal(res.Header.Get("X-A"), "A")
	is.Equal(res.Header.Get("X-B"), "")
	body, err := io.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "Hello, world!yoyo")
}

func TestMultipleWriteHeaders(t *testing.T) {
	is := is.New(t)
	req, err := http.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	is.NoErr(err)
	h := httpwrap.New().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusSeeOther)
	}))
	h.ServeHTTP(rec, req)
	res := rec.Result()
	is.Equal(res.StatusCode, http.StatusSeeOther)
	req, err = http.NewRequest("POST", "/", nil)
	is.NoErr(err)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res = rec.Result()
	is.Equal(res.StatusCode, http.StatusSeeOther)
}
