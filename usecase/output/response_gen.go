// Code generated genHandler. DO NOT EDIT.

package output

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/bufferpool"
)

var okBody = map[string]string{"message": "ok"}

type Response interface {
	Render(http.ResponseWriter)
}

type jsonErr struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func renderJSON(w http.ResponseWriter, status int, v interface{}) {
	if v == nil {
		je := jsonErr{
			Message: "nil response",
			Error:   http.StatusText(http.StatusInternalServerError),
		}
		renderJSON(w, status, je)
		return
	}
	if err, ok := v.(error); ok {
		je := jsonErr{
			Message: err.Error(),
			Error:   http.StatusText(status),
		}
		renderJSON(w, status, je)
		return
	}
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		renderJSON(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func renderPEM(w http.ResponseWriter, status int, pem []byte) {
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.WriteHeader(status)
	w.Write(pem)
}

func renderPNG(w http.ResponseWriter, status int, png []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(status)
	w.Write(png)
}

func renderJSONWithSessionID(w http.ResponseWriter, status int, err error, sessid string) {
	if err != nil {
		renderJSON(w, status, err)
		return
	}
	w.Header().Set("X-SESSION-ID", sessid)
	renderJSON(w, status, map[string]string{"message": "ok"})
}

type ExistsUserResponse struct {
	Status int
	Err    error

	Exists bool
}

func (resp ExistsUserResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}

type CreateUserResponse struct {
	Status int
	Err    error

	Message string

	SessionID string
}

func (resp CreateUserResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSONWithSessionID(w, resp.Status, resp.Err, resp.SessionID)
}

type TOTPQRCodeResponse struct {
	Status int
	Err    error

	QRCode []byte
}

func (resp TOTPQRCodeResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}

type VerifyTOTPResponse struct {
	Status int
	Err    error

	Message string
}

func (resp VerifyTOTPResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}

type UpdateEmailResponse struct {
	Status int
	Err    error

	Message string
}

func (resp UpdateEmailResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}

type VerifyEmailResponse struct {
	Status int
	Err    error

	Message string
}

func (resp VerifyEmailResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}

type AuthByTOTPResponse struct {
	Status int
	Err    error

	Message string

	SessionID string
}

func (resp AuthByTOTPResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSONWithSessionID(w, resp.Status, resp.Err, resp.SessionID)
}

type AuthByPasswordResponse struct {
	Status int
	Err    error

	Token string
}

func (resp AuthByPasswordResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}

type GetPublicKeyResponse struct {
	Status int
	Err    error

	PublicKeyPEM []byte
}

func (resp GetPublicKeyResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, okBody)
}