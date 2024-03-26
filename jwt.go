package keycloakmiddleware

import (
	"encoding/json"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/valyala/fasthttp"
	"net/http"
	"strings"
	"time"
)

type middleware struct {
	wrapperCode int // 0: default, 1:standard, 2:traceable
}

func Construct(wrapperCode int) middleware {
	return middleware{wrapperCode: wrapperCode}
}

func (middleware *middleware) Validate(scopes []string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		var isEnabled = getEnvOrDefault("KEYCLOAK_JWT_ENABLED", "false").(string)
		if strings.ToLower(isEnabled) == "false" || isEnabled == "0" {
			return
		}

		authHeader := string(ctx.Request.Header.Peek("Authorization"))
		s := strings.SplitN(authHeader, " ", 2)
		if len(s) != 2 {
			msg := "Authorization token is not found"
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		headerToken := s[1]
		unverifiedToken, err := jwt.Parse([]byte(headerToken))
		if err != nil {
			msg := err.Error()
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		kid := unverifiedToken.Header().KeyID
		key, err := getPublicKey(kid)
		if err != nil {
			msg := err.Error()
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		verifier, err := jwt.NewVerifierRS(jwt.RS256, key)
		if err != nil {
			msg := err.Error()
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		token, err := jwt.ParseAndVerifyString(headerToken, verifier)
		if err != nil {
			msg := err.Error()
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		var claims claims
		errClaims := json.Unmarshal(token.RawClaims(), &claims)
		if errClaims != nil {
			msg := errClaims.Error()
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		var iss = getEnv("KEYCLOAK_JWT_ISS")
		if claims.Issuer != iss {
			msg := "Token issuer is not valid"
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		if claims.ExpiresAt.Unix() < time.Now().Unix() {
			msg := "Token expired"
			middleware.abort(http.StatusUnauthorized, ctx, msg)
			return
		}

		if !isScopesValid(claims, scopes) {
			msg := "Access to this endpoint is not allowed"
			middleware.abort(http.StatusForbidden, ctx, msg)
			return
		}

		ctx.SetUserValue("keycloak_username", claims.Username)
		ctx.SetUserValue("keycloak_name", claims.Name)
		ctx.SetUserValue("keycloak_email", claims.Email)

		next(ctx)
	})
}

func (middleware *middleware) abort(status int, ctx *fasthttp.RequestCtx, message interface{}) {
	httpStatus := http.StatusOK
	if middleware.wrapperCode != 0 {
		httpStatus = status
	}
	ctx.SetStatusCode(httpStatus)
	ctx.SetContentType("application/json")
	response := middleware.wrapper(httpStatus, ctx, message, nil)
	fmt.Fprintf(ctx, prettyPrint(response))
}
