package gin_srv

import (
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func (s *GinService) setCookie(c *gin.Context) {
	if _, ok := c.Get("user_id"); !ok {
		cookie := s.cookie.createSignedCookie("user_id")
		http.SetCookie(c.Writer, cookie)
	}
}

func checkAuthentication(c *gin.Context) {
	if _, ok := c.Get("user_id"); !ok {
		c.AbortWithError(http.StatusUnauthorized, errors.New("authorization fail")).SetType(gin.ErrorTypePublic)
	}
}

func (s *GinService) authentication(c *gin.Context) {
	if value, err := c.Cookie("user_id"); err == nil {
		if id, ok := s.cookie.verifyCookie(value); ok {
			c.Set("userID", id)
		}
	}
}

func (e cookieEngine) verifyCookie(cookieValue string) (string, bool) {
	parts := strings.Split(cookieValue, "|")
	if len(parts) != 2 {
		return "", false
	}
	value := parts[0]
	signature := parts[1]
	expectedSignature := e.getSignature(value)
	return value, e.equal([]byte(signature), []byte(expectedSignature))
}

func (e cookieEngine) getSignature(value string) string {
	e.hasher.Write([]byte(value))
	expectedSignature := hex.EncodeToString(e.hasher.Sum(nil))
	e.hasher.Reset()
	return expectedSignature
}

func (e cookieEngine) createSignedCookie(name string) *http.Cookie {
	value := uuid.New().String()
	signature := e.getSignature(value)
	signedValue := value + "|" + signature

	return &http.Cookie{
		Name:     name,
		Value:    signedValue,
		HttpOnly: true,
	}
}
