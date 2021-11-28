package model

import "time"

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}
