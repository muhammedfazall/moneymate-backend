package jwtutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
	
)

type Config struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiryMins  int
	RefreshExpiryHrs  int
	TxTokenExpirySecs int
}

type AccessTokenParams struct {
	UserID       string
	Handle       string
	Roles        []string
	Permissions  []string
	TokenVersion int64
}

type RefreshTokenParams struct {
	UserID   string
	DeviceID string
}

type TransactionTokenParams struct {
	UserID string
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshTokenHash string
}



//generate access token 

func GenerateAccessToken(p AccessTokenParams, cfg Config) (string, error) {
    jti, err := generateJTI()
    if err != nil {
        return "", fmt.Errorf("generate jti: %w", err)
    }
    claims := &AccesClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   p.UserID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(
                time.Duration(cfg.AccessExpiryMins) * time.Minute,
            )),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            ID:       jti,
        },
        UserID:       p.UserID,
        Handle:       p.Handle,
        Roles:        p.Roles,
        Permissions:  p.Permissions,
        TokenVersion: p.TokenVersion,
    }

    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(cfg.AccessSecret))
}


//generate refresh token 

func GenerateRefreshToken(p RefreshTokenParams, cfg Config) (token, hash string, err error) {
    jti, err := generateJTI()
    if err != nil {
        return "", "", fmt.Errorf("generate jti: %w", err)
    }

    claims := &RefreshClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   p.UserID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(
                time.Duration(cfg.RefreshExpiryHrs) * time.Hour,
            )),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            ID:       jti,
        },
        UserID:   p.UserID,
        DeviceID: p.DeviceID,
    }

    signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(cfg.RefreshSecret))
    if err != nil {
        return "", "", fmt.Errorf("sign refresh token: %w", err)
    }

    return signed, HashToken(signed), nil
}


//generate access and referresh token 
func GenerateTokenPair(ap AccessTokenParams, rp RefreshTokenParams, cfg Config) (*TokenPair, error) {
    accessToken, err := GenerateAccessToken(ap, cfg)
    if err != nil {
        return nil, fmt.Errorf("access token: %w", err)
    }

    refreshToken, refreshHash, err := GenerateRefreshToken(rp, cfg)
    if err != nil {
        return nil, fmt.Errorf("refresh token: %w", err)
    }

    return &TokenPair{
        AccessToken:      accessToken,
        RefreshToken:     refreshToken,
        RefreshTokenHash: refreshHash,
    }, nil
}

//generate transaction token 

func GenerateTransactionToken(p TransactionTokenParams, cfg Config) (string, error) {
    jti, err := generateJTI()
    if err != nil {
        return "", fmt.Errorf("generate jti: %w", err)
    }

    expiry := cfg.TxTokenExpirySecs
    if expiry == 0 {
        expiry = 60
    }

    claims := &TransactionClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   p.UserID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(
                time.Duration(expiry) * time.Second,
            )),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            ID:       jti,
        },
        UserID: p.UserID,
    }

    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(cfg.AccessSecret))
}



//parse access token 

func ParseAccessToken(tokenString, secret string) (*AccesClaims, error) {
    claims := &AccesClaims{}

    _, err := jwt.ParseWithClaims(tokenString, claims, signingKeyFunc(secret))
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return claims, apperrors.ErrTokenExpired
        }
        return nil, apperrors.ErrInvalidToken
    }

    return claims, nil
}



func ParseRefreshToken(tokenString, secret string) (*RefreshClaims, error) {
    claims := &RefreshClaims{}

    _, err := jwt.ParseWithClaims(tokenString, claims, signingKeyFunc(secret))
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, apperrors.ErrTokenExpired
        }
        return nil, apperrors.ErrInvalidToken
    }

    return claims, nil
}



func ParseTransactionToken(tokenString, secret string) (*TransactionClaims, error) {
    claims := &TransactionClaims{}

    _, err := jwt.ParseWithClaims(tokenString, claims, signingKeyFunc(secret))
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, apperrors.ErrTokenExpired
        }
        return nil, apperrors.ErrInvalidToken
    }

    return claims, nil
}

// -------------------------------------helpers-----------------------------------


func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate jti: %w", err)
	}
	return hex.EncodeToString(b), nil
}


func HashToken(raw string) string {
    h := sha256.Sum256([]byte(raw))
    return hex.EncodeToString(h[:])
}


func signingKeyFunc(secret string) jwt.Keyfunc {
    return func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return []byte(secret), nil
    }
}