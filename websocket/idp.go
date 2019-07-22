package websocket

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	issuer                 = "hexya"
	duration time.Duration = 24 * time.Hour
)

var (
	// ErrConflict indicates usage of the existing email during account
	// registration.
	ErrConflict error = errors.New("already taken")

	// ErrInvalidCredentials indicates malformed account credentials.
	ErrInvalidCredentials error = errors.New("invalid email or password")

	// ErrMalformedClient indicates malformed client specification (e.g. empty name).
	ErrMalformedClient error = errors.New("malformed client specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess error = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound error = errors.New("non-existent entity")
)

// IdentityProvider specifies an API for identity management via security
// tokens.
type IdentityProvider interface {
	SignedClaims(claims jwt.MapClaims) (string, error)

	// TemporaryKey generates the temporary access token.
	TemporaryKey(string) (string, error)

	// RefreshKey generates the refresh access token.
	RefreshKey(id string) (string, error)

	// PermanentKey generates the non-expiring access token.
	PermanentKey(string) (string, error)

	// Identity extracts the entity identifier given its secret key.
	Identity(string) (string, error)

	IdentityMap(key string) (jwt.MapClaims, error)
}

var _ IdentityProvider = (*jwtIdentityProvider)(nil)

type jwtIdentityProvider struct {
	secret string
	aud    string
}

// NewIdentityProvider instantiates a JWT identity provider.
func NewIdentityProvider(secret string, aud string) IdentityProvider {
	return &jwtIdentityProvider{secret: secret, aud: aud}
}

func (idp *jwtIdentityProvider) TemporaryKey(id string) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(6 * 30 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"aud": idp.aud,
		"sub": id,
		"iss": issuer,
		"iat": now.Unix(),
		"exp": exp.Unix(),
	}
	return idp.jwt(claims)
}

func (idp *jwtIdentityProvider) RefreshKey(id string) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(duration)

	claims := jwt.MapClaims{
		"aud": idp.aud,
		"sub": id,
		"iss": issuer,
		"iat": now.Unix(),
		"exp": exp.Unix(),
	}
	return idp.jwt(claims)
}

func (idp *jwtIdentityProvider) PermanentKey(id string) (string, error) {
	claims := jwt.MapClaims{
		"aud": idp.aud,
		"sub": id,
		"iss": issuer,
		"iat": time.Now().UTC().Unix(),
	}
	return idp.jwt(claims)
}

func (idp *jwtIdentityProvider) jwt(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(idp.secret))
}

func (idp *jwtIdentityProvider) SignedClaims(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(idp.secret))
}

func (idp *jwtIdentityProvider) Identity(key string) (string, error) {
	token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorizedAccess
		}

		return []byte(idp.secret), nil
	})

	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string), nil
	}

	return "", ErrUnauthorizedAccess
}

func (idp *jwtIdentityProvider) IdentityMap(key string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorizedAccess
		}

		return []byte(idp.secret), nil
	})

	if err != nil {
		return nil, ErrUnauthorizedAccess
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrUnauthorizedAccess
}

var _, idp IdentityProvider

func DefaultIdentityProvider() IdentityProvider {
	return idp
}

func init() {
	secrect := fmt.Sprintf("$gutdoo@%d#", 982911000)
	idp = NewIdentityProvider(secrect, "hexya")
}
