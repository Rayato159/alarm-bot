package kawaiiauth

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Rayato159/awaken-discord-bot/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type IKawaiiAuth interface {
	ParseJwtToken(ctx context.Context, tokenStr string) (*Payload, error)
	SignJwtToken(ctx context.Context, exp int64, payload *MiniPayload) (string, error)
}

type kawaiiAuth struct {
	cfg config.IConfig
}

func NewKawaiiAuth(cfg config.IConfig) IKawaiiAuth {
	return &kawaiiAuth{
		cfg: cfg,
	}
}

type MiniPayload struct {
}

type Payload struct {
	jwt.RegisteredClaims
}

func (k *kawaiiAuth) ParseJwtToken(ctx context.Context, tokenStr string) (*Payload, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error, unexpected signing method: %v", token.Header["alg"])
		}
		return k.cfg.Jwt().SecretKey(), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			msg := "error, token format is invalid"
			return nil, fmt.Errorf(msg)
		} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			msg := "error, token had expired"
			return nil, fmt.Errorf(msg)
		} else {
			return nil, err
		}
	}

	if claims, ok := token.Claims.(*Payload); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("error, payload is invalid")
	}
}

func (k *kawaiiAuth) SignJwtToken(ctx context.Context, exp int64, payload *MiniPayload) (string, error) {
	claims := &Payload{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(int(exp) * int(math.Pow10(9))))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "atomikku",
			Subject:   fmt.Sprintf("access_token"),
			ID:        uuid.New().String(),
			Audience:  []string{"rayato159"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(k.cfg.Jwt().SecretKey())
	if err != nil {
		return "", fmt.Errorf("error, sign token failed with an error: %v", err)
	}
	return ss, nil
}
