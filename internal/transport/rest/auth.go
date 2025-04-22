package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"hippo/internal/domain"
	"hippo/internal/service"
)

func (h *Handler) handleSignUp(w http.ResponseWriter, r *http.Request) {
	const op = "handleSignUp"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var sInfo domain.SignUpInfo
	if err := json.NewDecoder(r.Body).Decode(&sInfo); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_request_body",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
		return
	}
	defer r.Body.Close()

	if err := sInfo.Validate(); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_request_data",
			Message: "Invalid sign-up data",
			Details: err.Error(),
		})
		return
	}

	id, err := h.usersService.SignUp(ctx, sInfo)
	if err != nil {
		h.logError(op, err)

		var duplicateEmail *service.ErrDuplicateEmail
		if errors.As(err, &duplicateEmail) {
			h.respondWithJSON(w, http.StatusConflict, op, ErrorResponse{
				Code:    "duplicate_email",
				Message: "User with provided email already exists.",
			})
			return
		}

		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to create user",
		})
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/auth/users/%d", id))
	h.respondWithJSON(w, http.StatusCreated, op, map[string]string{
		"message": "User created successfully",
	})
}

func (h *Handler) handleSignIn(w http.ResponseWriter, r *http.Request) {
	const op = "handleSignIn"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var sInfo domain.SignInInfo
	if err := json.NewDecoder(r.Body).Decode(&sInfo); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_request_body",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
		return
	}
	defer r.Body.Close()

	if err := sInfo.Validate(); err != nil {
		h.logError(op, err)
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_request_data",
			Message: "Invalid sign-in data",
			Details: err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := h.usersService.SignIn(ctx, sInfo)
	if err != nil {
		var invalidCred *service.ErrInvalidCredential
		if errors.As(err, &invalidCred) {
			h.respondWithJSON(w, http.StatusUnauthorized, op, ErrorResponse{
				Code:    "invalid_credential",
				Message: "Invalid credentials provided",
			})
			return
		}

		h.logError(op, err)
		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to sign in",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	h.respondWithJSON(w, http.StatusOK, op, map[string]string{
		"access_token": accessToken,
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	const op = "handleRefresh"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		h.logError(op, err)
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "missing_cookie",
			Message: "Refresh token cookie is missing",
		})
		return
	}

	accessToken, refreshToken, err := h.usersService.RefreshToken(ctx, cookie.Value)
	if err != nil {
		h.logError(op, err)

		var expired *service.ErrRefreshTokenExpired
		if errors.As(err, &expired) {
			h.respondWithJSON(w, http.StatusUnauthorized, op, ErrorResponse{
				Code:    "refresh_token_expired",
				Message: "Refresh token has expired. Please log in again.",
			})
			return
		}

		h.respondWithJSON(w, http.StatusUnauthorized, op, ErrorResponse{
			Code:    "invalid_refresh_token",
			Message: "Could not refresh token",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	h.respondWithJSON(w, http.StatusOK, op, map[string]string{
		"access_token": accessToken,
	})
}
