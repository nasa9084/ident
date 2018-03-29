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
	t.Run("IsUserExistsRequest", testIsUserExistsValidate)
	t.Run("CreateUserRequest", testCreateUserValidate)
	t.Run("TOTPQRCodeRequest", testTOTPQRCodeValidate)
	t.Run("VerifyTOTPRequest", testVerifyTOTPValidate)
	t.Run("UpdateEmailRequest", testUpdateEmailValidate)
	t.Run("VerifyEmailRequest", testVerifyEmailValidate)
	t.Run("AuthByTOTPRequest", testAuthByTOTPValidate)
	t.Run("AuthByPasswordReqeust", testAuthByPasswordValidate)
}

func testIsUserExistsValidate(t *testing.T) {
	candidates := []struct {
		request input.IsUserExistsRequest
		hasErr  bool
	}{
		{input.IsUserExistsRequest{UserID: "foo"}, false},
		{input.IsUserExistsRequest{}, true},
	}

	for _, c := range candidates {
		checkValidate(t, c.request, c.hasErr)
	}
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
		{input.VerifyTOTPRequest{Token: "foo", SessionID: "bar"}, false},
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
