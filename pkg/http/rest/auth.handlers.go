package rest

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"net/http"
)

// CheckForAuthenticationMiddleware returns a gorilla/mux middleware function that checks
// if the attached request has a valid authentication token.
func CheckForAuthenticationMiddleware(s *Setup) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return s.jwtBackend.key, nil
			})
			if err == nil && token.Valid && !s.jwtBackend.IsInBlacklist(r.Header.Get("Authorization")) {
				claimMap, _ := token.Claims.(jwt.MapClaims)
				username := claimMap["sub"]
				r.Header.Add("authorized_username", username.(string))
				next.ServeHTTP(w, r)
			} else {
				s.Logger.Log("unauthorized access attempt")
				w.WriteHeader(http.StatusUnauthorized)
			}
		})
	}
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
			s.Logger.Log("bad auth request")
			statusCode = http.StatusBadRequest
		} else {
			success, err := s.jwtBackend.Authenticate(requestUser)
			if err != nil {
				s.Logger.Log("auth failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when generating token"
				statusCode = http.StatusInternalServerError
			}
			if success {
				tokenString, err := s.jwtBackend.GenerateToken(requestUser.Username)
				if err != nil {
					s.Logger.Log("token generation failed because: %v", err)
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
				}
			} else {
				s.Logger.Log("unsuccessful auth attempt")
				response.Data = jSendFailData{
					ErrorReason:  "credentials",
					ErrorMessage: "incorrect username or password",
				}
				statusCode = http.StatusUnauthorized
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getTokenAuthRefresh returns a handler for GET /token-auth-refresh requests
func getTokenAuthRefresh(s *Setup) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		tokenString, err := s.jwtBackend.GenerateToken(r.Header.Get("authorized_username"))
		if err != nil {
			s.Logger.Log("token generation failed because: %v", err)
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
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getLogout returns a handler for GET /logout requests
func getLogout(s *Setup) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		err := invalidateAttachedToken(r, s)
		if err != nil {
			s.Logger.Log("logout failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when logging out"
			statusCode = http.StatusInternalServerError
		} else {
			response.Status = "success"
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// invalidateAttachedToken is a helper function.
func invalidateAttachedToken(req *http.Request, s *Setup) error {
	tokenString := req.Header.Get("Authorization")
	return s.jwtBackend.AddToBlacklist(tokenString)
}
