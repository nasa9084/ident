package usecase

import (
	"context"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	totp "github.com/nasa9084/go-totp"
	"github.com/nasa9084/ident/domain/repository"
	"github.com/nasa9084/ident/infra"
	"github.com/nasa9084/ident/usecase/input"
	"github.com/nasa9084/ident/usecase/output"
	qrcode "github.com/skip2/go-qrcode"
)

func statusFromError(err error) int {
	switch err.(type) {
	case redis.Error, *mysql.MySQLError:
		return http.StatusInternalServerError
	}
	switch err {
	case repository.ErrUserExists:
		return http.StatusConflict
	case redis.ErrNil:
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// ExistsUser returns given user id has been used or not.
func ExistsUser(ctx context.Context, req input.ExistsUserRequest, env *infra.Environment) output.Response {
	var resp output.ExistsUserResponse

	repo := env.GetUserRepository()
	exists, err := repo.ExistsUser(ctx, req.UserID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	resp.Status = http.StatusOK
	resp.Exists = exists
	return resp
}

// CreateUser creates a new user.
func CreateUser(ctx context.Context, req input.CreateUserRequest, env *infra.Environment) output.Response {
	var resp output.CreateUserResponse

	repo := env.GetUserRepository()
	sessid, err := repo.CreateUser(ctx, req.UserID, req.Password)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	resp.SessionID = sessid
	resp.Status = http.StatusCreated
	return resp
}

// TOTPQRCode returns a QR code including TOTP URI associated to given user.
func TOTPQRCode(ctx context.Context, req input.TOTPQRCodeRequest, env *infra.Environment) output.Response {
	var resp output.TOTPQRCodeResponse

	repo := env.GetUserRepository()
	u, err := repo.FindUserBySessionID(ctx, req.SessionID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	g := totp.New(u.TOTPSecret)
	png, err := qrcode.Encode(g.URI("ident", u.ID), qrcode.Medium, 256)
	if err != nil {
		resp.Err = err
		resp.Status = http.StatusInternalServerError
		return resp
	}
	resp.QRCode = png
	resp.Status = http.StatusOK
	return resp
}

// VerifyTOTP verifies the TOTP configuration is successfully done.
func VerifyTOTP(ctx context.Context, req input.VerifyTOTPRequest, env *infra.Environment) output.Response {
	var resp output.VerifyTOTPResponse

	repo := env.GetUserRepository()
	u, err := repo.FindUserBySessionID(ctx, req.SessionID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	g := totp.New(u.TOTPSecret)
	if g.GenerateString() != req.Token {
		resp.Err = errors.New("token invalid")
		resp.Status = http.StatusUnauthorized
		return resp
	}
	u.TOTPVerified = true
	if err := repo.UpdateUser(ctx, u); err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}

	resp.Status = http.StatusOK
	return resp
}

// UpdateEmail updates email for the user.
func UpdateEmail(ctx context.Context, req input.UpdateEmailRequest, env *infra.Environment) output.Response {
	var resp output.UpdateEmailResponse

	repo := env.GetUserRepository()
	u, err := repo.FindUserBySessionID(ctx, req.SessionID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	if !u.TOTPVerified {
		resp.Err = errors.New("TOTP verification has not done")
		resp.Status = http.StatusForbidden
		return resp
	}
	u.Email = req.Email
	if err := repo.UpdateUser(ctx, u); err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	sessid, err := repo.CreateSession(u)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	// Mail here
	if err := env.SendVerifyMail(env.MailFrom, u.Email, sessid); err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	resp.Status = http.StatusOK
	return resp
}

// VerifyEmail verifies the email is valid.
func VerifyEmail(ctx context.Context, req input.VerifyEmailRequest, env *infra.Environment) output.Response {
	var resp output.VerifyEmailResponse
	repo := env.GetUserRepository()
	u, err := repo.FindUserBySessionID(ctx, req.SessionID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	if err := repo.Verify(ctx, u); err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}

	resp.Status = http.StatusOK
	return resp
}

// AuthByTOTP authenticates using user ID and TOTP token.
// And returns SessionID.
func AuthByTOTP(ctx context.Context, req input.AuthByTOTPRequest, env *infra.Environment) output.Response {
	var resp output.AuthByTOTPResponse
	repo := env.GetUserRepository()
	u, err := repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	g := totp.New(u.TOTPSecret)
	if g.GenerateString() != req.Token {
		resp.Err = errors.New("token invalid")
		resp.Status = http.StatusUnauthorized
		return resp
	}

	sessid, err := repo.CreateSession(u)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	resp.SessionID = sessid
	resp.Status = http.StatusOK

	return resp
}

// AuthByPassword authenticates using password and session ID.
// Returns JWT Token.
func AuthByPassword(ctx context.Context, req input.AuthByPasswordRequest, env *infra.Environment) output.Response {
	var resp output.AuthByPasswordResponse
	repo := env.GetUserRepository()
	u, err := repo.FindUserBySessionID(ctx, req.SessionID)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}

	for i := 0; i < 30; i++ {
		h := sha512.Sum512([]byte(req.Password + u.ID))
		req.Password = hex.EncodeToString(h[:])
	}

	if u.Password != req.Password {
		resp.Err = errors.New("password invalid")
		resp.Status = http.StatusUnauthorized
		return resp
	}

	expire := time.Now().Add(1 * time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"exp":     expire,
		"user_id": u.ID,
	})
	signed, err := token.SignedString(env.PrivateKey)
	if err != nil {
		resp.Err = err
		resp.Status = statusFromError(err)
		return resp
	}
	resp.Token = signed
	resp.Status = http.StatusOK
	return resp
}

// GetPublicKey returns ECDSA public key.
func GetPublicKey(ctx context.Context, env *infra.Environment) output.Response {
	var resp output.GetPublicKeyResponse
	pubKey := env.PrivateKey.Public()
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		resp.Err = err
		resp.Status = http.StatusInternalServerError
	}
	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	resp.Status = http.StatusOK
	resp.PublicKeyPEM = pemKey

	return resp
}
