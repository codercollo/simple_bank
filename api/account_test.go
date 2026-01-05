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

	"github.com/codercollo/simple_bank/db/mock"
	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestCreateAccountAPI tests POST /accounts endpoint
func TestCreateAccountAPI(t *testing.T) {
	//Generate test account data
	account := randomAccount()

	//Define test cases
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mock.MockStore) {
				//Expect account creation with valid params
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 200 OK and correct response body
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		}, {
			name: "InvalidBody",
			body: gin.H{
				"owner": account.Owner,
				//Missing currency
			},
			buildStubs: func(store *mock.MockStore) {
				//Store should not be called
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expecte 400 Bad request
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mock.MockStore) {
				//Simulate database error
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 500 Internal Server Error
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	//Run all test cases
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			//Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//Initialize mock store
			store := mock.NewMockStore(ctrl)
			tc.buildStubs(store)

			//Start test server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Encode request body
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			//Create HTTP request
			request, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(data))
			require.NoError(t, err)

			//Send request and verify response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

// TestGetAccountAPI tests GET /accounts/:id endpoint
func TestGetAccountAPI(t *testing.T) {
	//Create a random account for test data
	account := randomAccount()

	//Define all test scenarios
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mock.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mock.MockStore) {
				//Expect GetAccount to be called once and succeed
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Verify HTTP 200 and correct response body
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},

		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mock.MockStore) {
				//Simulate account not found
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 404
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mock.MockStore) {
				//Simulate databse connection error
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 500
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},

		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mock.MockStore) {
				//Store should not be called for invalid ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 400
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
	}

	//Execute all test cases
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			//Initialize gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//Create mock store and apply stubs
			store := mock.NewMockStore(ctrl)
			tc.buildStubs(store)

			//Start test server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Build HTTP request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			//Serve request and validate response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})

	}

}

// TestListAccountAPI tests GET /accounts endpoint
func TestLisAccountAPI(t *testing.T) {
	//Generate test accounts
	accounts := []db.Account{
		randomAccount(),
		randomAccount(),
		randomAccount(),
	}

	//Define test cases
	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mock.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?page_id=1&page_size=5",
			buildStubs: func(store *mock.MockStore) {
				//Expect ListAccounts to be called once
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 200 OK
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "InvalidQuery",
			query: "?page_id=0&page_size=5",
			buildStubs: func(store *mock.MockStore) {
				//Store should not be called
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 400 Bad Request
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name:  "InternalError",
			query: "?page_id=1&page_size=5",
			buildStubs: func(store *mock.MockStore) {
				//Simulate database error
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 500 Internal Server Error
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	//Run all test cases
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			//Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//Initialize mock store
			store := mock.NewMockStore(ctrl)
			tc.buildStubs(store)

			//Start test server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Create HTTP request
			url := "/accounts" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			//Send request and verify  response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// randomAccount generates a random account for testing
func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

}

// requireBodyMatchAccount validates response body against expected account
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	//Read response body
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	//Decode JSON response
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)

	//Compare expected and actual account
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)

}

// TestUpdateAccountAPI tests PUT /accounts/:id endpoint
func TestUpdateAccountAPI(t *testing.T) {
	//Generate test account
	account := randomAccount()
	newBalance := util.RandomMoney()

	//Define test cases
	testCases := []struct {
		name          string
		accountID     int64
		body          gin.H
		buildStubs    func(store *mock.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			body: gin.H{
				"balance": newBalance,
			},
			buildStubs: func(store *mock.MockStore) {
				//Expect successful account update
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}

				updateAccount := account
				updateAccount.Balance = newBalance

				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updateAccount, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 200 OK
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			body: gin.H{
				"balance": newBalance,
			},
			buildStubs: func(store *mock.MockStore) {
				//Store should not be called
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 400 Bad Request
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			body: gin.H{
				"balance": newBalance,
			},
			buildStubs: func(store *mock.MockStore) {
				//Simulate account not found
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 404 Not Found
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			body: gin.H{
				"balance": newBalance,
			},
			buildStubs: func(store *mock.MockStore) {
				//Simulate database error
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 500 Internal Server Error
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	//Run all test cases
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			//Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//Initialize mock store
			store := mock.NewMockStore(ctrl)
			tc.buildStubs(store)

			//Start test server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Encode request body
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			//Create HTTP request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			//Send request and verify response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})
	}

}

// TestDeleteAccountAPI tests DELETE /accounts/:id endpoint
func TestDeleteAccountAPI(t *testing.T) {
	//Generate test account
	account := randomAccount()

	//Define test cases
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mock.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mock.MockStore) {
				//Expect successful deletion
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Ecpect 200 OK
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mock.MockStore) {
				//Store should not be called
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 400 Bad Request
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mock.MockStore) {
				//Simulate account not found
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect 404 Not Found
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mock.MockStore) {
				//Simulate database error
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	//Run all test cases
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			//Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//Initialize mock store
			store := mock.NewMockStore(ctrl)
			tc.buildStubs(store)

			//Start test server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Create HTTP request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			//Send request and verify response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
