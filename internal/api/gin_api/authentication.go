package gin_api

import (
	"Yandex/internal/models"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func (s *GinApi) setCookie(c *gin.Context) {
	if _, ok := c.Get(cookieName); !ok {
		cookie := s.cookie.createSignedCookie(cookieName)
		http.SetCookie(c.Writer, cookie)
		s.logger.Infof("Cookie: new cookie generated for %s", c.ClientIP())
	}
}

func checkAuthentication(c *gin.Context) {
	if _, ok := c.Get(cookieName); !ok {
		collectErrors(c, http.StatusUnauthorized, models.ErrorAuthorizationFailed, nil)
	}
}

func (s *GinApi) authentication(c *gin.Context) {
	if value, err := c.Cookie(cookieName); err == nil {
		s.logger.Infof("Authentication: %s found: %s for %s", cookieName, value, c.ClientIP())
		if id, ok := s.cookie.verifyCookie(value); ok {
			c.Set(cookieName, id)
			s.logger.Infof("Authentication for %s succeded", id)
			return
		}
		s.logger.Infof("Authentication for %s failed", c.ClientIP())
		return
	}
	s.logger.Infof("Authentication: %s not found for %s", cookieName, c.ClientIP())
}

func (e *cookieEngine) verifyCookie(cookieValue string) (string, bool) {
	parts := strings.Split(cookieValue, "|")
	if len(parts) != 2 {
		return "", false
	}
	value := parts[0]
	signature := parts[1]
	expectedSignature := e.getSignature(value)
	return value, e.equal([]byte(signature), []byte(expectedSignature))
}

func (e *cookieEngine) getSignature(value string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.hasher.Write([]byte(value))
	expectedSignature := hex.EncodeToString(e.hasher.Sum(nil))
	e.hasher.Reset()
	return expectedSignature
}

func (e *cookieEngine) createSignedCookie(name string) *http.Cookie {
	value := uuid.New().String()
	signature := e.getSignature(value)
	signedValue := value + "|" + signature

	return &http.Cookie{
		Name:     name,
		Value:    signedValue,
		HttpOnly: true,
	}
}
