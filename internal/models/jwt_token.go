package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT is model that will represent the JWT claims
//
// it contains standard jwt claims and private claims
type JWT struct {
	// we will embed all standard claims.
	jwt.RegisteredClaims

	// UserID will be encrypted. thats why it's string.
	//
	// 	it will be operator id if target user type has been set to enums.Operator
	UserID string `json:"user_id"`

	// Cellphone is the user's cellphone number
	Cellphone string `json:"cellphone,omitempty"`
}

// IsExpired will check that the target token has been expired or not.
func (j *JWT) IsExpired() bool {
	return j.ExpiresAt.After(time.Now())
}

// GetCellphone returns the cellphone from the token
func (j *JWT) GetCellphone() string {
	return j.Cellphone
}
