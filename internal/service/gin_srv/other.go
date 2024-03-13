package gin_srv

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

func (s *GinService) errorMiddleware(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		switch c.Errors[0].Type {
		case gin.ErrorTypePublic:
			c.String(-1, "Error: %s", c.Errors[0].Error())
		default:
			c.String(http.StatusInternalServerError, "Error: Something went wrong")
		}
		for _, err := range c.Errors {
			s.logger.WithFields(logrus.Fields{
				"method": c.Request.Method,
				"uri":    c.Request.RequestURI,
				"error":  err.Error(),
			}).Warn("errors occurred")
		}
	}
}

func unzipMiddleware(c *gin.Context) {
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

func (s *GinService) requestLoggerMiddleware(c *gin.Context) {
	startTime := time.Now()
	c.Next()
	duration := time.Since(startTime)

	s.logger.WithFields(logrus.Fields{
		"method":   c.Request.Method,
		"uri":      c.Request.RequestURI,
		"duration": duration,
	}).Info("request handled")
}

func (s *GinService) responseLoggerMiddleware(c *gin.Context) {
	c.Next()
	s.logger.WithFields(logrus.Fields{
		"status": c.Writer.Status(),
		"size":   c.Writer.Size(),
	}).Info("response handled")
}

func checkRequest(c *gin.Context) {
	if c.Request.Body == nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("empty body")).SetType(gin.ErrorTypePublic)
	}
	if !checkContent(c) {
		c.AbortWithError(http.StatusBadRequest, errors.New("wrong content-type")).SetType(gin.ErrorTypePublic)
	}
}

func checkContent(c *gin.Context) bool {
	checkContent := map[string]string{
		"/":              "text/plain",
		"/shorten":       "application/json",
		"/shorten/batch": "application/json",
	}
	if checkContent[c.Request.RequestURI] == c.GetHeader("Content-Type") {
		return true
	}
	return false
}

func (s *GinService) handleWildcard(c *gin.Context) {
	path := c.Param("id")
	if path == "/ping" {
		s.handlePing(c)
	} else {
		s.handleUrl(c)
	}
}
