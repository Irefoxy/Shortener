package gin_api

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

type ApiError struct {
	error
	Status int
	Data   any
}

func (s *GinApi) errorMiddleware(c *gin.Context) {
	c.Next()
	lastErr := c.Errors.Last()
	if lastErr == nil {
		return
	}
	err, ok := lastErr.Err.(ApiError)
	s.logError(c)
	if !ok || err.Status == http.StatusInternalServerError {
		c.String(http.StatusInternalServerError, "Something went wrong")
		return
	}
	if err.Data != nil {
		switch err.Data.(type) {
		case string:
			c.String(err.Status, "%s", err.Data)
		default:
			c.JSON(err.Status, err.Data)
		}
		return
	}
	c.Status(err.Status)
}

func (s *GinApi) logError(c *gin.Context) {
	for _, err := range c.Errors {
		s.logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"uri":    c.Request.RequestURI,
			"error":  err.Error(),
		}).Warn("errors occurred")
	}
}

func unzipMiddleware(c *gin.Context) {
	if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			collectErrors(c, http.StatusInternalServerError, err, nil)
			return
		}
		defer gz.Close()
		c.Request.Body = io.NopCloser(gz)
	}

	c.Next()
}

func (s *GinApi) requestLoggerMiddleware(c *gin.Context) {
	startTime := time.Now()
	c.Next()
	duration := time.Since(startTime)

	s.logger.WithFields(logrus.Fields{
		"method":   c.Request.Method,
		"uri":      c.Request.RequestURI,
		"duration": duration,
	}).Info("request handled")
}

func (s *GinApi) responseLoggerMiddleware(c *gin.Context) {
	c.Next()
	s.logger.WithFields(logrus.Fields{
		"status": c.Writer.Status(),
		"size":   c.Writer.Size(),
	}).Info("response handled")
}

func (s *GinApi) handleWildcard(c *gin.Context) {
	path := c.Param("id")
	switch path {
	case "/ping":
		s.handlePing(c)
	case "/api/user/urls":
		s.handleGetAll(c)
	default:
		s.handleUrl(c)
	}
}

func collectErrors(c *gin.Context, status int, err error, data any) {
	newErr := ApiError{
		error:  err,
		Status: status,
		Data:   data,
	}
	c.Error(newErr)
	c.Abort()
}
