package model

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	Role   string   `json:"role"`
	Action []string `json:"action,omitempty" bson:"action,omitempty"`
	jwt.StandardClaims
}
