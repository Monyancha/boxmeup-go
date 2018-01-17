package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cjsaylor/boxmeup-go/config"
	"github.com/cjsaylor/boxmeup-go/database"
	"github.com/cjsaylor/boxmeup-go/middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	chain "github.com/justinas/alice"
)

// Hook is the mechanism to plugin item module routes
type Hook struct{}

var routes = []config.Route{
	config.Route{
		Name:    "Login",
		Method:  "POST",
		Pattern: "/api/user/login",
		Handler: chain.New(middleware.JsonResponseHandler).ThenFunc(loginHandler),
	},
	config.Route{
		Name:    "Logout",
		Method:  "GET",
		Pattern: "/api/user/logout",
		Handler: chain.New(middleware.LogHandler).ThenFunc(logoutHandler),
	},
	config.Route{
		Name:    "Register",
		Method:  "POST",
		Pattern: "/api/user/register",
		Handler: chain.New(middleware.JsonResponseHandler).ThenFunc(registerHandler),
	},
	config.Route{
		Name:    "User",
		Method:  "GET",
		Pattern: "/api/user/current",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(userHandler),
	},
}

// Apply hooks related to items
func (h Hook) Apply(router *mux.Router) {
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.Handler)
	}
}

// LoginHandler authenticates via email and password
func loginHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()

	token, err := NewStore(db).Login(
		middleware.AuthConfig{
			LegacySalt: config.Config.LegacySalt,
			JWTSecret:  config.Config.JWTSecret,
		},
		req.PostFormValue("email"),
		req.PostFormValue("password"))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Authentication failure."})
	} else {
		expiration := time.Now().Add(14 * 24 * time.Hour)
		cookie := http.Cookie{
			Name:     middleware.SessionName,
			Value:    token,
			Expires:  expiration,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(res, &cookie)
		res.WriteHeader(http.StatusOK)
		jsonOut.Encode(map[string]string{
			"token": token,
		})
	}
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	cookie := http.Cookie{
		Name:     middleware.SessionName,
		Value:    "",
		Expires:  time.Now().Add(-100 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(res, &cookie)
	res.WriteHeader(http.StatusNoContent)
}

// RegisterHandler creates new users.
func registerHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()

	email := req.PostFormValue("email")
	password := req.PostFormValue("password")
	id, err := NewStore(db).Register(
		middleware.AuthConfig{
			LegacySalt: config.Config.LegacySalt,
			JWTSecret:  config.Config.JWTSecret,
		},
		email,
		password)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: err.Error()})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": id,
	})
}

// UserHandler returns basic user information of current user.
func userHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	user, err := NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "User specified not found."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(user)
}
