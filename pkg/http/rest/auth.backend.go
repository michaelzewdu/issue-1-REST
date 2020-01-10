package rest

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"time"
)

// JWTAuthenticationBackend provides methods for implementation of a JWT based authentication
type JWTAuthenticationBackend struct {
	setup     *Setup
	blacklist map[string]struct{}
}

// NewJWTAuthenticationBackend returns a new JWTAuthenticationBackend that uses the passed arguments
func NewJWTAuthenticationBackend(s *Setup) *JWTAuthenticationBackend {
	return &JWTAuthenticationBackend{
		setup:     s,
		blacklist: make(map[string]struct{}),
	}
}

// GenerateToken generates a new JWT token based on the given username.
// It uses the HS-SHA512 encryption standard.
func (backend *JWTAuthenticationBackend) GenerateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	mapClaim := jwt.MapClaims{}
	mapClaim["exp"] = time.Now().Add(backend.setup.TokenAccessLifetime).Unix()
	mapClaim["iat"] = time.Now().Unix()
	mapClaim["sub"] = username
	token.Claims = mapClaim
	tokenString, err := token.SignedString(backend.setup.TokenSigningSecret)
	if err != nil {
		return "", fmt.Errorf("token signing failed because %v", err)
	}
	return tokenString, nil
}

// Authenticate checks whether the given User struct holds appropriate credentials
func (backend *JWTAuthenticationBackend) Authenticate(user *user.User) (bool, error) {
	return backend.setup.UserService.Authenticate(user)
}

// AddToBlacklist adds a given token to the list of tokens that can not be used no more.
func (backend *JWTAuthenticationBackend) AddToBlacklist(tokenString string) error {
	(*backend).blacklist[tokenString] = struct{}{}
	return nil
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

// IsInBlacklist checks whether a given token is invalidated previously.
func (backend *JWTAuthenticationBackend) IsInBlacklist(token string) bool {
	_, ok := (*backend).blacklist[token]

	// ok will be true if it's in the blacklist
	return ok
}
