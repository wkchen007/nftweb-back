package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
	RDB           *redis.Client
}

type jwtUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Auth) GenerateTokenPair(user *jwtUser) (TokenPairs, error) {
	// Create a token
	token := jwt.New(jwt.SigningMethodHS256)
	accessJTI := fmt.Sprintf("acc-%d-%d", user.ID, time.Now().UnixNano())
	refreshJTI := fmt.Sprintf("ref-%d-%d", user.ID, time.Now().UnixNano())

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"
	claims["jti"] = accessJTI

	// Set the expiry for JWT
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// Create a signed token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// Create a refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	// Set the expiry for the refresh token
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()
	refreshTokenClaims["jti"] = refreshJTI

	// Create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// 寫入 Redis（allowlist）
	ctx := context.Background()
	if j.RDB != nil {
		// 只存 JTI 即可，值可放 userID 或 "ok"
		if err := j.RDB.Set(ctx, "access:"+accessJTI, user.ID, j.TokenExpiry).Err(); err != nil {
			return TokenPairs{}, fmt.Errorf("redis set access jti: %w", err)
		}
		if err := j.RDB.Set(ctx, "refresh:"+refreshJTI, user.ID, j.RefreshExpiry).Err(); err != nil {
			_ = j.RDB.Del(ctx, "access:"+accessJTI).Err()
			return TokenPairs{}, fmt.Errorf("redis set refresh jti: %w", err)
		}
		log.Printf("stored jti in redis: %s", claims["jti"])
	}

	// Create TokenPairs and populate with signed tokens
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	// Return TokenPairs
	return tokenPairs, nil
}

func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	w.Header().Add("Vary", "Authorization")

	// get auth header
	authHeader := r.Header.Get("Authorization")
	//log.Printf("auth header: %s", authHeader)
	// sanity check
	if authHeader == "" {
		log.Print("no auth header")
		return "", nil, fmt.Errorf("no auth header")
	}

	// split the header on spaces
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		log.Print("invalid auth header")
		return "", nil, fmt.Errorf("invalid auth header")
	}

	// check to see if we have the word Bearer
	if headerParts[0] != "Bearer" {
		log.Print("invalid auth header")
		return "", nil, fmt.Errorf("invalid auth header")
	}

	token := headerParts[1]

	// declare an empty claims
	claims := &Claims{}

	// parse the token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			log.Print("expired token")
			return "", nil, fmt.Errorf("expired token")
		}
		return "", nil, err
	}

	if claims.Issuer != j.Issuer {
		log.Print("invalid issuer")
		return "", nil, fmt.Errorf("invalid issuer")
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != j.Audience {
		return "", nil, fmt.Errorf("invalid audience")
	}

	//log.Printf("claims: %+v", claims)

	// Redis 檢查（allowlist & revoke）
	if j.RDB != nil {
		ctx := context.Background()
		jti := claims.ID
		if jti == "" {
			return "", nil, fmt.Errorf("missing jti")
		}

		// 是否被撤銷？
		revoked, err := j.RDB.Exists(ctx, "revoked:"+jti).Result()
		if err != nil {
			return "", nil, fmt.Errorf("redis error: %w", err)
		}
		if revoked == 1 {
			return "", nil, fmt.Errorf("token revoked")
		}

		// 是否在 allowlist？
		allowed, err := j.RDB.Exists(ctx, "access:"+jti).Result()
		if err != nil {
			return "", nil, fmt.Errorf("redis error: %w", err)
		}
		if allowed != 1 {
			return "", nil, fmt.Errorf("token not in allowlist")
		}
	}

	return token, claims, nil
}

// 登出：撤銷 access token
func (j *Auth) RevokeAccessToken(jti string, exp time.Time) error {
	if j.RDB == nil {
		return nil
	}
	ctx := context.Background()
	ttl := time.Until(exp)
	if ttl <= 0 {
		ttl = time.Minute // 防呆
	}
	pipe := j.RDB.TxPipeline()
	pipe.Set(ctx, "revoked:"+jti, 1, ttl)
	pipe.Del(ctx, "access:"+jti) // 從 allowlist 移除，立即失效
	_, err := pipe.Exec(ctx)
	return err
}
