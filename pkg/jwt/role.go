package jwt

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/qlc-hub/pkg/util"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrInvalidRoleclaims = errors.New("invalid role claims")
	Admin                = []string{admin}
	User                 = []string{user}
	Both                 = []string{admin, user}
)

const (
	admin = "admin"
	user  = "user"
)

type RoleClaims struct {
	jwt.StandardClaims
	Roles []string `json:"roles"`
}

func (r *RoleClaims) IsAuthorized(role string) bool {
	if err := r.Valid(); err != nil {
		return false
	}
	if role == "" {
		return false
	}
	for _, s := range r.Roles {
		if s == role {
			return true
		}
	}
	return false
}

func (r *RoleClaims) String() string {
	return util.ToString(r)
}

type JWTManager struct {
	pubKey        interface{}
	privateKey    interface{}
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	if privateKey, err := FromBase58(secretKey); err == nil {
		return &JWTManager{
			pubKey:        privateKey.Public(),
			privateKey:    privateKey,
			tokenDuration: tokenDuration,
		}, nil
	} else {
		return nil, err
	}
}

func (m *JWTManager) Generate(roles []string) (string, error) {
	var expiresAt int64
	if m.tokenDuration == 0 {
		expiresAt = 0
	} else {
		expiresAt = time.Now().Add(m.tokenDuration).Unix()
	}
	user := &RoleClaims{
		Roles: roles,
		StandardClaims: jwt.StandardClaims{
			Audience:  "QLCChain Bot",
			ExpiresAt: expiresAt,
			Id:        uuid.New().String(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "QLCChain Bot",
			NotBefore: 0,
			Subject:   "signer",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, user)
	return token.SignedString(m.privateKey)
}

func (m *JWTManager) Verify(token string) (*RoleClaims, error) {
	if parsedToken, err := jwt.ParseWithClaims(token, &RoleClaims{}, func(parsedToken *jwt.Token) (interface{}, error) {
		return m.pubKey, nil
	}); err == nil {
		if claims, ok := parsedToken.Claims.(*RoleClaims); ok {
			return claims, nil
		} else {
			return nil, ErrInvalidRoleclaims
		}
	} else {
		return nil, err
	}
}

func (m *JWTManager) Refresh(token string) (string, error) {
	if user, err := m.Verify(token); err == nil {
		if err := user.Valid(); err == nil {
			if user.ExpiresAt != 0 {
				user.ExpiresAt = time.Now().Add(m.tokenDuration).Unix()
			}
			token := jwt.NewWithClaims(jwt.SigningMethodES512, user)
			return token.SignedString(m.privateKey)
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
