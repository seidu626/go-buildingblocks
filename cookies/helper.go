package cookies

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"strings"
)

// https://www.alexedwards.net/blog/working-with-cookies-in-go
var (
	ErrNoCookie     = errors.New("no cookie found") // Custom error for missing cookies
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func Write(ctx *fasthttp.RequestCtx, cookie fasthttp.Cookie) error {
	cookie.SetValue(base64.URLEncoding.EncodeToString([]byte(cookie.Value())))

	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	ctx.Response.Header.SetCookie(&cookie)

	return nil
}

func Read(ctx *fasthttp.RequestCtx, name string) (string, error) {
	cookieValue := ctx.Request.Header.Cookie(name)
	if cookieValue == nil {
		return "", ErrNoCookie
	}

	value, err := base64.URLEncoding.DecodeString(string(cookieValue))
	if err != nil {
		return "", ErrInvalidValue
	}

	return string(value), nil
}

func WriteSigned(ctx *fasthttp.RequestCtx, cookie fasthttp.Cookie, secretKey []byte) error {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Key()))
	mac.Write([]byte(cookie.Value()))
	signature := mac.Sum(nil)

	cookie.SetValue(string(signature) + string(cookie.Value()))

	return Write(ctx, cookie)
}

func ReadSigned(ctx *fasthttp.RequestCtx, name string, secretKey []byte) (string, error) {
	signedValue, err := Read(ctx, name)
	if err != nil {
		return "", err
	}

	if len(signedValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	return value, nil
}

func WriteEncrypted(ctx *fasthttp.RequestCtx, cookie fasthttp.Cookie, secretKey []byte) error {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return err
	}

	plaintext := fmt.Sprintf("%s:%s", cookie.Key(), cookie.Value())

	encryptedValue := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	cookie.SetValue(string(encryptedValue))

	return Write(ctx, cookie)
}

func ReadEncrypted(ctx *fasthttp.RequestCtx, name string, secretKey []byte) (string, error) {
	encryptedValue, err := Read(ctx, name)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()

	if len(encryptedValue) < nonceSize {
		return "", ErrInvalidValue
	}

	nonce := encryptedValue[:nonceSize]
	ciphertext := encryptedValue[nonceSize:]

	plaintext, err := aesGCM.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return "", ErrInvalidValue
	}

	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok {
		return "", ErrInvalidValue
	}

	if expectedName != name {
		return "", ErrInvalidValue
	}

	return value, nil
}
