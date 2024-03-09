package service_impl

import (
	"Yandex/internal/service"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func (s *ServiceImpl) handleUrl(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	request := string(data)
	newUrl, done := s.processUri(c, request)
	if !done {
		return
	}
	c.String(201, "http://%s/%s", s.cfg.TargetAddress, newUrl)
}

func (s *ServiceImpl) processUri(c *gin.Context, request string) (string, bool) {
	newUrl, err := s.engine.Get(request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return "", false
	}
	err = s.repo.Set(newUrl, request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return "", false
	}
	return newUrl, true
}

func (s *ServiceImpl) handleRedirect(c *gin.Context) {
	param := c.Param("id")
	id := strings.TrimPrefix(param, "/")
	v, ok := s.repo.Get(id)
	if ok {
		c.Redirect(http.StatusTemporaryRedirect, v)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("no such short url")).SetType(gin.ErrorTypePublic)
}

func (s *ServiceImpl) handleJsonUrl(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	var request = service.URL{}
	if err := decoder.Decode(&request); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	newUrl, done := s.processUri(c, request.Url)
	if !done {
		return
	}
	c.JSON(http.StatusOK, service.Result{Result: newUrl})
}
