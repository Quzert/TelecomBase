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

	"telecombase/server/internal/store"
)

type appConfig struct {
	apiPort     string
	databaseURL string
	jwtSecret   string
}

type app struct {
	cfg appConfig
	db  *pgxpool.Pool
	st  *store.Queries
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

	application := &app{cfg: cfg, db: db, st: store.New(db)}
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
	created, err := a.st.CreateUserAutoAdmin(r.Context(), username, string(passwordHash))
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
	role = created.Role
	approved = created.Approved

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
	got, err := a.st.GetUserAuthByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, apiError{Error: "invalid_credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	passwordHash = got.PasswordHash
	role = got.Role
	approved = got.Approved

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

	rows, err := a.st.ListPendingUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	items := make([]pendingUserListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, pendingUserListItem{
			Id:        row.ID,
			Username:  row.Username,
			Role:      row.Role,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
		})
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

	affected, err := a.st.ApproveUserByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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

	rows, err := a.st.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	items := make([]userListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, userListItem{
			Id:        row.ID,
			Username:  row.Username,
			Role:      row.Role,
			Approved:  row.Approved,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
		})
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
	targetRole, err = a.st.GetUserRoleByID(r.Context(), id)
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

	affected, err := a.st.SetUserApprovedByID(r.Context(), id, req.Approved)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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
	target, err := a.st.GetUserUsernameRoleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	targetUsername = target.Username
	targetRole = target.Role

	if targetUsername == authUsername(r.Context()) {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "cannot_delete_self"})
		return
	}
	if targetRole == "admin" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "cannot_delete_admin"})
		return
	}

	affected, err := a.st.DeleteUserByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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
	rows, err := a.st.ListVendors(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	items := make([]vendorListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, vendorListItem{Id: row.ID, Name: row.Name, Country: row.Country})
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
	id, err := a.st.CreateVendor(r.Context(), name, nullIfEmpty(req.Country))
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

	affected, err := a.st.UpdateVendor(r.Context(), id, name, nullIfEmpty(req.Country))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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

	affected, err := a.st.DeleteVendor(r.Context(), id)
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
	if affected == 0 {
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
	id, err := a.st.CreateModel(r.Context(), req.VendorId, name, nullIfEmpty(req.DeviceType))
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

	affected, err := a.st.UpdateModel(r.Context(), id, req.VendorId, name, nullIfEmpty(req.DeviceType))
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
	if affected == 0 {
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

	affected, err := a.st.DeleteModel(r.Context(), id)
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
	if affected == 0 {
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
	id, err := a.st.CreateLocation(r.Context(), name, nullIfEmpty(req.Note))
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

	affected, err := a.st.UpdateLocation(r.Context(), id, name, nullIfEmpty(req.Note))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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

	affected, err := a.st.DeleteLocation(r.Context(), id)
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
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *app) handleDevicesList(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := a.st.ListDevices(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	items := make([]deviceListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, deviceListItem{
			Id:              row.ID,
			VendorName:      row.VendorName,
			ModelName:       row.ModelName,
			LocationName:    row.LocationName,
			SerialNumber:    row.SerialNumber,
			InventoryNumber: row.InventoryNumber,
			Status:          row.Status,
			InstalledAt:     row.InstalledAt,
		})
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
	id, err = a.st.CreateDevice(
		r.Context(),
		req.ModelId,
		req.LocationId,
		nullIfEmpty(req.SerialNumber),
		nullIfEmpty(req.InventoryNumber),
		status,
		installedAt,
		nullIfEmpty(req.Description),
	)
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

	row, err := a.st.GetDeviceByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	resp := deviceDetailsResponse{
		Id:              row.ID,
		ModelId:         row.ModelID,
		LocationId:      row.LocationID,
		SerialNumber:    row.SerialNumber,
		InventoryNumber: row.InventoryNumber,
		Status:          row.Status,
		InstalledAt:     row.InstalledAt,
		Description:     row.Description,
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

	affected, err := a.st.UpdateDevice(
		r.Context(),
		id,
		req.ModelId,
		req.LocationId,
		nullIfEmpty(req.SerialNumber),
		nullIfEmpty(req.InventoryNumber),
		status,
		installedAt,
		nullIfEmpty(req.Description),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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

	affected, err := a.st.DeleteDevice(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	if affected == 0 {
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
