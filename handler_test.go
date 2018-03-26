package ident

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pkg/errors"
)

type testRequest struct {
	Key string `json:"key"`
}

func (r testRequest) Validate() error { return nil }

func TestParseRequest(t *testing.T) {
	candidates := []struct {
		label    string
		input    string
		isErr    bool
		expected testRequest
	}{
		{"valid", `{"key": "value"}`, false, testRequest{"value"}},
		{"empty object", `{}`, false, testRequest{}},
		{"empty body", ``, true, testRequest{}},
		{"invalid json(not quoted)", `{foo: bar}`, true, testRequest{}},
		{"invalid json(plain text)", `foobar`, true, testRequest{}},
		{"invalid json(list)", `["foo", "bar"]`, true, testRequest{}},
	}
	for _, c := range candidates {
		t.Log(c.label)
		buf := bytes.NewBufferString(c.input)
		hr, _ := http.NewRequest(http.MethodPost, "", buf)
		var tr testRequest

		if err := parseRequest(hr, &tr); !c.isErr && (err != nil) {
			t.Error(err)
			return
		}
		if tr != c.expected {
			t.Errorf("%+v != %+v", tr, c.expected)
			return
		}
	}
}

type mockResponseWriter struct {
	header http.Header
	status int
	body   []byte
}

func (w *mockResponseWriter) Header() http.Header {
	return w.header
}

func (w *mockResponseWriter) WriteHeader(st int) {
	w.status = st
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.body = b
	return len(b), nil
}

func TestRenderErr(t *testing.T) {
	w := &mockResponseWriter{header: http.Header{}}
	renderErr(w, errors.New("some error"))

	if w.header.Get("Content-Type") != "application/json" {
		t.Errorf("%s != application/json", w.header.Get("Content-Type"))
		return
	}
	if w.status != http.StatusBadRequest {
		t.Errorf("%d != %d", w.status, http.StatusBadRequest)
		return
	}

	body := map[string]string{}
	if err := json.Unmarshal(w.body, &body); err != nil {
		t.Error(err)
		return
	}
	if body["error"] != http.StatusText(http.StatusBadRequest) {
		t.Errorf("%s != %s", body["error"], http.StatusText(http.StatusBadRequest))
		return
	}
	if body["message"] != "some error" {
		t.Errorf("%s != some error", body["message"])
		return
	}
}
