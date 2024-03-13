package gin_srv

import (
	"Yandex/internal/models"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func (s *GinService) handleUrl(c *gin.Context) {
	url, err := readPlainText(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	units, err := s.processUri(c, url)
	if err != nil {
		if !errors.Is(err, err) {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Status(http.StatusConflict)
	} else {
		c.Status(http.StatusCreated)
	}
	c.String(-1, "http://%s/%s", s.cfg.TargetAddress, units[0].ShortUrl)
}

func readPlainText(c *gin.Context) (models.Plain, error) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	request := models.Plain(data)
	return request, nil
}

type Requests interface {
	String() string
}

func (s *GinService) processUri(c *gin.Context, urls ...Requests) ([]models.ServiceUnit, error) {
	units, err := s.getUnits(urls, c.GetString("user_id"))
	if err != nil {
		return nil, err
	}

	err = s.repo.Set(units...)
	if err != nil {
		return nil, err
	}
	c.Status(http.StatusCreated)
	return units, nil
}

func (s *GinService) getUnits(urls []Requests, id string) ([]models.ServiceUnit, error) {
	units := make([]models.ServiceUnit, 0, len(urls))
	for _, url := range urls {
		newUrl, err := s.parser.Parse([]byte(url.String()))
		if err != nil {
			return nil, err
		}
		units = append(units, models.ServiceUnit{
			Id:          id,
			OriginalUrl: url.String(),
			ShortUrl:    newUrl,
		})
	}
	return units, nil
}

func (s *GinService) handleRedirect(c *gin.Context) {
	param := c.Param("id")
	id := strings.TrimPrefix(param, "/")
	v, err := s.repo.Get(models.ServiceUnit{
		Id:       c.GetString("user_id"),
		ShortUrl: id,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if v != nil {
		c.Redirect(http.StatusTemporaryRedirect, v[0].OriginalUrl)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("no such short url")).SetType(gin.ErrorTypePublic)
}

func (s *GinService) handleJsonUrl(c *gin.Context) {
	request, err := readJson[models.URL](c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	units, err := s.processUri(c, request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(-1, models.ShortURL{Result: s.cfg.TargetAddress + "/" + units[0].ShortUrl})
}

func (s *GinService) handleJsonBatch(c *gin.Context) {
	requests, err := readJson[[]models.BatchURL](c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var reqs []Requests
	for _, r := range requests {
		reqs = append(reqs, r)
	}
	units, err := s.processUri(c, reqs...)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var responses []models.BatchShortURL
	for i, unit := range units {
		responses = append(responses, models.BatchShortURL{
			Id:    requests[i].Id,
			Short: unit.ShortUrl,
		})
	}
	c.JSON(http.StatusOK, responses)
}

func readJson[T any](c *gin.Context) (request T, err error) {
	decoder := json.NewDecoder(c.Request.Body)
	if err = decoder.Decode(&request); err != nil {
		return
	}
	return
}

func (s *GinService) handlePing(c *gin.Context) {
	if dbRepo, ok := s.repo.(DbRepo); ok {
		if err := dbRepo.Ping(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithError(http.StatusInternalServerError, errors.New("no db connected"))
}
