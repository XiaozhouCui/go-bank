package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/XiaozhouCui/go-bank/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// create token and add it to req auth header
func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

type AuthMiddlewareTestCase struct {
	name          string
	setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []AuthMiddlewareTestCase{
		// happy case
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// create a token for "user" valid for 1 minute, add bearer token to auth header
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		// no authorization header
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// nothing added to auth header
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		// unsupported authorization type
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// replace "bearer" with "unsupported"
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		// invalid authorization format
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// replace "bearer" with ""
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		// expired token
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// use negative token duration
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// generate sub-test using t.Run()
		t.Run(tc.name, func(t *testing.T) {
			// main content of sub-test
			// create a test server
			server := newTestServer(t, nil) // nil as we don't need db.Store for this test
			// add a fake api route with auth middleware to the test server
			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)
			// record the call
			recorder := httptest.NewRecorder()
			// create a request
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)
			// add auth header to the request
			tc.setupAuth(t, request, server.tokenMaker)
			// send request to the test server
			server.router.ServeHTTP(recorder, request)
			// verify result
			tc.checkResponse(t, recorder)
		})
	}
}
