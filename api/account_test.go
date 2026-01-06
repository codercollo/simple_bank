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

	"github.com/codercollo/simple_bank/db/mock"
	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/token"
	"github.com/codercollo/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestCreateAccountAPI tests POST /accounts endpoint
func TestCreateAccountAPI(t *testing.T) {
	//Generate a random user and account for testing
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

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
				// "owner":account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mock.MockStore) {
				//Expect account creation with valid params
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  0,
				}
				//Expect CreateAccount to be called once with correct params
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
				//Store should not be called for invalid input
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

			//Add authorization header
			tokenMaker := server.tokenMaker
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)

			//Send request and verify response
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

// TestGetAccountAPI tests GET /accounts/:id endpoint
func TestGetAccountAPI(t *testing.T) {
	//Create test user and account
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	//Create a random account for test data

	//Define all test scenarios
	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//Add valid bearer token for the account owner
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//Token belongs to a different user than the account owner
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mock.MockStore) {
				//Account exists, but access should be denied
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 401 Unauthorized
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//No authorization header provided
			},
			buildStubs: func(store *mock.MockStore) {
				//Store must not be called when request is unauthenticated
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 401 Unauthorized
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//Valid token for account owner
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock.MockStore) {
				//Simulate account not existing in the database
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 404 Not Found
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//Valid token for account owner
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock.MockStore) {
				//Simulate databse connection error
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 500 Internal server error
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},

		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//Valid token but invalid account ID in URL
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mock.MockStore) {
				//Store should not be called for invalid ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				//Expect HTTP 400 Bad Request
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

			//Start test server with mock dependencies
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Build GET HTTP request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			//Attach authorization if required
			tc.setupAuth(t, request, server.tokenMaker)

			//Send request through router
			server.router.ServeHTTP(recorder, request)

			//Validate response
			tc.checkResponse(t, recorder)

		})

	}

}

// TestListAccountAPI tests GET /accounts endpoint
// func TestListAccountAPI(t *testing.T) {
// 	user, _ := randomUser(t)

// 	//Generate test accounts
// 	accounts := []db.Account{
// 		randomAccount(),
// 		randomAccount(),
// 		randomAccount(),
// 	}

// 	//Define test cases
// 	testCases := []struct {
// 		name          string
// 		query         string
// 		buildStubs    func(store *mock.MockStore)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name:  "OK",
// 			query: "?page_id=1&page_size=5",
// 			buildStubs: func(store *mock.MockStore) {
// 				//Expect ListAccounts to be called once
// 				store.EXPECT().
// 					ListAccounts(gomock.Any(), gomock.Any()).
// 					Times(1).
// 					Return(accounts, nil)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 200 OK
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name:  "InvalidQuery",
// 			query: "?page_id=0&page_size=5",
// 			buildStubs: func(store *mock.MockStore) {
// 				//Store should not be called
// 				store.EXPECT().
// 					ListAccounts(gomock.Any(), gomock.Any()).
// 					Times(0)

// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 400 Bad Request
// 				require.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},

// 		{
// 			name:  "InternalError",
// 			query: "?page_id=1&page_size=5",
// 			buildStubs: func(store *mock.MockStore) {
// 				//Simulate database error
// 				store.EXPECT().
// 					ListAccounts(gomock.Any(), gomock.Any()).
// 					Times(1).
// 					Return([]db.Account{}, sql.ErrConnDone)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 500 Internal Server Error
// 				require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 	}

