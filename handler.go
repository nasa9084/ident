package ident

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lestrrat-go/bufferpool"
	"github.com/nasa9084/ident/infra"
	"github.com/nasa9084/ident/usecase"
	"github.com/nasa9084/ident/usecase/input"
	"github.com/pkg/errors"
)

func parseRequest(r *http.Request, dest input.Request) error {
	if r.Method != http.MethodGet {
		if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
			return errors.Wrap(err, `parsing request body`)
		}
	}
	if sessReq, ok := dest.(input.SessionRequest); ok {
		authorization := r.Header.Get("Authorization")
		if strings.Contains(authorization, " ") {
			return errors.New("authorization header invalid")
		}
		sessReq.SetSessionID(authorization)
	}
	if arReq, ok := dest.(input.PathArgsRequest); ok {
		pathArgs := mux.Vars(r)
		arReq.SetPathArgs(pathArgs)
	}
	return dest.Validate()
}

func renderErr(w http.ResponseWriter, err error) {
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)

	v := map[string]string{
		"error":   http.StatusText(http.StatusBadRequest),
		"message": err.Error(),
	}
	json.NewEncoder(buf).Encode(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	buf.WriteTo(w)
}

// NotFoundHandler is a HTTP handler for 404 Not Found.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)

	v := map[string]string{
		"error":   http.StatusText(http.StatusNotFound),
		"message": "endpoint not found",
	}
	json.NewEncoder(buf).Encode(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	buf.WriteTo(w)
}

// MethodNotAllowedHandler is a HTTP handler for 405 Method Not Allowed.
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)

	v := map[string]string{
		"error":   http.StatusText(http.StatusMethodNotAllowed),
		"message": fmt.Sprintf("method %s is not allowed", r.Method),
	}
	json.NewEncoder(buf).Encode(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	buf.WriteTo(w)
}

// CreateUserHandler creates a new user.
func CreateUserHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.CreateUserRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.CreateUser(r.Context(), req, env).Render(w)
	}
}

// TOTPQRCodeHandler returns TOTP URI QRcode.
func TOTPQRCodeHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.TOTPQRCodeRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.GetTOTPQRCode(r.Context(), req, env).Render(w)
	}
}

// VerifyTOTPHandler verifies TOTP is successfully configured.
func VerifyTOTPHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.VerifyTOTPRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.VerifyTOTP(r.Context(), req, env).Render(w)
	}
}

// UpdateEmailHandler updates user email.
func UpdateEmailHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.UpdateEmailRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.UpdateEmail(r.Context(), req, env).Render(w)
	}
}

// VerifyEmailHandler verifies the user's email is valid.
func VerifyEmailHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.VerifyEmailRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.VerifyEmail(r.Context(), req, env).Render(w)
	}
}

// AuthByTOTPHandler authenticates with user ID and TOTP Token.
// This handler returns session ID if the authentication is passed.
func AuthByTOTPHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.AuthByTOTPRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.AuthByTOTP(r.Context(), req, env).Render(w)
	}
}

// AuthByPasswordHandler authenticates with session ID and Password.
func AuthByPasswordHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req input.AuthByPasswordRequest
		if err := parseRequest(r, &req); err != nil {
			renderErr(w, err)
			return
		}
		usecase.AuthByPassword(r.Context(), req, env).Render(w)
	}
}

// GetPublicKeyHandler returns ECDSA public key generated from private key
// for signing JWT token.
func GetPublicKeyHandler(env *infra.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usecase.GetPublicKey(r.Context(), env).Render(w)
	}
}
