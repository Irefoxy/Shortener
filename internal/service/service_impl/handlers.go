package service_impl

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func (s *ServiceImpl) handleUrl(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		return
	}
	request := string(data)
	v, ok := s.repo.Get(request)
	if ok {
		c.String(201, "text/plain", "http://"+"localhost:8888"+"/"+v)
		return
	}
	new_url, err := s.engine.Get(request)
	if err != nil {
		c.Error(err)
		return
	}
	err = s.repo.Set(new_url, request)
	if err != nil {
		c.Error(err)
		return
	}
	c.String(201, "http://%s/%s", "localhost:8888", new_url)
}

func (s *ServiceImpl) handleRedirect(c *gin.Context) {
	param := c.Param("id")
	id := strings.TrimPrefix(param, "/")
	v, ok := s.repo.Get(id)
	if ok {
		c.Redirect(http.StatusTemporaryRedirect, v)
		return
	}
	c.Error(errors.New("No such short url"))
}

func (s *ServiceImpl) errorMiddleware(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		switch c.Errors[0].Type {
		case gin.ErrorTypePublic:
			c.JSON(-1, gin.H{"error": c.Errors[0].Error()})
		default:
			c.JSON(-1, gin.H{"error": "Something went wrong"})
		}
	}
}
