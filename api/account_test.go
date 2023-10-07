package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/XiaozhouCui/go-bank/db/mock"
	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/XiaozhouCui/go-bank/db/util"
	"github.com/XiaozhouCui/go-bank/token"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	// generate a random account
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// use an anonymous class to store test data
	testCases := []struct {
		name          string                                                            // unique name
		accountID     int64                                                             // unique account ID
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) // setup authentication for each test case
		buildStubs    func(store *mockdb.MockStore)                                     // build different stub for each case
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)           // make different assertion for each case
	}{
		// scenario for happy path
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// create a token for current user valid for 1 minute, add bearer token to auth header
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		// scenario for unauthorized user
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// use an "unauthorized_user" token to make request
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		// scenario for no authorization
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// no authorization header
			},
			buildStubs: func(store *mockdb.MockStore) {
				// handler should NOT be called
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		// scenario for Not Found
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// return an empty account
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// assertion: StatusNotFound (404)
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		// scenario for Internal Error
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// return ErrConnDone
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// assertion: StatusInternalServerError (500)
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		// scenario for Bad Request
		{
			name:      "InavlidID",
			accountID: 0, // invalid addountID
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// call GetAccount 0 time and don't return anything
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// assertion: StatusInternalServerError (400)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	// iterate through test cases
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// use generated mockstore
			store := mockdb.NewMockStore(ctrl)

			// build stub
			tc.buildStubs(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()                        // will record response as HTTP response writer
			url := fmt.Sprintf("/accounts/%d", tc.accountID)          // different accountID depending on test cases
			request, err := http.NewRequest(http.MethodGet, url, nil) // nil for GET request body
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker) // setup authentication for current test case
			server.router.ServeHTTP(recorder, request)  // send request through router and record response
			tc.checkResponse(t, recorder)               // make assertion on response
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

// make assertion on response body
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	// read all data from response body
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	// unmarshal data to the gotAccount object
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
