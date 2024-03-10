package service_impl

import (
	"Yandex/internal/repo"
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
	v, err := s.repo.Get(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if v != "" {
		c.Redirect(http.StatusTemporaryRedirect, v)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("no such short url")).SetType(gin.ErrorTypePublic)
}

func (s *ServiceImpl) handleJsonUrl(c *gin.Context) {
	request := processJson[service.URL](c)
	newUrl, done := s.processUri(c, request.Url)
	if !done {
		return
	}
	c.JSON(http.StatusOK, service.Response{Result: newUrl})
}

func (s *ServiceImpl) handleJsonBatch(c *gin.Context) {
	requests := processJson[[]service.BatchUrl](c)
	var responses []service.BatchResponse
	for _, request := range requests {
		newUrl, done := s.processUri(c, request.Original)
		if !done {
			continue
		}
		responses = append(responses, service.BatchResponse{
			Id:    request.Id,
			Short: newUrl,
		})
	}
	if len(responses) == 0 {
		return
	}
	c.JSON(http.StatusOK, responses)
}

func processJson[T any](c *gin.Context) (request T) {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&request); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	return
}

func (s *ServiceImpl) handlePing(c *gin.Context) {
	if dbRepo, ok := s.repo.(repo.DbRepo); ok {
		if err := dbRepo.Ping(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithError(http.StatusInternalServerError, errors.New("no db connected"))
}
