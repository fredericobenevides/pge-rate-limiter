package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStoreStrategy struct {
	mock.Mock
}

func (m *MockStoreStrategy) AllowRequest(ctx context.Context, id string, requestLimit int, windowDuration time.Duration, blockDuration time.Duration) (bool, error) {
	args := m.Called(ctx, id, requestLimit, windowDuration, blockDuration)
	return args.Bool(0), args.Error(1)
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name                 string
		id                   string
		header               map[string]string
		remoteAddr           string
		respAllowed          bool
		respError            error
		expectedHttpResponse int
		expectedBodyResponse string
	}{
		{
			name:                 "API key request",
			id:                   "my_api_key",
			header:               map[string]string{"API_KEY": "my_api_key"},
			remoteAddr:           "127.0.0.1:1234",
			respAllowed:          true,
			respError:            nil,
			expectedHttpResponse: http.StatusOK,
			expectedBodyResponse: "",
		},
		{
			name:                 "IP request",
			id:                   "127.0.0.1",
			header:               map[string]string{},
			remoteAddr:           "127.0.0.1:1234",
			respAllowed:          true,
			respError:            nil,
			expectedHttpResponse: http.StatusOK,
			expectedBodyResponse: "",
		},

		{
			name:                 "Api Key internalerror",
			id:                   "my_api_key",
			header:               map[string]string{"API_KEY": "my_api_key"},
			remoteAddr:           "127.0.0.1:1234",
			respAllowed:          false,
			respError:            errors.New("internal error"),
			expectedHttpResponse: http.StatusInternalServerError,
			expectedBodyResponse: "Internal server error\n",
		},
		{
			name:                 "IP internalerror",
			id:                   "127.0.0.1",
			header:               map[string]string{},
			remoteAddr:           "127.0.0.1:1234",
			respAllowed:          false,
			respError:            errors.New("internal error"),
			expectedHttpResponse: http.StatusInternalServerError,
			expectedBodyResponse: "Internal server error\n",
		},

		{
			name:                 "Api Key not allowed",
			id:                   "my_api_key",
			header:               map[string]string{"API_KEY": "my_api_key"},
			remoteAddr:           "127.0.0.1:1234",
			respAllowed:          false,
			respError:            nil,
			expectedHttpResponse: http.StatusTooManyRequests,
			expectedBodyResponse: "you have reached the maximum number of requests or actions allowed within a certain time frame\n",
		},
		{
			name:                 "IP not allowed",
			id:                   "127.0.0.1",
			header:               map[string]string{},
			remoteAddr:           "127.0.0.1:1234",
			respAllowed:          false,
			respError:            nil,
			expectedHttpResponse: http.StatusTooManyRequests,
			expectedBodyResponse: "you have reached the maximum number of requests or actions allowed within a certain time frame\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStoreStrategy{}
			mockStore.On("AllowRequest", mock.Anything, tt.id, 10, time.Minute, time.Minute).
				Return(tt.respAllowed, tt.respError).
				Once()

			t.Cleanup(func() {
				mockStore.AssertExpectations(t)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)
			request.RemoteAddr = tt.remoteAddr

			for k, v := range tt.header {
				request.Header.Set(k, v)
			}

			middleware := NewRateLimitMiddleware(mockStore, 10, 10, time.Minute, time.Minute, time.Minute)

			handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(recorder, request)

			assert.Equal(t, tt.expectedHttpResponse, recorder.Code)
		})
	}
}

func TestGetClientIP(t *testing.T) {
	middleware := NewRateLimitMiddleware(nil, 10, 10, time.Minute, time.Minute, time.Minute)

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "177.100.150.200, 140.82.112.3"},
			expectedIP: "177.100.150.200",
		},
		{
			name:       "X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "10.0.0.1"},
			expectedIP: "10.0.0.1",
		},
		{
			name:       "IPv4 remote address",
			remoteAddr: "127.0.0.1:1234",
			expectedIP: "127.0.0.1",
		},
		{
			name:       "IPv6 local remote address",
			remoteAddr: "[::1]:1234",
			expectedIP: "127.0.0.1",
		},
		{
			name:       "IPv6 remote address",
			remoteAddr: "[2001:db8::1]:1234",
			expectedIP: "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip, err := middleware.getClientIP(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}

}