// 	//Run all test cases
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.name, func(t *testing.T) {
// 			//Setup gomock controller
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			//Initialize mock store
// 			store := mock.NewMockStore(ctrl)
// 			tc.buildStubs(store)

// 			//Start test server
// 			server := newTestServer(t, store)
// 			recorder := httptest.NewRecorder()

// 			//Create HTTP request
// 			url := "/accounts" + tc.query
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			require.NoError(t, err)

// 			//Send request and verify  response
// 			server.router.ServeHTTP(recorder, request)
// 			tc.checkResponse(t, recorder)
// 		})
// 	}
// }

// randomAccount generates a random account for testing
func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
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
// func TestUpdateAccountAPI(t *testing.T) {
// 	//Generate test account
// 	account := randomAccount()
// 	newBalance := util.RandomMoney()

// 	//Define test cases
// 	testCases := []struct {
// 		name          string
// 		accountID     int64
// 		body          gin.H
// 		buildStubs    func(store *mock.MockStore)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name:      "OK",
// 			accountID: account.ID,
// 			body: gin.H{
// 				"balance": newBalance,
// 			},
// 			buildStubs: func(store *mock.MockStore) {
// 				//Expect successful account update
// 				arg := db.UpdateAccountParams{
// 					ID:      account.ID,
// 					Balance: newBalance,
// 				}

// 				updateAccount := account
// 				updateAccount.Balance = newBalance

// 				store.EXPECT().
// 					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return(updateAccount, nil)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 200 OK
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name:      "InvalidID",
// 			accountID: 0,
// 			body: gin.H{
// 				"balance": newBalance,
// 			},
// 			buildStubs: func(store *mock.MockStore) {
// 				//Store should not be called
// 				store.EXPECT().
// 					UpdateAccount(gomock.Any(), gomock.Any()).
// 					Times(0)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 400 Bad Request
// 				require.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name:      "NotFound",
// 			accountID: account.ID,
// 			body: gin.H{
// 				"balance": newBalance,
// 			},
// 			buildStubs: func(store *mock.MockStore) {
// 				//Simulate account not found
// 				store.EXPECT().
// 					UpdateAccount(gomock.Any(), gomock.Any()).
// 					Times(1).
// 					Return(db.Account{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 404 Not Found
// 				require.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name:      "InternalError",
// 			accountID: account.ID,
// 			body: gin.H{
// 				"balance": newBalance,
// 			},
// 			buildStubs: func(store *mock.MockStore) {
// 				//Simulate database error
// 				store.EXPECT().
// 					UpdateAccount(gomock.Any(), gomock.Any()).
// 					Times(1).
// 					Return(db.Account{}, sql.ErrConnDone)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 500 Internal Server Error
// 				require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 	}

// 	//Run all test cases
// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			//Setup gomock controller
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			//Initialize mock store
// 			store := mock.NewMockStore(ctrl)
// 			tc.buildStubs(store)

// 			//Start test server
// 			server := newTestServer(t, store)
// 			recorder := httptest.NewRecorder()

// 			//Encode request body
// 			data, err := json.Marshal(tc.body)
// 			require.NoError(t, err)

// 			//Create HTTP request
// 			url := fmt.Sprintf("/accounts/%d", tc.accountID)
// 			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
// 			require.NoError(t, err)

// 			//Send request and verify response
// 			server.router.ServeHTTP(recorder, request)
// 			tc.checkResponse(t, recorder)

// 		})
// 	}

// }

// // TestDeleteAccountAPI tests DELETE /accounts/:id endpoint
// func TestDeleteAccountAPI(t *testing.T) {
// 	//Generate test account
// 	account := randomAccount()

// 	//Define test cases
// 	testCases := []struct {
// 		name          string
// 		accountID     int64
// 		buildStubs    func(store *mock.MockStore)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name:      "OK",
// 			accountID: account.ID,
// 			buildStubs: func(store *mock.MockStore) {
// 				//Expect successful deletion
// 				store.EXPECT().
// 					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
// 					Times(1).
// 					Return(nil)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Ecpect 200 OK
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name:      "InvalidID",
// 			accountID: 0,
// 			buildStubs: func(store *mock.MockStore) {
// 				//Store should not be called
// 				store.EXPECT().
// 					DeleteAccount(gomock.Any(), gomock.Any()).
// 					Times(0)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 400 Bad Request
// 				require.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name:      "NotFound",
// 			accountID: account.ID,
// 			buildStubs: func(store *mock.MockStore) {
// 				//Simulate account not found
// 				store.EXPECT().
// 					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
// 					Times(1).
// 					Return(sql.ErrNoRows)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				//Expect 404 Not Found
// 				require.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name:      "InternalError",
// 			accountID: account.ID,
// 			buildStubs: func(store *mock.MockStore) {
// 				//Simulate database error
// 				store.EXPECT().
// 					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
// 					Times(1).
// 					Return(sql.ErrConnDone)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 	}

// 	//Run all test cases
// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			//Setup gomock controller
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			//Initialize mock store
// 			store := mock.NewMockStore(ctrl)
// 			tc.buildStubs(store)

// 			//Start test server
// 			server := newTestServer(t, store)
// 			recorder := httptest.NewRecorder()

// 			//Create HTTP request
// 			url := fmt.Sprintf("/accounts/%d", tc.accountID)
// 			request, err := http.NewRequest(http.MethodDelete, url, nil)
// 			require.NoError(t, err)

// 			//Send request and verify response
// 			server.router.ServeHTTP(recorder, request)
// 			tc.checkResponse(t, recorder)
// 		})
// 	}
// }
