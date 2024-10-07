package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	Minute   = 60
	Hour     = 60 * Minute
	Day      = 24 * Hour
	Month    = 30 * Day
	SixMonth = 6 * Month
)

var ErrTokenExpired = errors.New("время токена вышло")

type IJWT interface {
	// Generate jwt token
	// minutes - how long token will be valid in minutes.
	GenerateJWT(uuid.UUID, string, string, bool, int) string
	GenerateRefreshToken(uuid.UUID, string, string, bool, int) (string, time.Time)

	GenerateTokenCookie(string, string, time.Time) *http.Cookie

	RefreshAccessToken(string) (string, time.Time, error)

	ParseJWT(string) (Claims, error)

	SetRefreshTokenValidator(func(string) (bool, error))
	SetInvalidateToken(func(string) (bool, error))

	ValidateRefreshToken(string) (bool, error)
	InvalidateRefreshToken(string) (bool, error)
}

type JWT struct {
	secret                string
	refreshTokenValidator func(string) (bool, error)
	invalidateToken       func(string) (bool, error)
}

func New(secret string) IJWT {
	return &JWT{
		secret: secret,
	}
}

func (j *JWT) SetRefreshTokenValidator(fn func(string) (bool, error)) {
	j.refreshTokenValidator = fn
}

func (j *JWT) SetInvalidateToken(fn func(string) (bool, error)) {
	j.invalidateToken = fn
}

func (j *JWT) ValidateRefreshToken(token string) (bool, error) {
	if token == "" {
		return false, errors.New("refresh token is empty")
	}

	if j.refreshTokenValidator == nil {
		return false, errors.New("refresh token validator fn is not set")
	}

	return j.refreshTokenValidator(token)
}

func (j *JWT) InvalidateRefreshToken(token string) (bool, error) {
	if token == "" {
		return false, errors.New("token is empty")
	}

	if j.invalidateToken == nil {
		return false, errors.New("invalidate token fn is not set")
	}

	return j.invalidateToken(token)
}

func (j *JWT) GenerateJWT(uid uuid.UUID, email, name string, isValid bool, seconds int) string {
	claims := Claims{
		UUID:    uid,
		Email:   email,
		Name:    name,
		IsValid: isValid,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "todo",
			Subject:   "user",
			Audience:  jwt.ClaimStrings{"todo"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(seconds))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return ""
	}

	return secret
}

func (j *JWT) GenerateRefreshToken(uid uuid.UUID, email, name string, isValid bool, seconds int) (string, time.Time) {
	exp := time.Now().Add(time.Second * time.Duration(seconds))

	claims := Claims{
		UUID:    uid,
		Email:   email,
		Name:    name,
		IsValid: isValid,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "todo",
			Subject:   "refresh",
			Audience:  jwt.ClaimStrings{"todo"},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret, err := token.SignedString([]byte(j.secret))
	if err != nil {
		logrus.Errorf("[jwt][email:%s] error generating jwt", email)
		return "", exp
	}

	return secret, exp
}

func (j *JWT) RefreshAccessToken(refreshToken string) (secret string, expAt time.Time, err error) {
	if len(refreshToken) < 10 {
		return secret, expAt, errors.New("invalid refresh token")
	}

	claims, err := j.ParseJWT(refreshToken)
	if err != nil {
		return secret, expAt, err
	}

	if !claims.IsRefresh() {
		return secret, expAt, errors.New("invalid refresh token")
	}

	if ok, err := j.ValidateRefreshToken(refreshToken); !ok {
		return secret, expAt, err
	}

	secret = j.GenerateJWT(claims.UUID, claims.Email, claims.Name, claims.IsValid, Minute*10)

	return secret, claims.ExpiresAt.Time, nil
}

func (j *JWT) ParseJWT(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		return Claims{}, ErrTokenExpired
	}

	if err != nil {
		return Claims{}, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return *claims, nil
	}

	return Claims{}, nil
}

func (j *JWT) GenerateTokenCookie(access, refresh string, expAfter time.Time) *http.Cookie {
	value := ""
	if access != "" && refresh != "" {
		value = fmt.Sprintf("%s&%s", access, refresh)
	}

	return &http.Cookie{
		Name:     "TOKEN",
		Value:    value,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  expAfter,
		Path:     "/",
	}
}
