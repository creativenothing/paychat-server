package sessions

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SessionClaims struct {
	UserID string
	jwt.StandardClaims
}

func (us *UserSession) GetJWT() string {
	claims := SessionClaims{
		UserID: us.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: (time.Now().Add(time.Hour)).UTC().Unix(),
			Issuer:    "nameOfWebsiteHere",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte("secureSecretText"))

	return signedToken
}

func GetUserSessionByJWT(webtoken string) *UserSession {
	token, err := jwt.ParseWithClaims(
		webtoken,
		&SessionClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte("secureSecretText"), nil
		},
	)
	if err != nil {
		return nil
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok {
		return nil
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil
	}

	return GetUserSessionByID(claims.UserID)
}
