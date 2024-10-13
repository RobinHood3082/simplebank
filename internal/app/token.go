package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(w http.ResponseWriter, r *http.Request) {
	var req renewAccessTokenRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		server.writeError(w, http.StatusUnauthorized, err)
		return
	}

	session, err := server.store.GetSession(r.Context(), pgtype.UUID{Bytes: refreshPayload.ID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			server.writeError(w, http.StatusNotFound, err)
			return
		}

		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	if session.IsBlocked {
		server.writeError(w, http.StatusUnauthorized, fmt.Errorf("session is blocked"))
		return
	}

	if session.Username != refreshPayload.Username {
		server.writeError(w, http.StatusUnauthorized, fmt.Errorf("refresh token is not valid for this user"))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		server.writeError(w, http.StatusUnauthorized, fmt.Errorf("mismatched refresh token"))
		return
	}

	if time.Now().After(session.ExpiresAt.Time) {
		server.writeError(w, http.StatusUnauthorized, fmt.Errorf("session has expired"))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		refreshPayload.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	err = server.writeJSON(w, http.StatusOK, rsp, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}
