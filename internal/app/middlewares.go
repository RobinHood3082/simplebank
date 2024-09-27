package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// LogRequest logs all incoming requests
func (server *Server) LogRequest(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.logger.Info(fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI()))
		next.ServeHTTP(w, r)
	})
}

// RecoverPanic recovers from panics and returns a 500 Internal Server Error
func (server *Server) RecoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				server.writeError(w, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = AutorizationPayloadKey("authorization_payload")
)

// AutorizationPayloadKey is the key for the authorization payload in the request context
type AutorizationPayloadKey string

// Authenticate checks if the request is authenticated
func (server *Server) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.logger.Info("Authenticating request")
		authorizationHeader := r.Header.Get(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			server.writeError(w, http.StatusUnauthorized, errors.New("authorization header is missing"))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) != 2 {
			server.writeError(w, http.StatusUnauthorized, errors.New("authorization header is not in the format 'Bearer <token>'"))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			server.writeError(w, http.StatusUnauthorized, fmt.Errorf("unsupported authorization type %s", authorizationType))
			return
		}

		accessToken := fields[1]
		payload, err := server.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			server.writeError(w, http.StatusUnauthorized, err)
			return
		}

		ctx := context.WithValue(r.Context(), AuthorizationPayloadKey, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
