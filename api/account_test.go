package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/XiaozhouCui/go-bank/db/mock"
	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/XiaozhouCui/go-bank/db/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// use generated mockstore
	store := mockdb.NewMockStore(ctrl)

	// build stubs
	store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)

	// start test server and send request
	server := NewServer(store)
	recorder := httptest.NewRecorder() // will record response as HTTP response writer

	url := fmt.Sprintf("/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodGet, url, nil) // nil for GET request body
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request) // send request through router and record response

	// make assertion on response
	require.Equal(t, http.StatusOK, recorder.Code)
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}
