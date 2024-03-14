package gin_srv

import (
	"Yandex/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *GinService) handleUrl(c *gin.Context) {
	urls, err := readRequest[string](c)
	if errorSetter(c, http.StatusBadRequest, err) != nil {
		return
	}
	shortUrl, err := s.processSingleRequest(c, urls[0])
	switch {
	case errors.Is(err, models.ErrorConflict):
		c.String(http.StatusConflict, "http://%s/%s", s.cfg.TargetAddress, shortUrl)
	case err != nil:
		errorSetter(c, http.StatusInternalServerError, err)
	default:
		c.String(http.StatusCreated, "http://%s/%s", s.cfg.TargetAddress, shortUrl)
	}
}

func (s *GinService) handleJsonUrl(c *gin.Context) {
	request, err := readRequest[models.URL](c)
	if errorSetter(c, http.StatusBadRequest, err) != nil {
		return
	}
	shortUrl, err := s.processSingleRequest(c, request[0].Url)
	switch {
	case errors.Is(err, models.ErrorConflict):
		c.JSON(http.StatusConflict, models.ShortURL{Result: s.cfg.TargetAddress + "/" + shortUrl})
	case err != nil:
		errorSetter(c, http.StatusInternalServerError, err)
	default:
		c.JSON(http.StatusCreated, models.ShortURL{Result: s.cfg.TargetAddress + "/" + shortUrl})
	}
}

func (s *GinService) handleJsonBatch(c *gin.Context) {
	requests, err := readRequest[models.BatchURL](c)
	if errorSetter(c, http.StatusInternalServerError, err) != nil {
		return
	}
	urls := getUrlsFromJsonBatch(requests)
	newUrls, err := s.getNewUrls(urls...)
	if errorSetter(c, http.StatusInternalServerError, err) != nil {
		return
	}
	units := getServiceUnits(c.GetString(cookieName), newUrls, urls...)
	err = s.repo.SetBatch(c.Request.Context(), units)
	responses := buildResponses(requests, newUrls)
	switch {
	case errors.Is(err, models.ErrorConflict):
		c.JSON(http.StatusConflict, responses)
	case err != nil:
		errorSetter(c, http.StatusInternalServerError, err)
	default:
		c.JSON(http.StatusCreated, responses)
	}
}

func buildResponses(requests []models.BatchURL, urls []string) (responses []models.BatchShortURL) {
	for i, request := range requests {
		responses = append(responses, models.BatchShortURL{
			Id:    request.Id,
			Short: urls[i],
		})
	}
	return
}

func (s *GinService) handleRedirect(c *gin.Context) {
	shortUrlWithPrefix := c.Param("id")
	shortUrl := strings.TrimPrefix(shortUrlWithPrefix, "/")
	v, err := s.repo.Get(c.Request.Context(), models.ServiceUnit{
		Id:       c.GetString(cookieName),
		ShortUrl: shortUrl,
	})
	switch {
	case err != nil:
		errorSetter(c, http.StatusInternalServerError, err)
	case v != nil:
		c.Redirect(http.StatusTemporaryRedirect, v.OriginalUrl)
	default:
		errorSetter(c, http.StatusBadRequest, models.ErrorShortURLNotExist)
	}
}

func (s *GinService) handlePing(c *gin.Context) {
	if _, ok := s.repo.(DbRepo); !ok {
		errorSetter(c, http.StatusInternalServerError, models.ErrorDBNotConnected)
		return
	}
	if errorSetter(c, http.StatusInternalServerError, s.repo.(DbRepo).Ping()) != nil {
		return
	}
	c.Status(http.StatusOK)
	return
}

func readRequest[T any](c *gin.Context) ([]T, error) {
	var request []T
	if err := c.ShouldBind(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func getServiceUnits(id string, short []string, original ...string) (units []models.ServiceUnit) {
	for i, s := range original {
		units = append(units, models.ServiceUnit{
			Id:          id,
			OriginalUrl: s,
			ShortUrl:    short[i],
		})
	}
	return
}

func (s *GinService) processSingleRequest(c *gin.Context, url string) (string, error) {
	newUrls, err := s.getNewUrls(url)
	if err != nil {
		return "", err
	}
	units := getServiceUnits(c.GetString(cookieName), newUrls, url)
	return units[0].ShortUrl, s.repo.Set(c.Request.Context(), units[0])
}

func (s *GinService) getNewUrls(urls ...string) ([]string, error) {
	var newUrls []string
	for _, url := range urls {
		newUrl, err := s.parser.Parse([]byte(url))
		if err != nil {
			return nil, err
		}
		newUrls = append(newUrls, newUrl)
	}
	return newUrls, nil
}

func getUrlsFromJsonBatch(batch []models.BatchURL) (urls []string) {
	for _, elm := range batch {
		urls = append(urls, elm.Original)
	}
	return
}
