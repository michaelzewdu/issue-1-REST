package rest

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"net/http"
)

// CheckForAuthenticationMiddleware returns a gorrila/mux middleware function that checks 
// if the attached request has a valid authentication token. 
func CheckForAuthenticationMiddleware(authBackend *JWTAuthenticationBackend, logger *Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return authBackend.key, nil
			})
			if err == nil && token.Valid && !authBackend.IsInBlacklist(r.Header.Get("Authorization")) {
				claimMap, _ := token.Claims.(jwt.MapClaims)
				username := claimMap["sub"]
				r.Header.Add("authorized_username", username.(string))
				next.ServeHTTP(w, r)
			} else {
				(*logger).Log("unauthorized access attempt")
				w.WriteHeader(http.StatusUnauthorized)
			}
		})
	}
}

// postTokenAuth returns a handler for POST /token-auth requests
func postTokenAuth(authBackend *JWTAuthenticationBackend, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		responseCode := http.StatusOK
		response.Status = "fail"

		requestUser := new(user.User)
		err := json.NewDecoder(r.Body).Decode(&requestUser)
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, use format {"username":"username","password":"password"}`
			response.Data = responseData
			(*logger).Log("bad auth request")
			responseCode = http.StatusBadRequest
		} else {
			success, err := (*authBackend).Authenticate(requestUser)
			if err != nil {
				(*logger).Log("auth failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when generating token"
				response.Data = responseData
				responseCode = http.StatusInternalServerError
			}
			if success {
				tokenString, err := (*authBackend).GenerateToken(requestUser.Username)
				if err != nil {
					(*logger).Log("token generation failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when authenticating"
					response.Data = responseData
					responseCode = http.StatusInternalServerError
				} else {
					response.Status = "success"
					var responseData struct {
						Data string `json:"token"`
					}
					responseData.Data = tokenString
					response.Data = responseData
				}
			} else {
				(*logger).Log("unsuccessful auth attempt")
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "incorrect username or password"
				response.Data = responseData
				responseCode = http.StatusUnauthorized
			}

		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		w.WriteHeader(responseCode)
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// getTokenAuthRefresh returns a handler for GET /token-auth-refresh requests
func getTokenAuthRefresh(authBackend *JWTAuthenticationBackend, logger *Logger) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		var response jSendResponse
		responseCode := http.StatusOK
		response.Status = "fail"

		tokenString, err := (*authBackend).GenerateToken(r.Header.Get("authorized_username"))
		if err != nil {
			(*logger).Log("token generation failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when authenticating"
			response.Data = responseData
			responseCode = http.StatusInternalServerError
		} else {
			response.Status = "success"
			var responseData struct {
				Data string `json:"token"`
			}
			responseData.Data = tokenString
			response.Data = responseData
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		w.WriteHeader(responseCode)
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// getLogout returns a handler for GET /logout requests
func getLogout(authBackend *JWTAuthenticationBackend, logger *Logger) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		var response jSendResponse
		responseCode := http.StatusOK
		response.Status = "fail"

		err := invalidateAttachedToken(r, authBackend)
		if err != nil {
			(*logger).Log("logout failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when logging out"
			response.Data = responseData
			responseCode = http.StatusInternalServerError
		} else {
			response.Status = "success"
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		w.WriteHeader(responseCode)
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// invalidateAttachedToken is a helper function.
func invalidateAttachedToken(req *http.Request, authBackend *JWTAuthenticationBackend) error {
	tokenString := req.Header.Get("Authorization")
	return authBackend.AddToBlacklist(tokenString)
}
