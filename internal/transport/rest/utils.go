package rest

import (
	"errors"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func getTokenFromRequest(r *http.Request) (string, error) {
	const bearerPrefix = "Bearer "

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("authorization header format must be 'Bearer {token}'")
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", errors.New("token cannot be empty")
	}

	return token, nil
}

func maskToken(token string) string {
	if len(token) < 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
