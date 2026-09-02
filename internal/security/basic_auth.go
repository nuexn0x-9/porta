package security

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// CheckBasicAuth verifies HTTP Basic Authentication headers against configured credentials
func CheckBasicAuth(r *http.Request, configuredPassword string) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}

	// Support both "username:password" and single password format (default user "porta")
	expectedUser := "porta"
	expectedPass := configuredPassword

	if strings.Contains(configuredPassword, ":") {
		parts := strings.SplitN(configuredPassword, ":", 2)
		expectedUser = parts[0]
		expectedPass = parts[1]
	}

	userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(expectedUser)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPass)) == 1

	return userMatch && passMatch
}

// CheckTokenAuth verifies Bearer Token or query parameter against configured token
func CheckTokenAuth(r *http.Request, configuredToken string) bool {
	if configuredToken == "" {
		return false
	}

	// 1. Check Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(token), []byte(configuredToken)) == 1 {
			return true
		}
	}

	// 2. Check query parameter ?porta_token=<token>
	queryToken := r.URL.Query().Get("porta_token")
	if queryToken != "" {
		if subtle.ConstantTimeCompare([]byte(queryToken), []byte(configuredToken)) == 1 {
			return true
		}
	}

	return false
}

// RequireBasicAuth sends a 401 Unauthorized challenge response
func RequireBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="PORTA Protected Environment"`)
	http.Error(w, "401 Unauthorized: Credentials Required (PORTA)", http.StatusUnauthorized)
}

// RequireTokenAuth sends a 403 Forbidden response
func RequireTokenAuth(w http.ResponseWriter) {
	http.Error(w, "403 Forbidden: Invalid or Missing Access Token (PORTA)", http.StatusForbidden)
}
