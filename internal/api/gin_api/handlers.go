package gin_api

import (
	m "Yandex/internal/api/gin_api/models"
	"Yandex/internal/converters"
	"Yandex/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *GinService) handleUrl(c *gin.Context) {
	shortUrl, err := s.handleSingleURL(c)
	sendResponse(c, shortUrl, err)
}

func (s *GinService) handleJsonUrl(c *gin.Context) {
	response, err := s.handleSingleJSON(c)
	sendResponse(c, response, err)
}

func (s *GinService) handleJsonBatch(c *gin.Context) {
	response, err := s.handleMultipleJSON(c)
	sendResponse(c, response, err)
}

func (s *GinService) handleRedirect(c *gin.Context) {
	shortUrlWithPrefix := c.Param("parameterName")
	shortUrl := strings.TrimPrefix(shortUrlWithPrefix, "/")
	v, err := s.service.Get(c.Request.Context(), converters.ApiShortUrlsToEntry(c.GetString(cookieName), shortUrl)[0])
	sendRedirect(c, v, err)
}

func (s *GinService) handlePing(c *gin.Context) {
	err := s.service.Ping(c.Request.Context())
	if err == nil {
		collectErrors(c, http.StatusInternalServerError, err, nil)
	}
	c.Status(http.StatusOK)
}

func (s *GinService) handleGetAll(c *gin.Context) {
	response, err := s.service.GetAll(c.Request.Context(), c.GetString(cookieName))
	sendAllURLS(c, response, err)
}

func (s *GinService) handleDelete(c *gin.Context) {
	err := s.processDeletion(c)
	switch {
	case err != nil:
		collectErrors(c, http.StatusInternalServerError, err, nil)
	default:
		c.Status(http.StatusAccepted)
	}
}

func (s *GinService) processDeletion(c *gin.Context) error {
	requests, err := readRequest[[]string](c)
	if err != nil {
		return err
	}
	return s.service.Delete(c.Request.Context(), converters.ApiShortUrlsToEntry(c.GetString(cookieName), requests...))
}

func (s *GinService) handleSingleURL(c *gin.Context) (result string, err error) {
	request, err := readRequest[string](c)
	if err != nil {
		return
	}
	operationReturn, err := s.service.Add(c.Request.Context(), converters.ApiUrlToEntry(request, c.GetString(cookieName)))
	if err == nil || errors.Is(err, models.ErrorConflict) {
		result = converters.EntryToApiUrl(operationReturn[0], *s.cfg.TargetAddress)
	}
	return
}

func (s *GinService) handleSingleJSON(c *gin.Context) (result m.ShortURL, err error) {
	request, err := readRequest[m.URL](c)
	if err != nil {
		return
	}
	operationReturn, err := s.service.Add(c.Request.Context(), converters.ApiJSONUrlToEntry(request, c.GetString(cookieName)))
	if err == nil || errors.Is(err, models.ErrorConflict) {
		result = converters.EntryToApiJSONUrl(operationReturn[0], *s.cfg.TargetAddress)
	}
	return
}

func (s *GinService) handleMultipleJSON(c *gin.Context) (result []m.BatchShortURL, err error) {
	requests, err := readRequest[[]m.BatchURL](c)
	if err != nil {
		return
	}
	operationReturn, err := s.service.Add(c.Request.Context(), converters.ApiJSONUrlBatchToEntry(requests, c.GetString(cookieName)))
	if err == nil || errors.Is(err, models.ErrorConflict) {
		result = converters.EntryToApiJSONUrlBatch(operationReturn, *s.cfg.TargetAddress, requests)
	}
	return
}

func readRequest[T any](c *gin.Context) (T, error) {
	var request T
	if err := c.ShouldBind(&request); err != nil {
		return request, err
	}
	return request, nil
}

func sendResponse(c *gin.Context, response any, err error) {
	if err != nil {
		if errors.Is(err, models.ErrorConflict) {
			collectErrors(c, http.StatusConflict, err, response)
		} else {
			collectErrors(c, http.StatusInternalServerError, err, nil)
		}
		return
	}
	status := http.StatusCreated
	switch response.(type) {
	case string:
		c.String(status, "%s", response)
	default:
		c.JSON(status, response)
	}
}

func sendRedirect(c *gin.Context, value *models.Entry, err error) {
	switch {
	case errors.Is(err, models.ErrorDeleted):
		collectErrors(c, http.StatusGone, err, nil)
	case err != nil:
		collectErrors(c, http.StatusInternalServerError, err, nil)
	case value == nil:
		collectErrors(c, http.StatusNotFound, models.ErrorShortURLNotExist, nil)
	default:
		c.Redirect(http.StatusTemporaryRedirect, value.OriginalUrl)
	}
}

func sendAllURLS(c *gin.Context, response []models.Entry, err error) {
	switch {
	case err != nil:
		collectErrors(c, http.StatusInternalServerError, err, nil)
	case response == nil:
		collectErrors(c, http.StatusNoContent, models.ErrorNoContent, nil)
	default:
		c.JSON(http.StatusOK, response)
	}
}
