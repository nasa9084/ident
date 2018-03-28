package output

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/bufferpool"
	"github.com/nasa9084/ident/infra"
	sendgrid "github.com/sendgrid/sendgrid-go"
)

// Response interface renders into http response.
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

// IsUserExistsResponse is output type for IsUserExists.
type IsUserExistsResponse struct {
	Status int
	Err    error

	Exists bool
}

// Render implements Response interface.
func (resp IsUserExistsResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]bool{"status": resp.Exists})
}

// CreateUserResponse is output type for CreateUser.
type CreateUserResponse struct {
	Status int
	Err    error

	SessionID string
}

// Render implements Response interface.
func (resp CreateUserResponse) Render(w http.ResponseWriter) {
	renderJSONWithSessionID(w, resp.Status, resp.Err, resp.SessionID)
}

// TOTPQRCodeResponse is output type for TOTPQRCode.
type TOTPQRCodeResponse struct {
	Status int
	Err    error
	QRCode []byte
}

// Render implements Response interface.
func (resp TOTPQRCodeResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderPNG(w, resp.Status, resp.QRCode)
}

// VerifyTOTPResponse is output type for VerifyTOTP.
type VerifyTOTPResponse struct {
	Status int
	Err    error
}

// Render implements Response interface.
func (resp VerifyTOTPResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
}

// UpdateEmailResponse is output type for UpdateEmail.
type UpdateEmailResponse struct {
	Status int
	Err    error

	Mail      sendgrid.Client
	Email     string
	SessionID string
}

// Render implements Response interface.
func (resp UpdateEmailResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}

	_, err := resp.Mail.Send(infra.NewVerificationMail(resp.Email, resp.SessionID))
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err)
		return
	}
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
}

// VerifyEmailResponse is output type for VerifyEmail.
type VerifyEmailResponse struct {
	Status int
	Err    error
}

// Render implements Response interface.
func (resp VerifyEmailResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}
	renderJSON(w, resp.Status, map[string]string{"status": "ok"})
}

// AuthByTOTPResponse is output type for AuthByTOTP.
type AuthByTOTPResponse struct {
	Status int
	Err    error

	SessionID string
}

// Render implements Response interface
func (resp AuthByTOTPResponse) Render(w http.ResponseWriter) {
	renderJSONWithSessionID(w, resp.Status, resp.Err, resp.SessionID)
}

// AuthByPasswordResponse is output type for AuthByPassword.
type AuthByPasswordResponse struct {
	Status int
	Err    error

	Token string
}

// Render implements Response interface
func (resp AuthByPasswordResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}

	renderJSON(w, resp.Status, map[string]string{"token": resp.Token})
}

// GetPublicKeyResponse is output type fr GetPublicKey.
type GetPublicKeyResponse struct {
	Status int
	Err    error

	PublicKeyPEM []byte
}

// Render implements Response interface
func (resp GetPublicKeyResponse) Render(w http.ResponseWriter) {
	if resp.Err != nil {
		renderJSON(w, resp.Status, resp.Err)
		return
	}

	renderPEM(w, resp.Status, resp.PublicKeyPEM)
}
