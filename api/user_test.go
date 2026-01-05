package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/codercollo/simple_bank/db/mock"
	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// eqCreateUserParamsMatcher validates CreateUser params including hashed password
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

// Matches checks that the input matches expected params and password hash
func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	//Assert correct argument type
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	//Verify hashed password matches plaintext password
	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	//Align hashed password for deep equality check
	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)

}

// String provides readable matcher output for test failures
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg  %v and pasword %v", e.arg, e.password)
}

// EqCreateUserParams creates a custom gomock matcher for CreateUser arguments
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

// TestCreateUserAPI tests the POST /users endpoint using table-driven tests
func TestCreatedUserAPI(t *testing.T) {
	//Set Gin to test mode to avoid noisy logs
	gin.SetMode(gin.TestMode)

	//Generate a random valid user for test cases
	user, password := randomUser(t)

	//Define all test scenarios
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			//Expect CreateUser with validated arguments via custom matcher
			buildStubs: func(store *mock.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			//Verify HTTP 200 and response body
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			//Simulate databse connection error
			buildStubs: func(store *mock.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			//Expect HTTP 500
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			//Simulate unique constraint violation
			buildStubs: func(store *mock.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			//Expect HTTP 403 Forbidden
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			//Validation should fail before DB call
			buildStubs: func(store *mock.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			//Expect HTTP 400 Bad Request
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			//Validation should fail before DB call
			buildStubs: func(store *mock.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			//Expect HTTP 400 Bad Request
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			//Validation should fail before DB call
			buildStubs: func(store *mock.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			//Expect HTTP 400 Bad Request
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	//Execute each test case
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			//Initialize gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//Create mock store and build expectations
			store := mock.NewMockStore(ctrl)
			tc.buildStubs(store)

			//Initialize test server and response recorder
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//Marshal request body to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			//Create HTTP POST request
			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			//Serve the HTTP request
			server.router.ServeHTTP(recorder, request)

			//Validate response
			tc.checkResponse(recorder)
		})
	}
}

// randomUser generates a valid random user and plaintext password for testing
func randomUser(t *testing.T) (user db.User, password string) {
	//Generate random plaintext password
	password = util.RandomString(6)

	//Hash password for storage
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	//Build user model
	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

// requireBodyMatchUser verifies that the response body matches the expected user
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	//Read response body
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	//Unmarshal JSON into user struct
	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)

	//Compare returned fields
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)

	//Ensure password hash is not exposed
	require.Empty(t, gotUser.HashedPassword)

}
