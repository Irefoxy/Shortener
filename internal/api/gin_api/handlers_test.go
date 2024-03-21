package gin_api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// add init
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Get(hash string) (string, bool) {
	args := m.Called(hash)
	return args.String(0), args.Bool(1)
}
func (m *MockRepo) Set(_, _ string) error {
	return nil
}

type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) Get(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func initMock() *GinApi {
	mockRepo := new(MockRepo)
	mockEngine := new(MockEngine)

	mockRepo.On("Generate", "3JRsVv5L").Return("https://yandex.ru", true)
	mockRepo.On("Generate", "asd").Return("", false)

	mockEngine.On("Generate", "https://yandex.ru").Return("3JRsVv5L", nil)
	return &GinApi{
		repo:   mockRepo,
		parser: mockEngine,
	}
}

func getRouter() *gin.Engine {
	srv := initMock()
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.Use(srv.errorMiddleware)
	router.POST("/", srv.checkRequest, srv.handleUrl)
	router.GET("/*id", srv.handleRedirect)
	return router
}

func produceRequest(method, url, contentType string, reader io.Reader) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	router := getRouter()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, nil
}

func TestEndpoints(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		url          string
		contentType  string
		body         io.Reader
		expectedCode int
		expectedBody string
	}{
		{"Invalid URL Handler - Empty Body", "POST", "/", "text/plain", nil, http.StatusBadRequest, "Error: empty body"},
		{"Invalid URL Handler - Wrong Path", "POST", "/test", "text/plain", nil, http.StatusNotFound, "404 page not found"},
		{"Invalid URL Handler - Wrong ContentType", "POST", "/", "html", strings.NewReader("https://yandex.ru"), http.StatusBadRequest, "Error: wrong content-type"},
		{"Valid URL Handler", "POST", "/", "text/plain", strings.NewReader("https://yandex.ru"), http.StatusCreated, "http://localhost:8888/3JRsVv5L"},
		{"Invalid Redirect Handler", "GET", "/asd", "text/plain", nil, http.StatusBadRequest, "Error: no such short url"},
		{"Valid Redirect Handler", "GET", "/3JRsVv5L", "text/plain", nil, http.StatusTemporaryRedirect, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := produceRequest(tc.method, tc.url, tc.contentType, tc.body)
			if err != nil {
				t.Fatal(err)
			}

			if response.Code != tc.expectedCode {
				t.Errorf("Expected status code %d, but got %d", tc.expectedCode, response.Code)
			}

			if tc.expectedBody != "" && response.Body.String() != tc.expectedBody {
				t.Errorf("Expected body '%s', but got '%s'", tc.expectedBody, response.Body.String())
			}

			if tc.name == "Valid Redirect Handler" {
				if location := response.Header().Get("Location"); location != "https://yandex.ru" {
					t.Errorf("Expected location 'https://yandex.ru', but got '%s'", location)
				}
			}
		})
	}
}
