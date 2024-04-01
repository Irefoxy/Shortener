package gin_api

import (
	"Yandex/internal/api/gin_api/mocks"
	"Yandex/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type ApiSuite struct {
	suite.Suite
	api *GinApi
	srv Service
	w   *httptest.ResponseRecorder
}

func (s *ApiSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.w = httptest.NewRecorder()
	s.srv = mocks.NewMockService(ctrl)

	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
	s.api = New(s.srv, &models.ApiConf{}, logger)
}

func (s *ApiSuite) produceRequest(method, url, contentType string, reader io.Reader) error {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	router := s.api.init()
	router.ServeHTTP(s.w, req)
	return nil
}

func (s *ApiSuite) TestInvalidURLHandler_EmptyBody() {
	err := s.produceRequest("POST", "/", "text/plain", nil)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, s.w.Code)
	s.Equal("Error: empty body", s.w.Body.String())
}

func (s *ApiSuite) TestInvalidURLHandler_WrongPath() {
	err := s.produceRequest("POST", "/test", "text/plain", nil)
	s.NoError(err)
	s.Equal(http.StatusNotFound, s.w.Code)
	s.Equal("404 page not found", s.w.Body.String())
}

func (s *ApiSuite) TestInvalidURLHandler_WrongContentType() {
	err := s.produceRequest("POST", "/", "html", strings.NewReader("https://yandex.ru"))
	s.NoError(err)
	s.Equal(http.StatusBadRequest, s.w.Code)
	s.Equal("Error: wrong content-type", s.w.Body.String())
}

func (s *ApiSuite) TestValidURLHandler() {
	err := s.produceRequest("POST", "/", "text/plain", strings.NewReader("https://yandex.ru"))
	s.NoError(err)
	s.Equal(http.StatusCreated, s.w.Code)
	s.Equal("http://localhost:8888/3JRsVv5L", s.w.Body.String())
}

func (s *ApiSuite) TestInvalidRedirectHandler() {
	err := s.produceRequest("GET", "/asd", "text/plain", nil)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, s.w.Code)
	s.Equal("Error: no such short url", s.w.Body.String())
}

func (s *ApiSuite) TestValidRedirectHandler() {
	err := s.produceRequest("GET", "/3JRsVv5L", "text/plain", nil)
	s.NoError(err)
	s.Equal(http.StatusTemporaryRedirect, s.w.Code)
	location := s.w.Header().Get("Location")
	s.Equal("https://yandex.ru", location)
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}
