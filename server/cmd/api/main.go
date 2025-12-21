package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type appConfig struct {
	apiPort     string
	databaseURL string
	jwtSecret   string
}

type app struct {
	cfg appConfig
	db  *pgxpool.Pool
}

type healthResponse struct {
	Status string `json:"status"`
}

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

type pendingUserListItem struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

type userListItem struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Approved  bool   `json:"approved"`
	CreatedAt string `json:"createdAt"`
}

type userApprovalRequest struct {
	Approved bool `json:"approved"`
}

type tokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type deviceListItem struct {
	Id              int64  `json:"id"`
	VendorName      string `json:"vendorName"`
	ModelName       string `json:"modelName"`
	LocationName    string `json:"locationName"`
	SerialNumber    string `json:"serialNumber"`
	InventoryNumber string `json:"inventoryNumber"`
	Status          string `json:"status"`
	InstalledAt     string `json:"installedAt"`
}

type deviceUpsertRequest struct {
	ModelId         int64  `json:"modelId"`
	LocationId      *int64 `json:"locationId"`
	SerialNumber    string `json:"serialNumber"`
	InventoryNumber string `json:"inventoryNumber"`
	Status          string `json:"status"`
	InstalledAt     string `json:"installedAt"`
	Description     string `json:"description"`
}

type deviceUpsertResponse struct {
	Id int64 `json:"id"`
}

type idResponse struct {
	Id int64 `json:"id"`
}

type vendorListItem struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

type vendorUpsertRequest struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

type modelUpsertRequest struct {
	VendorId   int64  `json:"vendorId"`
	Name       string `json:"name"`
	DeviceType string `json:"deviceType"`
}

type locationUpsertRequest struct {
	Name string `json:"name"`
	Note string `json:"note"`
}

type deviceDetailsResponse struct {
	Id              int64  `json:"id"`
	ModelId         int64  `json:"modelId"`
	LocationId      *int64 `json:"locationId"`
	SerialNumber    string `json:"serialNumber"`
	InventoryNumber string `json:"inventoryNumber"`
	Status          string `json:"status"`
	InstalledAt     string `json:"installedAt"`
	Description     string `json:"description"`
}

func main() {
	cfg := appConfig{
		apiPort:     getEnv("API_PORT", "8080"),
		databaseURL: os.Getenv("DATABASE_URL"),
		jwtSecret:   getEnv("JWT_SECRET", "dev-secret"),
	}
	if cfg.databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	application := &app{cfg: cfg, db: db}
	if v := strings.ToLower(strings.TrimSpace(getEnv("SEED_DEMO", ""))); v == "1" || v == "true" || v == "yes" {
		application.seedIfEmpty(ctx)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", application.handleHealth)
	mux.HandleFunc("POST /auth/register", application.handleAuthRegister)
	mux.HandleFunc("POST /auth/login", application.handleAuthLogin)

	// Protected endpoints
	mux.HandleFunc("GET /vendors", application.requireAuth(application.handleVendorsList))
	mux.HandleFunc("POST /vendors", application.requireAuth(application.handleVendorsCreate))
	mux.HandleFunc("PUT /vendors/{id}", application.requireAuth(application.handleVendorsUpdate))
	mux.HandleFunc("DELETE /vendors/{id}", application.requireAuth(application.handleVendorsDelete))

	mux.HandleFunc("GET /models", application.requireAuth(application.handleModelsList))
	mux.HandleFunc("POST /models", application.requireAuth(application.handleModelsCreate))
	mux.HandleFunc("PUT /models/{id}", application.requireAuth(application.handleModelsUpdate))
	mux.HandleFunc("DELETE /models/{id}", application.requireAuth(application.handleModelsDelete))

	mux.HandleFunc("GET /locations", application.requireAuth(application.handleLocationsList))
	mux.HandleFunc("POST /locations", application.requireAuth(application.handleLocationsCreate))
	mux.HandleFunc("PUT /locations/{id}", application.requireAuth(application.handleLocationsUpdate))
	mux.HandleFunc("DELETE /locations/{id}", application.requireAuth(application.handleLocationsDelete))

	mux.HandleFunc("GET /devices", application.requireAuth(application.handleDevicesList))
	mux.HandleFunc("GET /devices/{id}", application.requireAuth(application.handleDevicesGet))
	mux.HandleFunc("POST /devices", application.requireAuth(application.handleDevicesCreate))
	mux.HandleFunc("PUT /devices/{id}", application.requireAuth(application.handleDevicesUpdate))
	mux.HandleFunc("DELETE /devices/{id}", application.requireAuth(application.handleDevicesDelete))

	// Admin: approve users
	mux.HandleFunc("GET /users/pending", application.requireAuth(application.handleUsersPendingList))
	mux.HandleFunc("POST /users/{id}/approve", application.requireAuth(application.handleUsersApprove))
	// Admin: manage users
	mux.HandleFunc("GET /users", application.requireAuth(application.handleUsersList))
	mux.HandleFunc("PUT /users/{id}/approval", application.requireAuth(application.handleUsersSetApproval))
	mux.HandleFunc("DELETE /users/{id}", application.requireAuth(application.handleUsersDelete))

	srv := &http.Server{
		Addr:              ":" + cfg.apiPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", cfg.apiPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Print("api stopped")
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := a.db.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_unavailable"})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
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
	var approved bool
	err = a.db.QueryRow(
		r.Context(),
		"INSERT INTO users(username, password_hash, role, approved) VALUES($1, $2, CASE WHEN (SELECT COUNT(*) FROM users WHERE role = 'admin') = 0 THEN 'admin' ELSE 'user' END, CASE WHEN (SELECT COUNT(*) FROM users WHERE role = 'admin') = 0 THEN TRUE ELSE FALSE END) RETURNING role, approved",
		username,
		string(passwordHash),
	).Scan(&role, &approved)
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

	if !approved && role != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "account_pending_approval"})
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
	var approved bool
	err := a.db.QueryRow(
		r.Context(),
		"SELECT password_hash, role, approved FROM users WHERE username = $1",
		username,
	).Scan(&passwordHash, &role, &approved)
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

	if !approved && role != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "account_pending_approval"})
		return
	}

	token, err := a.issueToken(username, role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "token_issue_failed"})
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, Username: username, Role: role})
}

