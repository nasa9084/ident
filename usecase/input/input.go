package input

import (
	"errors"
	"unicode"
)

// Request interface which has Validate method.
type Request interface {
	Validate() error
}

// SessionRequest interface represents a request with session ID.
type SessionRequest interface {
	Request
	SetSessionID(string)
}

// PathArgsRequest interface represents a request with path variable(s).
type PathArgsRequest interface {
	Request
	SetPathArgs(map[string]string)
}

// IsUserExistsRequest is input type for IsUserExists.
type IsUserExistsRequest struct {
	UserID string `json:"user_id"`
}

// Validate method implements Request interface.
func (r IsUserExistsRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// CreateUserRequest is input type for CreateUser.
type CreateUserRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

// Validate implementes Request interface.
func (r CreateUserRequest) Validate() error {
	switch {
	case r.UserID == "":
		return errors.New("user_id is required")
	case r.Password == "":
		return errors.New("password is required")
	}
	return nil
}

// TOTPQRCodeRequest is input type for TOTPQRCode.
type TOTPQRCodeRequest struct {
	SessionID string `json:"-"`
}

// Validate implements Request interface.
func (r TOTPQRCodeRequest) Validate() error {
	if r.SessionID == "" {
		return errors.New("authorization header is required")
	}
	return nil
}

// SetSessionID implements SessionRequest interface.
func (r *TOTPQRCodeRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}

// VerifyTOTPRequest is input type for VerifyTOTP.
type VerifyTOTPRequest struct {
	Token     string `json:"token"`
	SessionID string `json:"-"`
}

// Validate implements Request interface.
func (r VerifyTOTPRequest) Validate() error {
	switch {
	case r.Token == "":
		return errors.New("token is required")
	case r.SessionID == "":
		return errors.New("authorization header is required")
	}
	return nil
}

// SetSessionID implements SessionRequest interface.
func (r *VerifyTOTPRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}

// UpdateEmailRequest is input type for UpdateEmail.
type UpdateEmailRequest struct {
	Email     string `json:"email"`
	SessionID string `json:"-"`
}

// Validate implements Request interface.
func (r UpdateEmailRequest) Validate() error {
	switch {
	case r.Email == "":
		return errors.New("email is required")
	case r.SessionID == "":
		return errors.New("authorization header is required")
	}
	return nil
}

// SetSessionID implements SessionRequest interface.
func (r *UpdateEmailRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}

// VerifyEmailRequest is input type for VerifyEmail.
type VerifyEmailRequest struct {
	SessionID string `json:"-"`
}

// Validate implements Request interface.
func (r VerifyEmailRequest) Validate() error {
	if r.SessionID == "" {
		return errors.New("session ID is required")
	}
	return nil
}

// SetPathArgs implements PathArgsRequest interface.
func (r *VerifyEmailRequest) SetPathArgs(args map[string]string) {
	sessid := args["sessid"]
	r.SessionID = sessid
}

// AuthByTOTPRequest is input type for AuthByTOTP.
type AuthByTOTPRequest struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// Validate implements Request interface.
func (r AuthByTOTPRequest) Validate() error {
	switch {
	case r.UserID == "":
		return errors.New("user_id is required")
	case r.Token == "":
		return errors.New("token is required")
	case len(r.Token) != 6:
		return errors.New("token length invalid")
	}
	for _, r := range r.Token {
		if !unicode.IsDigit(r) {
			return errors.New("token must be digit")
		}
	}
	return nil
}

// AuthByPasswordRequest is input type for AuthByPassword.
type AuthByPasswordRequest struct {
	SessionID string `json:"-"`
	Password  string `json:"password"`
}

// Validate implements Request interface.
func (r AuthByPasswordRequest) Validate() error {
	switch {
	case r.SessionID == "":
		return errors.New("authorization header is required")
	case r.Password == "":
		return errors.New("password is required")
	}
	return nil
}

// SetSessionID implements SessionRequest interface.
func (r *AuthByPasswordRequest) SetSessionID(sessid string) {
	r.SessionID = sessid
}
