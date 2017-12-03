// Package jwt provides a service to create and verify
// JWT auth tokens for the bebop web app.
package jwt

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gopkg.in/dgrijalva/jwt-go.v3"
)

// Service is a JWT helper service that creates and verifies auth tokens.
type Service interface {
	Create(userID int64) (token string, err error)
	Verify(token string) (userID int64, issuedAt time.Time, err error)
}

// NewService creates a new JWT service using the given secret (32-byte hex-encoded).
func NewService(secret string) (Service, error) {
	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwt secret from hex: %s", err)
	}

	if len(secretBytes) < 32 {
		return nil, errors.New("jwt: secret too short")
	}

	return &service{secret: secretBytes}, nil
}

type service struct {
	secret []byte
}

type claims struct {
	jwt.StandardClaims
	UserID *int64 `json:"_uid"`
}

// Create creates a JWT string using the given secret key.
func (s *service) Create(userID int64) (string, error) {
	c := claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		UserID: &userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", errors.New("jwt: token signing failed: " + err.Error())
	}

	return tokenString, nil
}

// Verify verifies the JWT string using the given secret key.
// On success it returns the user ID and the time the token was issued.
func (s *service) Verify(tokenString string) (int64, time.Time, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("jwt: unexpected signing method")
			}
			return s.secret, nil
		},
	)
	if err != nil {
		return 0, time.Time{}, errors.New("jwt: ParseWithClaims failed: " + err.Error())
	}
	if !token.Valid {
		return 0, time.Time{}, errors.New("jwt: token is not valid")
	}

	c, ok := token.Claims.(*claims)
	if !ok {
		return 0, time.Time{}, errors.New("jwt: failed to get token claims")
	}

	if c.UserID == nil {
		return 0, time.Time{}, errors.New("jwt: UserID claim is not valid")
	}

	if c.IssuedAt == 0 {
		return 0, time.Time{}, errors.New("jwt: IssuedAt claim is not valid")
	}

	return *c.UserID, time.Unix(c.IssuedAt, 0), nil
}
