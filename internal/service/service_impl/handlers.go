package service_impl

import (
	"Yandex/internal/service"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func setAbort(c *gin.Context) {
	if len(c.Errors) > 0 {
		c.Abort()
	}
}

func (s *ServiceImpl) checkRequest(c *gin.Context) {
	defer setAbort(c)
	if c.Request.Body == nil {
		c.Error(errors.New("empty body")).SetType(gin.ErrorTypePublic)
	}
	if contentType := c.GetHeader("Content-Type"); contentType != "text/plain" && c.Request.RequestURI != "/shorten" || contentType != "application/json" {
		c.Error(errors.New("wrong content-type")).SetType(gin.ErrorTypePublic)
	}
}

func (s *ServiceImpl) handleUrl(c *gin.Context) {
	defer setAbort(c)
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		return
	}
	request := string(data)
	newUrl, done := s.processUri(c, request)
	if !done {
		return
	}
	c.String(201, "http://%s/%s", s.cfg.GetTargetAddress(), newUrl)
}

func (s *ServiceImpl) processUri(c *gin.Context, request string) (string, bool) {
	newUrl, err := s.engine.Get(request)
	if err != nil {
		c.Error(err)
		return "", false
	}
	err = s.repo.Set(newUrl, request)
	if err != nil {
		c.Error(err)
		return "", false
	}
	return newUrl, true
}

func (s *ServiceImpl) handleRedirect(c *gin.Context) {
	defer setAbort(c)
	param := c.Param("id")
	id := strings.TrimPrefix(param, "/")
	v, ok := s.repo.Get(id)
	if ok {
		c.Redirect(http.StatusTemporaryRedirect, v)
		return
	}
	c.Error(errors.New("no such short url")).SetType(gin.ErrorTypePublic)
}

func (s *ServiceImpl) errorMiddleware(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		switch c.Errors[0].Type {
		case gin.ErrorTypePublic:
			c.String(400, "Error: %s", c.Errors[0].Error())
		default:
			c.String(http.StatusInternalServerError, "Error: Something went wrong")
		}
	}
}

func (s *ServiceImpl) requestLoggerMiddleware(c *gin.Context) {
	startTime := time.Now()
	c.Next()
	duration := time.Since(startTime)

	s.logger.WithFields(logrus.Fields{
		"method":   c.Request.Method,
		"uri":      c.Request.RequestURI,
		"duration": duration,
	}).Info("request handled")
}

func (s *ServiceImpl) responseLoggerMiddleware(c *gin.Context) {
	c.Next()
	s.logger.WithFields(logrus.Fields{
		"status": c.Writer.Status(),
		"size":   c.Writer.Size(),
	}).Info("response handled")
}

func (s *ServiceImpl) handleJsonUrl(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	var request = service.URL{}
	if err := decoder.Decode(&request); err != nil {
		c.Error(err)
		return
	}
	newUrl, done := s.processUri(c, request.Url)
	if !done {
		return
	}
	c.JSON(http.StatusOK, service.Result{Result: newUrl})
}
