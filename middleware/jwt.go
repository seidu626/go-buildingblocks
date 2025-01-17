package middleware

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"strings"
)

// JWTMiddleware validates the JWT token
func JWTMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		tokenString := string(ctx.Request.Header.Peek("Authorization"))
		if tokenString == "" {
			ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Get the public key
			// For simplicity, skip key fetching in this example
			return []byte("your-public-key"), nil
		})

		if err != nil || !token.Valid {
			ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
			return
		}

		// Set user information in context
		ctx.SetUserValue("user", token.Claims.(jwt.MapClaims))
		next(ctx)
	}
}