func (a *app) handleUsersPendingList(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	rows, err := a.db.Query(r.Context(), "SELECT id, username, role, created_at FROM users WHERE approved = FALSE ORDER BY created_at")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	defer rows.Close()

	items := make([]pendingUserListItem, 0)
	for rows.Next() {
		var it pendingUserListItem
		var createdAt time.Time
		if err := rows.Scan(&it.Id, &it.Username, &it.Role, &createdAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
			return
		}
		it.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, it)
	}
	if rows.Err() != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (a *app) handleUsersApprove(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "UPDATE users SET approved = TRUE WHERE id = $1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (a *app) handleUsersList(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	rows, err := a.db.Query(r.Context(), "SELECT id, username, role, approved, created_at FROM users ORDER BY created_at")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	defer rows.Close()

	items := make([]userListItem, 0)
	for rows.Next() {
		var it userListItem
		var createdAt time.Time
		if err := rows.Scan(&it.Id, &it.Username, &it.Role, &it.Approved, &createdAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
			return
		}
		it.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, it)
	}
	if rows.Err() != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (a *app) handleUsersSetApproval(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var req userApprovalRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}

	// Safety: don't disable admin accounts.
	var targetRole string
	err = a.db.QueryRow(r.Context(), "SELECT role FROM users WHERE id = $1", id).Scan(&targetRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if targetRole == "admin" && !req.Approved {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "cannot_disable_admin"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "UPDATE users SET approved = $2 WHERE id = $1", id, req.Approved)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "approved": req.Approved})
}

func (a *app) handleUsersDelete(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var targetUsername string
	var targetRole string
	err = a.db.QueryRow(r.Context(), "SELECT username, role FROM users WHERE id = $1", id).Scan(&targetUsername, &targetRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	if targetUsername == authUsername(r.Context()) {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "cannot_delete_self"})
		return
	}
	if targetRole == "admin" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "cannot_delete_admin"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
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
	username = strings.TrimSpace(username)
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

func (a *app) handleVendorsList(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "SELECT id, name, COALESCE(country, '') FROM vendors ORDER BY name")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	defer rows.Close()

	items := make([]vendorListItem, 0)
	for rows.Next() {
		var it vendorListItem
		if err := rows.Scan(&it.Id, &it.Name, &it.Country); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
			return
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (a *app) handleVendorsCreate(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	var req vendorUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name_required"})
		return
	}

	var id int64
	err := a.db.QueryRow(
		r.Context(),
		"INSERT INTO vendors(name, country) VALUES($1, $2) RETURNING id",
		name,
		nullIfEmpty(req.Country),
	).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusCreated, idResponse{Id: id})
}

func (a *app) handleVendorsUpdate(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var req vendorUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name_required"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "UPDATE vendors SET name=$1, country=$2 WHERE id=$3", name, nullIfEmpty(req.Country), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, idResponse{Id: id})
}

func (a *app) handleVendorsDelete(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "DELETE FROM vendors WHERE id=$1", id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				writeJSON(w, http.StatusConflict, apiError{Error: "in_use"})
				return
			}
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *app) handleModelsCreate(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	var req modelUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}
	if req.VendorId <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "vendor_required"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name_required"})
		return
	}

	var id int64
	err := a.db.QueryRow(
		r.Context(),
		"INSERT INTO models(vendor_id, name, device_type) VALUES($1, $2, $3) RETURNING id",
		req.VendorId,
		name,
		nullIfEmpty(req.DeviceType),
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				writeJSON(w, http.StatusBadRequest, apiError{Error: "vendor_not_found"})
				return
			}
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusCreated, idResponse{Id: id})
}

