package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newcastile/snippetbox/internal/assert"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, req)

	rs := rr.Result()

	tests := []struct {
		name string
		val  string
		want string
	}{
		{
			name: "Content-Security-Policy",
			val:  rs.Header.Get("Content-Security-Policy"),
			want: "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		},
		{
			name: "Referrer-Policy",
			val:  rs.Header.Get("Referrer-Policy"),
			want: "origin-when-cross-origin",
		},
		{
			name: "X-Content-Type-Options",
			val:  rs.Header.Get("X-Content-Type-Options"),
			want: "nosniff",
		},
		{
			name: "X-Frame-Options",
			val:  rs.Header.Get("X-Frame-Options"),
			want: "deny",
		},
		{
			name: "X-XSS-Protection",
			val:  rs.Header.Get("X-XSS-Protection"),
			want: "0",
		},
		{
			name: "Status Code",
			val:  fmt.Sprint(rs.StatusCode),
			want: fmt.Sprint(http.StatusOK),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.val, tt.want)
		})
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")
}
