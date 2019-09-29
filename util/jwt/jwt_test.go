package jwt_test

import (
	"github.com/eudore/website/util/jwt"
	"testing"
	"time"
)

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjU0MDMyMDAsIm5hbWUiOiJyb290IiwidWlkIjoiMSJ9.rBBE_T4hwRmwhZOTEjWdhK-8s4Rqi6UqpcwisI1kjeI

func TestJwt(t *testing.T) {
	fn := jwt.NewVerifyHS256([]byte("secret"))
	t.Log(fn.Signed(map[string]interface{}{
		"name":    "root",
		"userid":  "1",
		"expires": time.Date(2022, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	}))
	t.Log(fn([]byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjU0MDMyMDAsInVpZCI6IjEifQ")))
}
