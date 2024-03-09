package service_impl

import (
	"compress/gzip"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s *ServiceImpl) errorMiddleware(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		switch c.Errors[0].Type {
		case gin.ErrorTypePublic:
			c.String(-1, "Error: %s", c.Errors[0].Error())
		default:
			c.String(http.StatusInternalServerError, "Error: Something went wrong")
		}
	}
}

func (s *ServiceImpl) unzipMiddleware(c *gin.Context) {
	if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
			return
		}
		defer gz.Close()
		c.Request.Body = io.NopCloser(gz)
	}

	c.Next()
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

func (s *ServiceImpl) checkRequest(c *gin.Context) {
	if c.Request.Body == nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("empty body")).SetType(gin.ErrorTypePublic)
	}
	if contentType := c.GetHeader("Content-Type"); !checkContent(contentType, c) {
		c.AbortWithError(http.StatusBadRequest, errors.New("wrong content-type")).SetType(gin.ErrorTypePublic)
	}
}

func checkContent(contentType string, c *gin.Context) bool {
	if contentType == "text/plain" && c.Request.RequestURI == "/" {
		return true
	}
	if contentType == "application/json" && c.Request.RequestURI == "/shorten" {
		return true
	}
	return false
}
