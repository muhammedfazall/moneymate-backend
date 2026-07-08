package jwtutil

import "github.com/golang-jwt/jwt/v5"

type AccesClaims struct {
	jwt.RegisteredClaims
    UserID       string   `json:"uid"`
    Handle       string   `json:"handle"`
    Roles        []string `json:"roles"`
    Permissions  []string `json:"permissions"`
    TokenVersion int64    `json:"ver"`
}

type RefreshClaims struct {
    jwt.RegisteredClaims
    UserID   string `json:"uid"`
    DeviceID string `json:"did"`
}

type TransactionClaims struct {
    jwt.RegisteredClaims          
    UserID string `json:"uid"`
}