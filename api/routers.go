package api

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"net/http"
	"strings"
	"ticketing-service/data"
	"ticketing-service/logger"
	"ticketing-service/service"
	utils "ticketing-service/util"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Validation struct {
	validate *validator.Validate
}

type AuthHandler struct {
	validator   *data.Validation
	authService service.Authentication
}

func NewAuthHandler(v *data.Validation, auth service.Authentication) *AuthHandler {
	return &AuthHandler{
		validator:   v,
		authService: auth,
	}
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	logger.Log().Debug(authHeader)
	authHeaderContent := strings.Split(authHeader, " ")
	if len(authHeaderContent) < 2 || len(authHeaderContent) != 2 {
		return "", errors.New("Token not provided or malformed")
	}
	sDec, _ := b64.StdEncoding.DecodeString(authHeaderContent[1])
	logger.Log().Debug(string(sDec))

	if string(sDec) != "admin:admin" {
		return "", errors.New("Unauthorized user")
	}
	return authHeaderContent[1], nil
}

type UserIDKey struct{}

func (ah *AuthHandler) MiddlewareValidateAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		logger.Log().Debug("validating access token")
		token, err := extractToken(r)
		if err != nil {
			logger.Log().Error("Token not provided or malformed")
			w.WriteHeader(http.StatusBadRequest)
			out := "{\"Status\": false, Message: \"Authentication failed. Token not provided or malformed\"}"
			http.Error(w, string(out), http.StatusBadRequest)
			return
		}
		logger.Log().Debug("token present in header", token)
		userID := "admin"
		logger.Log().Debug("access token validated")
		ctx := context.WithValue(r.Context(), UserIDKey{}, userID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func NewRouter() *mux.Router {
	sm := mux.NewRouter()
	//var handler http.Handler
	configs := utils.NewConfigurations()
	validator := data.NewValidation()
	authService := service.NewAuthService(configs)
	uh := NewAuthHandler(validator, authService)

	bookTicket := sm.PathPrefix("/").Methods(http.MethodPost).Subrouter()
	bookTicket.HandleFunc("/tickets", BookTicket)
	bookTicket.Use(uh.MiddlewareValidateAccessToken)

	updateTicket := sm.PathPrefix("/").Methods(http.MethodPut).Subrouter()
	updateTicket.HandleFunc("/tickets", UpdateReservation)
	updateTicket.Use(uh.MiddlewareValidateAccessToken)

	cancelTicket := sm.PathPrefix("/").Methods(http.MethodDelete).Subrouter()
	cancelTicket.HandleFunc("/tickets", CancelTicket)
	cancelTicket.Use(uh.MiddlewareValidateAccessToken)
	return sm
}
