package app

import (
	"net/http"
	"time"

	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	Username string `json:"username" validate:"required,alphanum"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type userResponse struct {
	Username          string             `json:"username"`
	FullName          string             `json:"full_name"`
	Email             string             `json:"email"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at"`
	CreatedAt         pgtype.Timestamptz `json:"created_at"`
}

func newUserResponse(user persistence.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	arg := persistence.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(r.Context(), arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				server.writeError(w, http.StatusForbidden, err)
				return
			}
		}

		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	rsp := newUserResponse(user)

	server.logger.Info("User created", "User", rsp)
	err = server.writeJSON(w, http.StatusOK, rsp, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}

type loginUserRequest struct {
	Username string `json:"username" validate:"required,alphanum"`
	Password string `json:"password" validate:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(w http.ResponseWriter, r *http.Request) {
	var req loginUserRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	user, err := server.store.GetUser(r.Context(), req.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			server.writeError(w, http.StatusNotFound, err)
			return
		}

		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		server.writeError(w, http.StatusUnauthorized, err)
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	session, err := server.store.CreateSession(r.Context(), persistence.CreateSessionParams{
		ID:           pgtype.UUID{Bytes: refreshPayload.ID, Valid: true},
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    r.UserAgent(),
		ClientIp:     r.RemoteAddr,
		IsBlocked:    false,
		ExpiresAt:    pgtype.Timestamptz{Time: refreshPayload.ExpiredAt, Valid: true},
	})
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID.Bytes,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	server.logger.Info("User logged in", "User", user)
	err = server.writeJSON(w, http.StatusOK, rsp, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}
