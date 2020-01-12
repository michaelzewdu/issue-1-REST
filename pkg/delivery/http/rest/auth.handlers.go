package rest

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/slim-crown/issue-1-REST/pkg/services/domain/user"
	"net/http"
	"time"
)

// ParseAuthTokenMiddleware checks  if the attached request has a valid
// authentication token. If valid JWT token found, it'll extract the
// sub, the username in this case and attaches it to the passed request.
// If no token is found, it'll attach an invalid username.
func ParseAuthTokenMiddleware(s *Setup) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
				func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return s.TokenSigningSecret, nil
				})
			if err == nil && token.Valid && !s.jwtBackend.IsInBlacklist(r.Header.Get("Authorization")) {
				claimMap, _ := token.Claims.(jwt.MapClaims)
				if claimMap.VerifyExpiresAt(time.Now().Unix(), true) {
					// if valid and not expired
					username := claimMap["sub"]
					r.Header.Set("authorized_username", username.(string))
					r.Header.Del("authorized_username_expired")
					next.ServeHTTP(w, r)
					return
				} else if claimMap.VerifyExpiresAt(time.Now().Add(-s.TokenRefreshLifetime).Unix(), true) {
					// if expired but still refreshable
					username := claimMap["sub"]
					r.Header.Set("authorized_username_expired", username.(string))
					r.Header.Del("authorized_username")
					next.ServeHTTP(w, r)
					return
				} else {
					s.Logger.Printf("unauthorized access with expired token")
				}
			}
			// if not valid
			r.Header.Set("authorized_username", "HerUsernameIs25LettersLng")
			r.Header.Del("authorized_username_expired")
			next.ServeHTTP(w, r)
		})
	}
}

// CheckForAuthMiddleware blocks access if there's no valid credential's attached
// on the request from the ParseAuthTokenMiddleware.
func CheckForAuthMiddleware(s *Setup) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isAuthenticated(r, s) {
				next.ServeHTTP(w, r)
				return
			}
			s.Logger.Printf("unauthenticated access attempt")
			w.WriteHeader(http.StatusUnauthorized)
		})
	}
}

func isAuthenticated(r *http.Request, s *Setup) bool {
	authUsername := r.Header.Get("authorized_username")
	if authUsername != "" && len(authUsername) < 25 {
		return true
	}
	return false
}

// postTokenAuth returns a handler for POST /token-auth requests
func postTokenAuth(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		requestUser := new(user.User)
		err := json.NewDecoder(r.Body).Decode(&requestUser)
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "request format",
				ErrorMessage: `bad request, use format {"username":"username","password":"password"}`,
			}
			s.Logger.Printf("bad auth request")
			statusCode = http.StatusBadRequest
		} else {
			success, err := s.jwtBackend.Authenticate(requestUser)
			switch err {
			case nil:
				if success {
					tokenString, err := s.jwtBackend.GenerateToken(requestUser.Username)
					if err != nil {
						s.Logger.Printf("token generation failed because: %v", err)
						response.Status = "error"
						response.Message = "server error when authenticating"
						statusCode = http.StatusInternalServerError
					} else {
						response.Status = "success"
						var responseData struct {
							Data string `json:"token"`
						}
						responseData.Data = tokenString
						response.Data = responseData
						s.Logger.Printf("user %s got token", requestUser.Username)
					}
				} else {
					s.Logger.Printf("unsuccessful authentication attempt on nonexisting user")
					response.Data = jSendFailData{
						ErrorReason:  "credentials",
						ErrorMessage: "incorrect username or password",
					}
					statusCode = http.StatusUnauthorized
				}
			case user.ErrUserNotFound:
				s.Logger.Printf("unsuccessful authentication attempt")
				response.Data = jSendFailData{
					ErrorReason:  "credentials",
					ErrorMessage: "incorrect username or password",
				}
				statusCode = http.StatusUnauthorized
			default:
				s.Logger.Printf("auth failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when generating token"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getTokenAuthRefresh returns a handler for GET /token-auth-refresh requests
func getTokenAuthRefresh(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		{ // this block secures the route
			if isAuthenticated(r, s) {
				// pass, all is good
			} else if r.Header.Get("authorized_username_expired") != "" {
				r.Header.Set("authorized_username", r.Header.Get("authorized_username_expired"))
			} else {
				s.Logger.Printf("unauthorized refresh request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		tokenString, err := s.jwtBackend.GenerateToken(r.Header.Get("authorized_username"))
		if err != nil {
			s.Logger.Printf("token generation failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when authenticating"
			statusCode = http.StatusInternalServerError
		} else {
			response.Status = "success"
			var responseData struct {
				Data string `json:"token"`
			}
			responseData.Data = tokenString
			response.Data = responseData
			s.Logger.Printf("user %s refreshed token", r.Header.Get("authorized_username"))
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getLogout returns a handler for GET /logout requests
func getLogout(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		err := invalidateAttachedToken(r, s)
		if err != nil {
			s.Logger.Printf("logout failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when logging out"
			statusCode = http.StatusInternalServerError
		} else {
			response.Status = "success"
			s.Logger.Printf("token was invalidated")
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// invalidateAttachedToken is a helper function.
func invalidateAttachedToken(req *http.Request, s *Setup) error {
	tokenString := req.Header.Get("Authorization")
	return s.jwtBackend.AddToBlacklist(tokenString)
}
