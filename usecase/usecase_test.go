package usecase_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	totp "github.com/nasa9084/go-totp"
	"github.com/nasa9084/ident/infra"
	"github.com/nasa9084/ident/usecase"
	"github.com/nasa9084/ident/usecase/input"
	"github.com/nasa9084/ident/usecase/output"
)

func getEnv(t *testing.T) *infra.Environment {
	rdb, err := infra.OpenMySQL("localhost:3306", "root", "", "ident")
	if err != nil {
		t.Fatal(err)
	}
	kvs, err := infra.OpenRedis("localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	key, err := infra.LoadPrivateKey(os.Getenv("TEST_KEYPATH"))
	if err != nil {
		t.Fatal(err)
	}
	env := &infra.Environment{
		RDB:        rdb,
		KVS:        kvs,
		PrivateKey: key,
	}
	return env
}

const (
	mockPassword = "password"
	mockEmail    = "email"
	aliceID      = "alice"
)

func TestUserCreationProcess(t *testing.T) {
	env := getEnv(t)

	// before create alice
	euReq := input.ExistsUserRequest{UserID: aliceID}
	euResp := usecase.ExistsUser(context.Background(), euReq, env).(output.ExistsUserResponse)
	if euResp.Exists {
		t.Error("alice should not exists")
		return
	}
	// create alice
	cReq := input.CreateUserRequest{UserID: aliceID, Password: mockPassword}
	cResp := usecase.CreateUser(context.Background(), cReq, env).(output.CreateUserResponse)
	if cResp.Err != nil {
		t.Error(cResp.Err)
		return
	}
	if cResp.Status != http.StatusCreated {
		t.Errorf("%d != %d", cResp.Status, http.StatusCreated)
		return
	}

	// get secret
	secret, err := redis.String(env.KVS.Do("HGET", "user:"+aliceID, "totp_secret"))
	if err != nil {
		t.Error(err)
		return
	}

	g := totp.New(secret)
	// verify TOTP
	vtReq := input.VerifyTOTPRequest{Token: g.GenerateString(), SessionID: cResp.SessionID}
	vtResp := usecase.VerifyTOTP(context.Background(), vtReq, env).(output.VerifyTOTPResponse)

	if vtResp.Status != http.StatusOK {
		t.Errorf("%d != %d", vtResp.Status, http.StatusOK)
		return
	}

	umReq := input.UpdateEmailRequest{Email: mockEmail, SessionID: cResp.SessionID}
	umResp := usecase.UpdateEmail(context.Background(), umReq, env).(output.UpdateEmailResponse)

	if umResp.Status != http.StatusOK {
		t.Errorf("%d != %d", umResp.Status, http.StatusOK)
		return
	}

	keys, err := redis.Strings(env.KVS.Do("KEYS", "session:*"))
	if err != nil {
		t.Fatal(err)
	}
	var sessid string
	for _, key := range keys {
		val, err := redis.String(env.KVS.Do("GET", key))
		if err != nil {
			t.Fatal(err)
		}
		if val == aliceID {
			sessid = strings.Split(key, ":")[1]
		}
	}

	vmReq := input.VerifyEmailRequest{SessionID: sessid}
	vmResp := usecase.VerifyEmail(context.Background(), vmReq, env).(output.VerifyEmailResponse)

	if vmResp.Status != http.StatusOK {
		t.Errorf("%d != %d", vmResp.Status, http.StatusOK)
		return
	}

	atReq := input.AuthByTOTPRequest{UserID: aliceID, Token: g.GenerateString()}
	atResp := usecase.AuthByTOTP(context.Background(), atReq, env).(output.AuthByTOTPResponse)
	if atResp.Status != http.StatusOK {
		t.Errorf("%d != %d", atResp.Status, http.StatusOK)
		return
	}

	apReq := input.AuthByPasswordRequest{SessionID: atResp.SessionID, Password: mockPassword}
	apResp := usecase.AuthByPassword(context.Background(), apReq, env).(output.AuthByPasswordResponse)
	if apResp.Status != http.StatusOK {
		t.Errorf("%d != %d", apResp.Status, http.StatusOK)
		return
	}

	pkResp := usecase.GetPublicKey(context.Background(), env).(output.GetPublicKeyResponse)
	if pkResp.Status != http.StatusOK {
		t.Errorf("%d != %d", pkResp.Status, http.StatusOK)
		return
	}
	pubKey, err := jwt.ParseECPublicKeyFromPEM(pkResp.PublicKeyPEM)
	if err != nil {
		t.Error(err)
		return
	}
	tk, err := jwt.Parse(apResp.Token, func(*jwt.Token) (interface{}, error) { return pubKey, nil })
	if err != nil {
		t.Error(err)
		return
	}
	if !tk.Valid {
		t.Error("token appears not valid")
		return
	}
}
