// Code generated by genHandler. DO NOT EDIT.

package ident

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nasa9084/ident/infra"
)

func bindRoutes(router *mux.Router, env *infra.Environment) {
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedHandler)
	router.HandleFunc(`/v1/auth/totp`, AuthByTOTPHandler(env)).Methods(http.MethodPost)
	router.HandleFunc(`/v1/auth/password`, AuthByPasswordHandler(env)).Methods(http.MethodPost)
	router.HandleFunc(`/v1/publickey`, GetPublicKeyHandler(env)).Methods(http.MethodGet)
	router.HandleFunc(`/v1/user/exists/{user_id}`, ExistsUserHandler(env)).Methods(http.MethodGet)
	router.HandleFunc(`/v1/user`, CreateUserHandler(env)).Methods(http.MethodPost)
	router.HandleFunc(`/v1/user/totp`, TOTPQRCodeHandler(env)).Methods(http.MethodGet)
	router.HandleFunc(`/v1/user/totp`, VerifyTOTPHandler(env)).Methods(http.MethodPut)
	router.HandleFunc(`/v1/user/email`, UpdateEmailHandler(env)).Methods(http.MethodPut)
	router.HandleFunc(`/v1/user/email/{sessid}`, VerifyEmailHandler(env)).Methods(http.MethodGet)
}
