package output

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/bufferpool"
)

type Response interface {
	Render(http.ResponseWriter)
}

type jsonErr struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func renderJSON(w http.ResponseWriter, status int, v interface{}) {
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

type IsUserExistsResponse struct {
	Status int
	Err    error

	Exists bool
}

func (resp IsUserExistsResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]bool{"status": resp.Exists})
}

type CreateUserResponse struct {
	Status int
	Err    error

	SessionID string
}

func (resp CreateUserResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	w.Header().Set("X-SESSION-ID", resp.SessionID)
	renderJSON(w, resp.Status, map[string]string{"message": "ok"})
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
	renderPNG(w, resp.Status, resp.QRCode)
}

type VerifyTOTPResponse struct {
	Status int
	Err    error
}

func (resp VerifyTOTPResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
}

type UpdateEmailResponse struct {
	Status int
	Err    error
}

func (resp UpdateEmailResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
}

type VerifyEmailResponse struct {
	Status int
	Err    error
}

func (resp VerifyEmailResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
}

type AuthByTOTPResponse struct {
	Status int
	Err    error

	SessionID string
}

func (resp AuthByTOTPResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}

	w.Header().Set("X-SESSION-ID", resp.SessionID)
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
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

	renderJSON(w, resp.Status, map[string]string{"token": resp.Token})
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

	renderPEM(w, resp.Status, resp.PublicKeyPEM)
}
