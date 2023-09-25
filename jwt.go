package hin

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"time"
)

type JwtClaims struct {
	Username string
	jwt.RegisteredClaims
}

type JwtTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var (
	JwtAccessKey  = "ACCESS"
	JwtRefreshKey = "REFRESH"
)

type jwtOptions struct{}

var JwtOptions = new(jwtOptions)

type jwtOption func(*JwtClaims)

func (jwtOptions) WithExpireTime(t time.Time) jwtOption {
	return func(claims *JwtClaims) {
		claims.ExpiresAt = jwt.NewNumericDate(t)
	}
}

func (jwtOptions) WithId(id string) jwtOption {
	return func(claims *JwtClaims) {
		claims.ID = id
	}
}

func (jwtOptions) WithUsername(username string) jwtOption {
	return func(claims *JwtClaims) {
		claims.Username = username
	}
}

func (jwtOptions) WithClaims(c *JwtClaims) jwtOption {
	return func(claims *JwtClaims) {
		_ = copier.Copy(claims, c)
	}
}

func MakeDoubleToken(opts ...jwtOption) (*JwtTokens, error) {
	claims := new(JwtClaims)

	for _, o := range opts {
		o(claims)
	}

	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	secretKey := getJwtSecret()
	tokens := new(JwtTokens)

	claims.Issuer = JwtAccessKey
	claims.ExpiresAt = getExpire("jwt.expire.access")
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	if tk, err := t.SignedString(secretKey); err != nil {
		return tokens, err
	} else {
		tokens.AccessToken = tk
	}

	claims.Issuer = JwtRefreshKey
	claims.ExpiresAt = getExpire("jwt.expire.refresh")
	t = jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	if tk, err := t.SignedString(secretKey); err != nil {
		return tokens, err
	} else {
		tokens.RefreshToken = tk
	}

	return tokens, nil
}

func ParseToken(tk string) (*JwtClaims, error) {
	secretKey := getJwtSecret()

	tokenClaims, err := jwt.ParseWithClaims(tk, new(JwtClaims), func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*JwtClaims); ok && tokenClaims.Valid {
		return claims, nil
	}

	return nil, err
}

func getJwtSecret() []byte {
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		secret = "hancens"
	}
	return []byte(secret)
}

func getExpire(key string) *jwt.NumericDate {
	n := time.Now()
	expire := viper.GetInt(key)
	if expire != 0 {
		return jwt.NewNumericDate(n.Add(time.Second * time.Duration(expire)))
	} else {
		return jwt.NewNumericDate(n.Add(time.Hour * 24 * 7))
	}
}
