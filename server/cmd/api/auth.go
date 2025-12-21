//go:build ignore
// +build ignore

package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type tokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (a *app) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}

	username := normalizeUsername(req.Username)
	password := req.Password

	if err := validateUsernamePassword(username, password); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "password_hash_failed"})
		return
	}

	var role string
	err = a.db.QueryRow(
		r.Context(),
		"INSERT INTO users(username, password_hash, role) VALUES($1, $2, 'user') RETURNING role",
		username,
		string(passwordHash),
	).Scan(&role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				writeJSON(w, http.StatusConflict, apiError{Error: "username_taken"})
				return
			}
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	token, err := a.issueToken(username, role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "token_issue_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token, Username: username, Role: role})
}

func (a *app) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}

	username := normalizeUsername(req.Username)
	password := req.Password
	if username == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "username_and_password_required"})
		return
	}

	var passwordHash string
	var role string
	err := a.db.QueryRow(
		r.Context(),
		"SELECT password_hash, role FROM users WHERE username = $1",
		username,
	).Scan(&passwordHash, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_credentials"})
		return
	}

	token, err := a.issueToken(username, role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "token_issue_failed"})
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, Username: username, Role: role})
}

func (a *app) issueToken(username, role string) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(a.cfg.jwtSecret))
}

func validateUsernamePassword(username, password string) error {
	if username == "" {
		return errors.New("username_required")
	}
	if len(username) < 3 || len(username) > 64 {
		return errors.New("username_length_invalid")
	}
	if len(password) < 8 || len(password) > 128 {
		return errors.New("password_length_invalid")
	}
	return nil
}


























































































































































}	return nil	}		return errors.New("password_length_invalid")	if len(password) < 8 || len(password) > 128 {	}		return errors.New("username_length_invalid")	if len(username) < 3 || len(username) > 64 {	}		return errors.New("username_required")	if username == "" {func validateUsernamePassword(username, password string) error {}	return t.SignedString([]byte(a.cfg.jwtSecret))	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)	}		},			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),			IssuedAt:  jwt.NewNumericDate(now),			Subject:   username,		RegisteredClaims: jwt.RegisteredClaims{		Role: role,	claims := tokenClaims{	now := time.Now()func (a *app) issueToken(username, role string) (string, error) {}	writeJSON(w, http.StatusOK, authResponse{Token: token, Username: username, Role: role})	}		return		writeJSON(w, http.StatusInternalServerError, apiError{Error: "token_issue_failed"})	if err != nil {	token, err := a.issueToken(username, role)	}		return		writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_credentials"})	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {	}		return		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})		}			return			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_credentials"})		if errors.Is(err, pgx.ErrNoRows) {	if err != nil {	).Scan(&passwordHash, &role)		username,		"SELECT password_hash, role FROM users WHERE username = $1",		r.Context(),	err := a.db.QueryRow(	var role string	var passwordHash string	}		return		writeJSON(w, http.StatusBadRequest, apiError{Error: "username_and_password_required"})	if username == "" || password == "" {	password := req.Password	username := normalizeUsername(req.Username)	}		return		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})	if err := readJSON(r, &req); err != nil {	var req loginRequestfunc (a *app) handleAuthLogin(w http.ResponseWriter, r *http.Request) {}	writeJSON(w, http.StatusCreated, authResponse{Token: token, Username: username, Role: role})	}		return		writeJSON(w, http.StatusInternalServerError, apiError{Error: "token_issue_failed"})	if err != nil {	token, err := a.issueToken(username, role)	}		return		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})		}			}				return				writeJSON(w, http.StatusConflict, apiError{Error: "username_taken"})			if pgErr.Code == "23505" {		if errors.As(err, &pgErr) {		var pgErr *pgconn.PgError	if err != nil {	).Scan(&role)		string(passwordHash),		username,		"INSERT INTO users(username, password_hash, role) VALUES($1, $2, 'user') RETURNING role",		r.Context(),	err = a.db.QueryRow(	var role string	}		return		writeJSON(w, http.StatusInternalServerError, apiError{Error: "password_hash_failed"})	if err != nil {	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)	}		return		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})	if err := validateUsernamePassword(username, password); err != nil {	password := req.Password	username := normalizeUsername(req.Username)	}		return		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})	if err := readJSON(r, &req); err != nil {	var req registerRequestfunc (a *app) handleAuthRegister(w http.ResponseWriter, r *http.Request) {}	jwt.RegisteredClaims	Role string `json:"role"`type tokenClaims struct {}	Role     string `json:"role"`	Username string `json:"username"`	Token    string `json:"token"`type authResponse struct {}	Password string `json:"password"`	Username string `json:"username"`type loginRequest struct {}	Password string `json:"password"`	Username string `json:"username"`type registerRequest struct {)	"golang.org/x/crypto/bcrypt"	"github.com/jackc/pgx/v5"	"github.com/jackc/pgconn"	"github.com/golang-jwt/jwt/v5"	"time"	"net/http"	"errors"import (