func (a *app) handleModelsUpdate(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var req modelUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}
	if req.VendorId <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "vendor_required"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name_required"})
		return
	}

	cmd, err := a.db.Exec(
		r.Context(),
		"UPDATE models SET vendor_id=$1, name=$2, device_type=$3 WHERE id=$4",
		req.VendorId,
		name,
		nullIfEmpty(req.DeviceType),
		id,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				writeJSON(w, http.StatusBadRequest, apiError{Error: "vendor_not_found"})
				return
			}
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, idResponse{Id: id})
}

func (a *app) handleModelsDelete(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "DELETE FROM models WHERE id=$1", id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				writeJSON(w, http.StatusConflict, apiError{Error: "in_use"})
				return
			}
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *app) handleLocationsCreate(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	var req locationUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name_required"})
		return
	}

	var id int64
	err := a.db.QueryRow(r.Context(), "INSERT INTO locations(name, note) VALUES($1, $2) RETURNING id", name, nullIfEmpty(req.Note)).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusCreated, idResponse{Id: id})
}

func (a *app) handleLocationsUpdate(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var req locationUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name_required"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "UPDATE locations SET name=$1, note=$2 WHERE id=$3", name, nullIfEmpty(req.Note), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, idResponse{Id: id})
}

func (a *app) handleLocationsDelete(w http.ResponseWriter, r *http.Request) {
	if authRole(r.Context()) != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "DELETE FROM locations WHERE id=$1", id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				writeJSON(w, http.StatusConflict, apiError{Error: "in_use"})
				return
			}
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *app) handleDevicesList(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	query := "SELECT d.id, v.name, m.name, COALESCE(l.name, ''), COALESCE(d.serial_number, ''), COALESCE(d.inventory_number, ''), d.status, COALESCE(to_char(d.installed_at, 'YYYY-MM-DD'), '') " +
		"FROM devices d " +
		"JOIN models m ON m.id = d.model_id " +
		"JOIN vendors v ON v.id = m.vendor_id " +
		"LEFT JOIN locations l ON l.id = d.location_id "

	args := []any{}
	if q != "" {
		query += "WHERE (d.serial_number ILIKE '%' || $1 || '%' OR d.inventory_number ILIKE '%' || $1 || '%' OR m.name ILIKE '%' || $1 || '%' OR v.name ILIKE '%' || $1 || '%' OR d.status ILIKE '%' || $1 || '%') "
		args = append(args, q)
	}
	query += "ORDER BY d.id DESC"

	rows, err := a.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	defer rows.Close()

	items := make([]deviceListItem, 0)
	for rows.Next() {
		var it deviceListItem
		if err := rows.Scan(&it.Id, &it.VendorName, &it.ModelName, &it.LocationName, &it.SerialNumber, &it.InventoryNumber, &it.Status, &it.InstalledAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
			return
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (a *app) handleDevicesCreate(w http.ResponseWriter, r *http.Request) {
	var req deviceUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}

	if req.ModelId <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "model_required"})
		return
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	installedAt, err := parseDateYYYYMMDD(req.InstalledAt)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_installed_at"})
		return
	}

	var id int64
	err = a.db.QueryRow(
		r.Context(),
		"INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		req.ModelId,
		req.LocationId,
		nullIfEmpty(req.SerialNumber),
		nullIfEmpty(req.InventoryNumber),
		status,
		installedAt,
		nullIfEmpty(req.Description),
	).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusCreated, deviceUpsertResponse{Id: id})
}

func (a *app) handleDevicesGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var resp deviceDetailsResponse
	err = a.db.QueryRow(
		r.Context(),
		"SELECT id, model_id, location_id, COALESCE(serial_number, ''), COALESCE(inventory_number, ''), status, COALESCE(to_char(installed_at, 'YYYY-MM-DD'), ''), COALESCE(description, '') FROM devices WHERE id=$1",
		id,
	).Scan(&resp.Id, &resp.ModelId, &resp.LocationId, &resp.SerialNumber, &resp.InventoryNumber, &resp.Status, &resp.InstalledAt, &resp.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (a *app) handleDevicesUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	var req deviceUpsertRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_json"})
		return
	}

	if req.ModelId <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "model_required"})
		return
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	installedAt, err := parseDateYYYYMMDD(req.InstalledAt)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_installed_at"})
		return
	}

	cmd, err := a.db.Exec(
		r.Context(),
		"UPDATE devices SET model_id=$1, location_id=$2, serial_number=$3, inventory_number=$4, status=$5, installed_at=$6, description=$7 WHERE id=$8",
		req.ModelId,
		req.LocationId,
		nullIfEmpty(req.SerialNumber),
		nullIfEmpty(req.InventoryNumber),
		status,
		installedAt,
		nullIfEmpty(req.Description),
		id,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, deviceUpsertResponse{Id: id})
}

func (a *app) handleDevicesDelete(w http.ResponseWriter, r *http.Request) {
	role := authRole(r.Context())
	if role != "admin" {
		writeJSON(w, http.StatusForbidden, apiError{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_id"})
		return
	}

	cmd, err := a.db.Exec(r.Context(), "DELETE FROM devices WHERE id=$1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func parseDateYYYYMMDD(v string) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func nullIfEmpty(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
