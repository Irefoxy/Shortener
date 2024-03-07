package service_impl

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s *ServiceImpl) checkRequest(c *gin.Context) {
	defer func() {
		if len(c.Errors) > 0 {
			c.Abort()
		}
	}()
	if c.Request.Body == nil {
		c.Error(errors.New("empty body")).SetType(gin.ErrorTypePublic)
	}
	if contentType := c.GetHeader("Content-Type"); contentType != "text/plain" {
		c.Error(errors.New("wrong content-type")).SetType(gin.ErrorTypePublic)
	}
}

func (s *ServiceImpl) handleUrl(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		return
	}
	request := string(data)
	newUrl, err := s.engine.Get(request)
	if err != nil {
		c.Error(err)
		return
	}
	err = s.repo.Set(newUrl, request)
	if err != nil {
		c.Error(err)
		return
	}
	c.String(201, "http://%s/%s", s.cfg.GetTargetAddress(), newUrl)
}

func (s *ServiceImpl) handleRedirect(c *gin.Context) {
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
