package auth

import (
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/dchest/captcha"
	"github.com/eudore/eudore"
	"golang.org/x/crypto/scrypt"
)

var salf = "9999"

type (
	captchaSigned struct {
		Verify  string `json:"verify"`
		Expires int64  `json:"expires"`
	}
)

var letterBytes = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
var captchaSecret = []byte("secret")

func getCaptcha(ctx eudore.Context) {
	h := ctx.Response().Header()
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")

	var data = make([]byte, 6)
	for i := 0; i < 6; i++ {
		data[i] = letterBytes[rand.Intn(10)]
	}

	val, _ := scrypt.Key(data, captchaSecret, 16384, 8, 1, 32)
	h.Set("Captcha", tokenVerify.Signed(&captchaSigned{
		Verify:  hex.EncodeToString(val),
		Expires: time.Now().Add(5 * time.Minute).Unix(),
	}))

	img := captcha.NewImage(salf, data, 120, 40)
	// 兼容base64图片
	if ctx.GetHeader("Accept") == "application/base64" {
		h.Set("Content-Type", "application/base64")
		ctx.WriteString("data:image/png;base64,")
		encoder := base64.NewEncoder(base64.StdEncoding, ctx)
		img.WriteTo(encoder)
		encoder.Close()
		return
	}
	h.Set("Content-Type", "image/png")
	img.WriteTo(ctx)
}

func (cs captchaSigned) CheckVerify(key string) bool {
	var data = make([]byte, 6)
	for i := 0; i < 6; i++ {
		data[i] = key[i] - '0'
	}

	val, _ := scrypt.Key(data, captchaSecret, 16384, 8, 1, 32)
	return hex.EncodeToString(val) == cs.Verify
}
