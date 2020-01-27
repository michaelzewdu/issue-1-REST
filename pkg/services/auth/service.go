package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Service defines an interface for authentication
type Service interface {
	Authenticate(user *User) (bool, error)
	GenerateToken(username string) (string, error)
	AddToBlacklist(tokenString string) error
	IsInBlacklist(token string) (bool, error)
}

// Repository defines an interface that provides persistence functionality for the search service.
type Repository interface {
	Authenticate(user *User) (bool, error)
	AddToBlacklist(tokenString string) error
	IsInBlacklist(token string) (bool, error)
}

// ErrUserNotFound is returned when the the username specified isn't recognized
var ErrUserNotFound = fmt.Errorf("user not found")

// jWTAuthenticationBackend provides methods for implementation of a JWT based authentication
type jWTAuthenticationBackend struct {
	TokenAccessLifetime, TokenRefreshLifetime time.Duration
	TokenSigningSecret                        []byte
	repo                                      *Repository
}

// NewAuthService returns a new JWTAuthenticationBackend that uses the passed arguments
func NewAuthService(r *Repository, tokenAccessLifetime, tokenRefreshLifetime time.Duration, tokenSigningSecret []byte) Service {
	return &jWTAuthenticationBackend{
		TokenAccessLifetime:  tokenAccessLifetime,
		TokenRefreshLifetime: tokenRefreshLifetime,
		TokenSigningSecret:   tokenSigningSecret,
		repo:                 r,
	}
}

// GenerateToken generates a new JWT token based on the given username.
// It uses the HS-SHA512 encryption standard.
func (s *jWTAuthenticationBackend) GenerateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	mapClaim := jwt.MapClaims{}
	mapClaim["exp"] = time.Now().Add(s.TokenAccessLifetime).Unix()
	mapClaim["iat"] = time.Now().Unix()
	mapClaim["sub"] = username
	token.Claims = mapClaim
	tokenString, err := token.SignedString(s.TokenSigningSecret)
	if err != nil {
		return "", fmt.Errorf("token signing failed because %v", err)
	}
	return tokenString, nil
}

// Authenticate checks whether the given User struct holds appropriate credentials
func (s *jWTAuthenticationBackend) Authenticate(user *User) (bool, error) {
	return (*s.repo).Authenticate(user)
}

// AddToBlacklist adds a given token to the list of tokens that can not be used no more.
func (s *jWTAuthenticationBackend) AddToBlacklist(tokenString string) error {
	return (*s.repo).AddToBlacklist(tokenString)
}

// IsInBlacklist checks whether a given token is invalidated previously.
func (s *jWTAuthenticationBackend) IsInBlacklist(tokenString string) (bool, error) {
	return (*s.repo).IsInBlacklist(tokenString)
}

/*func (backend *JWTAuthenticationBackend) getTokenRemainingValidity(timestamp interface{}) int {
	const expireOffset = 3600
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainder := tm.Sub(time.Now())
		if remainder > 0 {
			return int(remainder.Seconds() + expireOffset)
		}
	}
	return expireOffset
}*/
