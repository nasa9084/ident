package input_test

import (
	"testing"

	"github.com/nasa9084/ident/usecase/input"
)

func checkValidate(t *testing.T, r input.Request, hasErr bool) {
	t.Helper()
	if err := r.Validate(); hasErr != (err != nil) {
		t.Log(r)
		if !hasErr {
			t.Error(err)
			return
		}
		t.Error("error should be occurred, but not")
		return
	}
}

func TestRequestValidate(t *testing.T) {
	t.Run("CreateUserRequest", testCreateUserValidate)
	t.Run("TOTPQRCodeRequest", testTOTPQRCodeValidate)
	t.Run("VerifyTOTPRequest", testVerifyTOTPValidate)
	t.Run("UpdateEmailRequest", testUpdateEmailValidate)
	t.Run("VerifyEmailRequest", testVerifyEmailValidate)
	t.Run("AuthByTOTPRequest", testAuthByTOTPValidate)
	t.Run("AuthByPasswordReqeust", testAuthByPasswordValidate)
}

func testCreateUserValidate(t *testing.T) {
	candidates := []struct {
		request input.CreateUserRequest
		hasErr  bool
	}{
		{input.CreateUserRequest{UserID: "foo", Password: "bar"}, false},
		{input.CreateUserRequest{UserID: "foo"}, true},
		{input.CreateUserRequest{Password: "bar"}, true},
	}

	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

func testTOTPQRCodeValidate(t *testing.T) {
	candidates := []struct {
		request input.TOTPQRCodeRequest
		hasErr  bool
	}{
		{input.TOTPQRCodeRequest{SessionID: "foo"}, false},
		{input.TOTPQRCodeRequest{}, true},
	}
	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

func testVerifyTOTPValidate(t *testing.T) {
	candidates := []struct {
		request input.VerifyTOTPRequest
		hasErr  bool
	}{
		{input.VerifyTOTPRequest{Token: "000000", SessionID: "bar"}, false},
		{input.VerifyTOTPRequest{Token: "foo", SessionID: "bar"}, true},
		{input.VerifyTOTPRequest{Token: "foo"}, true},
		{input.VerifyTOTPRequest{SessionID: "bar"}, true},
	}
	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

func testUpdateEmailValidate(t *testing.T) {
	candidates := []struct {
		request input.UpdateEmailRequest
		hasErr  bool
	}{
		{input.UpdateEmailRequest{Email: "foo", SessionID: "bar"}, false},
		{input.UpdateEmailRequest{Email: "foo"}, true},
		{input.UpdateEmailRequest{SessionID: "bar"}, true},
	}
	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

func testVerifyEmailValidate(t *testing.T) {
	candidates := []struct {
		request input.VerifyEmailRequest
		hasErr  bool
	}{
		{input.VerifyEmailRequest{SessionID: "foo"}, false},
		{input.VerifyEmailRequest{}, true},
	}
	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

func testAuthByTOTPValidate(t *testing.T) {
	candidates := []struct {
		request input.AuthByTOTPRequest
		hasErr  bool
	}{
		{input.AuthByTOTPRequest{UserID: "foo", Token: "000000"}, false},
		{input.AuthByTOTPRequest{UserID: "foo", Token: "abcdef"}, true},
		{input.AuthByTOTPRequest{UserID: "foo", Token: "1"}, true},
		{input.AuthByTOTPRequest{UserID: "foo"}, true},
		{input.AuthByTOTPRequest{Token: "000000"}, true},
		{input.AuthByTOTPRequest{Token: "1"}, true},
		{input.AuthByTOTPRequest{}, true},
	}
	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

func testAuthByPasswordValidate(t *testing.T) {
	candidates := []struct {
		request input.AuthByPasswordRequest
		hasErr  bool
	}{
		{input.AuthByPasswordRequest{SessionID: "foo", Password: "bar"}, false},
		{input.AuthByPasswordRequest{SessionID: "foo"}, true},
		{input.AuthByPasswordRequest{Password: "bar"}, true},
		{input.AuthByPasswordRequest{}, true},
	}
	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
}

const sessid = "foobarbaz"

func TestSetSessionID(t *testing.T) {
	t.Run("TOTPQRCode", testTOTPQRCodeSetSessionID)
	t.Run("VerifyTOTP", testVerifyTOTPSetSessionID)
}

func testTOTPQRCodeSetSessionID(t *testing.T) {
	totpQRCode := input.TOTPQRCodeRequest{}
	totpQRCode.SetSessionID(sessid)
	if totpQRCode.SessionID != sessid {
		t.Errorf("%s != %s", totpQRCode.SessionID, sessid)
		return
	}
}

func testVerifyTOTPSetSessionID(t *testing.T) {
	verifyTOTP := input.VerifyTOTPRequest{}
	verifyTOTP.SetSessionID(sessid)
	if verifyTOTP.SessionID != sessid {
		t.Errorf("%s != %s", verifyTOTP.SessionID, sessid)
		return
	}
}
