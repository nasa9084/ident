package output

import (
	"bytes"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

type mockResponseWriter struct {
	header http.Header
	status int
	body   []byte
}

func (w *mockResponseWriter) Header() http.Header {
	return w.header
}

func (w *mockResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.body = b
	return len(b), nil
}

func TestRenderJSON(t *testing.T) {
	candidates := []struct {
		input    interface{}
		expected []byte
	}{
		{map[string]string{"foo": "bar"}, []byte(`{"foo":"bar"}`)},
		{[]string{"foo", "bar"}, []byte(`["foo","bar"]`)},
		{nil, []byte(`{"message":"nil response","error":"Internal Server Error"}`)},
		{errors.New("some error"), []byte(`{"message":"some error","error":"OK"}`)},
	}

	for _, c := range candidates {
		w := &mockResponseWriter{header: http.Header{}}
		renderJSON(w, http.StatusOK, c.input)

		if !reflect.DeepEqual(bytes.Trim(w.body, "\n"), c.expected) {
			t.Errorf("%s != %s", w.body, c.expected)
			return
		}
	}
}
