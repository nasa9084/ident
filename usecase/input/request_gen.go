// Code generated by genHandler. DO NOT EDIT.

package input

import (
	"errors"
	"unicode"
)

var _ = unicode.UpperCase

type Request interface {
	Validate() error
}

type SessionRequest interface {
	Request
	SetSessionID(string)
}

type PathArgsRequest interface {
	Request
	SetPathArgs(map[string]string)
}

type VerifyEmailRequest struct {
	SessionID string `json:"-"`
}

func (r VerifyEmailRequest) Validate() error {
	switch {
	case r.SessionID == "":
		return errors.New("sessid is required")
	}
	return nil
}

type AuthByTOTPRequest struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

func (r AuthByTOTPRequest) Validate() error {
	switch {
	case r.UserID == "":
		return errors.New("user_id is required ")
	case r.Token == "":
		return errors.New("token is required ")
	case len(r.Token) != 6:
		return errors.New("length of token is not valid")
	}
	for _, r := range r.Token {
		if !unicode.IsDigit(r) {
			return errors.New("token must be digit")
		}
	}
	return nil
}

type AuthByPasswordRequest struct {
	Password string `json:"password"`

	SessionID string `json:"-"`
}

func (r AuthByPasswordRequest) Validate() error {
	switch {
	case r.SessionID == "":
		return errors.New("authorization header is required")
	case r.Password == "":
		return errors.New("password is required ")
	}
	return nil
}

func (r AuthByPasswordRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}

type ExistsUserRequest struct {
	UserID string `json:"-"`
}

func (r ExistsUserRequest) Validate() error {
	switch {
	case r.UserID == "":
		return errors.New("user_id is required")
	}
	return nil
}

type CreateUserRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

func (r CreateUserRequest) Validate() error {
	switch {
	case r.UserID == "":
		return errors.New("user_id is required ")
	case r.Password == "":
		return errors.New("password is required ")
	}
	return nil
}

type TOTPQRCodeRequest struct {
	SessionID string `json:"-"`
}

func (r TOTPQRCodeRequest) Validate() error {
	switch {
	case r.SessionID == "":
		return errors.New("authorization header is required")
	}
	return nil
}

func (r TOTPQRCodeRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}

type VerifyTOTPRequest struct {
	Token string `json:"token"`

	SessionID string `json:"-"`
}

func (r VerifyTOTPRequest) Validate() error {
	switch {
	case r.SessionID == "":
		return errors.New("authorization header is required")
	case r.Token == "":
		return errors.New("token is required ")
	case len(r.Token) != 6:
		return errors.New("length of token is not valid")
	}
	for _, r := range r.Token {
		if !unicode.IsDigit(r) {
			return errors.New("token must be digit")
		}
	}
	return nil
}

func (r VerifyTOTPRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}

type UpdateEmailRequest struct {
	Email string `json:"email"`

	SessionID string `json:"-"`
}

func (r UpdateEmailRequest) Validate() error {
	switch {
	case r.SessionID == "":
		return errors.New("authorization header is required")
	case r.Email == "":
		return errors.New("email is required ")
	}
	return nil
}

func (r UpdateEmailRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}
