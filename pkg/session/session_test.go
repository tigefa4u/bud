package session_test

import (
	"encoding/gob"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/livebud/bud/pkg/mux"
	"github.com/livebud/bud/pkg/session"
	"github.com/livebud/bud/pkg/session/cookiestore"
	"github.com/livebud/bud/pkg/session/internal/cookies"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
)

func equal(t testing.TB, jar *cookiejar.Jar, h http.Handler, r *http.Request, expect string) {
	t.Helper()
	for _, cookie := range jar.Cookies(r.URL) {
		r.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	w := rec.Result()
	jar.SetCookies(r.URL, w.Cookies())
	dump, err := httputil.DumpResponse(w, true)
	if err != nil {
		if err.Error() != expect {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	diff.TestHTTP(t, expect, string(dump))
}

func TestSetGetCookie(t *testing.T) {
	is := is.New(t)
	jar, err := cookiejar.New(nil)
	is.NoErr(err)
	router := mux.New()
	router.Get("/set", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "cookie_name", Value: "cookie_value"})
	}))
	router.Get("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("cookie_name")
		is.NoErr(err)
		http.SetCookie(w, cookie)
	}))
	req := httptest.NewRequest(http.MethodGet, "http://example.com/set", nil)
	equal(t, jar, router, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: cookie_name=cookie_value
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/get", nil)
	equal(t, jar, router, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: cookie_name=cookie_value
	`)
}

func TestSession(t *testing.T) {
	is := is.New(t)
	jar, err := cookiejar.New(nil)
	is.NoErr(err)
	sessions := session.New(cookiestore.New(cookies.New()))
	lastValue := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := sessions.Load(r, "sid")
		is.NoErr(err)
		visits := session.Increment("visits")
		is.Equal(visits, lastValue+1)
		lastValue++
		err = sessions.Save(w, r, session)
		is.NoErr(err)
	})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjF9Cg
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjJ9Cg
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjN9Cg
	`)
}

func TestSessionCounter(t *testing.T) {
	is := is.New(t)
	jar, err := cookiejar.New(nil)
	is.NoErr(err)
	type Session struct {
		Visits int `json:"visits"`
	}
	cookies := cookies.New()
	sessions := session.New(cookiestore.New(cookies))
	lastValue := -1
	handler := sessions.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := session.From(r.Context())
		is.NoErr(err)
		visits, ok := session.Int("visits")
		if !ok {
			visits = 0
		}
		is.Equal(visits, lastValue+1)
		visits++
		lastValue++
		session.Set("visits", visits)
	}))
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjF9Cg
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjJ9Cg
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjN9Cg
	`)
}

func TestSessionNested(t *testing.T) {
	is := is.New(t)
	jar, err := cookiejar.New(nil)
	is.NoErr(err)

	type User struct {
		ID int `json:"id"`
	}
	type Session struct {
		Visits int  `json:"visits"`
		User   User `json:"user,omitempty"`
	}
	gob.Register(User{})
	sessions := session.New(cookiestore.New(cookies.New()))
	handler := sessions.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := session.From(r.Context())
		is.NoErr(err)
		visits := session.Increment("visits")
		if visits == 2 {
			session.Set("user", &User{ID: 1})
		}
		if visits == 3 {
			user, ok := session.Get("user").(map[string]any)
			is.True(ok)
			is.Equal(user["id"], float64(1))
		}
		if visits == 4 {
			session.Delete("user")
		}
		if visits == 5 {
			_, ok := session.Get("user").(map[string]any)
			is.True(!ok)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjF9Cg
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ1c2VyIjp7ImlkIjoxfSwidmlzaXRzIjoyfQo
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ1c2VyIjp7ImlkIjoxfSwidmlzaXRzIjozfQo
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjR9Cg
	`)
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	equal(t, jar, handler, req, `
		HTTP/1.1 200 OK
		Connection: close
		Set-Cookie: sid=eyJ2aXNpdHMiOjV9Cg
	`)
}