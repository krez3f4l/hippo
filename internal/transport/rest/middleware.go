package rest

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type CtxKey int

const (
	ctxUserIDKey CtxKey = iota
	ctxUserTokenKey
)

type responseWriter struct {
	http.ResponseWriter
	status        int
	headerWritten bool
	bodyWritten   bool
	mu            sync.Mutex
}

func newRespWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if !rw.headerWritten {
		rw.status = code
		rw.ResponseWriter.WriteHeader(code)
		rw.headerWritten = true
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if !rw.bodyWritten {
		rw.bodyWritten = true
		return rw.ResponseWriter.Write(b)
	} else {
		return 0, http.ErrHandlerTimeout
	}
}

func (h *Handler) timeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "timeoutMiddleware"

		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		defer cancel()

		done := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		ww, ok := w.(*responseWriter)
		if !ok {
			ww = newRespWriter(w)
		}

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			next.ServeHTTP(ww, r.WithContext(ctx))
			close(done)
		}()

		select {
		case <-ctx.Done():
			ww.mu.Lock()
			if !ww.headerWritten {
				w.WriteHeader(http.StatusGatewayTimeout)
				_, _ = w.Write([]byte(`{"error": "request timeout"}`))
				ww.bodyWritten = true

				h.log.WithFields(map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
				}).Warn("request timeout")
			}
			ww.mu.Unlock()

		case <-done:
			return

		case p := <-panicChan:
			panic(p)
		}
	})
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "loggingMiddleware"

		start := time.Now()

		ww, ok := w.(*responseWriter)
		if !ok {
			ww = newRespWriter(w)
		}

		defer func() {
			h.log.WithFields(map[string]interface{}{
				"method":   r.Method,
				"path":     r.URL.Path,
				"status":   ww.status,
				"duration": time.Since(start).String(),
				"ip":       r.RemoteAddr,
			}).Info("request completed")
		}()

		next.ServeHTTP(ww, r)
	})
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "authMiddleware"

		// jwt
		token, err := getTokenFromRequest(r)
		if err != nil {
			h.logError(op, err)
			h.respondWithJSON(w, http.StatusUnauthorized, op, map[string]string{
				"error":   "authentication required",
				"details": err.Error(),
			})
			return
		}

		userID, err := h.usersService.ParseToken(r.Context(), token)
		if err != nil {
			h.logError(op, err)
			h.respondWithJSON(w, http.StatusUnauthorized, op, map[string]string{
				"error": "invalid authentication token",
			})
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
		ctx = context.WithValue(ctx, ctxUserTokenKey, maskToken(token))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
