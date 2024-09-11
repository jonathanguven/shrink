package utils

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// generate JWT token for user given their user ID
func GenerateJWT(userID uint, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// set JWT token in HTTP-only cookie
func SetCookie(w http.ResponseWriter, token string) {
	log.WithFields(log.Fields{
		"secure": os.Getenv("ENVIRONMENT"),
	}).Infof("Environment: %s", os.Getenv("ENVIRONMENT"))
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})
}

// retrieve JWT token from a HTTP-only cookie
func GetJWTFromCookie(r *http.Request) (*jwt.Token, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}

	return jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
