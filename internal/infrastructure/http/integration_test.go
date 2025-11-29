package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/i18n"
	"apocapoc-api/internal/infrastructure/auth"
	"apocapoc-api/internal/infrastructure/crypto"
	"apocapoc-api/internal/infrastructure/persistence/sqlite"

	_ "modernc.org/sqlite"
)

type TestServer struct {
	Router *http.Handler
	DB     *sql.DB
}

func setupTestServer(t *testing.T) *TestServer {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := sqlite.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	jwtService := auth.NewJWTService("test-secret", 24)
	passwordHasher := crypto.NewBcryptHasher()

	userRepo := sqlite.NewUserRepository(db)
	habitRepo := sqlite.NewHabitRepository(db)
	entryRepo := sqlite.NewHabitEntryRepository(db)
	refreshTokenRepo := sqlite.NewRefreshTokenRepository(db)
	passwordResetTokenRepo := sqlite.NewPasswordResetTokenRepository(db)

	registerHandler := commands.NewRegisterUserHandler(userRepo, passwordHasher, nil, "", "open", false)
	loginHandler := queries.NewLoginUserHandler(userRepo, passwordHasher)
	refreshTokenHandler := queries.NewRefreshTokenHandler(refreshTokenRepo, userRepo)
	revokeTokenHandler := commands.NewRevokeTokenHandler(refreshTokenRepo)
	revokeAllTokensHandler := commands.NewRevokeAllTokensHandler(refreshTokenRepo)
	verifyEmailHandler := commands.NewVerifyEmailHandler(userRepo, nil, false)
	resendVerificationEmailHandler := commands.NewResendVerificationEmailHandler(userRepo, nil, "")
	requestPasswordResetHandler := commands.NewRequestPasswordResetHandler(userRepo, passwordResetTokenRepo, nil, "")
	resetPasswordHandler := commands.NewResetPasswordHandler(userRepo, passwordResetTokenRepo, passwordHasher)
	createHandler := commands.NewCreateHabitHandler(habitRepo)
	getTodaysHandler := queries.NewGetTodaysHabitsHandler(habitRepo, entryRepo)
	getUserHabitsHandler := queries.NewGetUserHabitsHandler(habitRepo)
	getHabitByIDHandler := queries.NewGetHabitByIDHandler(habitRepo)
	getHabitEntriesHandler := queries.NewGetHabitEntriesHandler(habitRepo, entryRepo)
	getHabitStatsHandler := queries.NewGetHabitStatsHandler(habitRepo, entryRepo)
	updateHandler := commands.NewUpdateHabitHandler(habitRepo)
	archiveHandler := commands.NewArchiveHabitHandler(habitRepo)
	markHandler := commands.NewMarkHabitHandler(entryRepo, habitRepo)
	unmarkHandler := commands.NewUnmarkHabitHandler(habitRepo, entryRepo)

	refreshTokenExpiry := 7 * 24 * time.Hour

	deleteUserHandler := commands.NewDeleteUserHandler(userRepo)

	translator, _ := i18n.NewTranslator()

	authHandlers := NewAuthHandlers(registerHandler, loginHandler, refreshTokenHandler, revokeTokenHandler, revokeAllTokensHandler, verifyEmailHandler, resendVerificationEmailHandler, requestPasswordResetHandler, resetPasswordHandler, jwtService, refreshTokenRepo, refreshTokenExpiry, translator)
	habitHandlers := NewHabitHandlers(createHandler, getTodaysHandler, getUserHabitsHandler, getHabitByIDHandler, getHabitEntriesHandler, updateHandler, archiveHandler, markHandler, unmarkHandler, translator)
	statsHandlers := NewStatsHandlers(getHabitStatsHandler, translator)
	healthHandlers := NewHealthHandlers(db)
	userHandlers := NewUserHandlers(deleteUserHandler, translator)

	router := NewRouter("http://localhost:3000", habitHandlers, authHandlers, statsHandlers, healthHandlers, userHandlers, jwtService, translator)

	handler := http.Handler(router)
	return &TestServer{
		Router: &handler,
		DB:     db,
	}
}

func (ts *TestServer) Close() {
	ts.DB.Close()
}

func makeRequest(t *testing.T, handler http.Handler, method, path string, body interface{}, authToken string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	if err := json.NewDecoder(rr.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func registerAndLogin(t *testing.T, router http.Handler, email, password string) string {
	registerBody := RegisterRequest{
		Email:    email,
		Password: password,
		Timezone: "UTC",
	}
	makeRequest(t, router, "POST", "/api/v1/auth/register", registerBody, "")

	loginBody := LoginRequest{
		Email:    email,
		Password: password,
	}
	rr := makeRequest(t, router, "POST", "/api/v1/auth/login", loginBody, "")

	var authResp AuthResponse
	decodeResponse(t, rr, &authResp)
	return authResp.Token
}
