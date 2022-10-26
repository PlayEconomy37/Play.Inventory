package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/PlayEconomy37/Play.Common/common"
	"github.com/PlayEconomy37/Play.Common/configuration"
	"github.com/PlayEconomy37/Play.Common/database"
	"github.com/PlayEconomy37/Play.Common/filters"
	"github.com/PlayEconomy37/Play.Common/logger"
	"github.com/PlayEconomy37/Play.Common/opentelemetry"
	"github.com/PlayEconomy37/Play.Common/permissions"
	"github.com/PlayEconomy37/Play.Common/types"
	"github.com/PlayEconomy37/Play.Inventory/internal/constants"
	"github.com/PlayEconomy37/Play.Inventory/internal/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// accessTokenUser1 is the access token for user with ID 1 and an expiry date of 100 years from now
	accessTokenUser1 = "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjQ0NDUiLCJzdWIiOiIxIiwiYXVkIjpbImh0dHA6Ly9sb2NhbGhvc3Q6MzAwMCJdLCJleHAiOjQ4MjA3MjYwMTUuODExNTI1LCJuYmYiOjE2NjUwNDg4MTUuODExNTI1MywiaWF0IjoxNjY1MDQ4ODE1LjgxMTUyNTN9.aLsLBwlj_1_Ie2h4PRmtv7d6fEftLaqlmh_0jWFsoAu8TGNxNV_-33CqMkxlF_sInYlUja8I0cmXSvEMUtwgxrkOIoOtzV-kJ9z5QWu8kGMGD4-cmAgxeeM2Ml9XB4oygjU8x0rz9P1Hg2RnOKkJQV2tbIzC8cb2t3CIIOzLQAofwX9dBYGvwrFJ0eX3gcP9BP6j0jj6Eei8JS0fF0bMEJAmbkPTu2UklYLKW_23ZlRfaQWZlezbf96eqfBkc_SlqQCuxnhPYjYaIX_a32NCP7cpNT8Kpgxu5T6gdRGA6ebWpRBFlQDIx_SKG-6W62ffkz6jkLAx_U7w9yIlH0fIiYbumsWxZwXVSGMtCI4IwjDNuuRLKxmqyo-82g5s03Y85jMuz4UNJaVYXKEC7mZYYNaCDwYULukWtca4hWba61FGRIKagHcLdWUHt6hn0K3uZW8_pliAcHM7qwFfOr9uVBNb6R5wseGyFRgV_zZuf4v269l6hCRxSDS9yabPsTup9k5HdbRUm-6oMH9XFW-bzuywuZs_MajlMaYMPU_kzVICGmGnoKaGc3W_79PhAXk6cYBeTACXrI8pkqZ0qoypI5A3WJENWllB2MG-mRT-M3vh-svQVHSf1GIW8gfr5YAy7JEIdS_-DKLeE8oOqbiTW66MuTwbGjfhT8Br1aH-GG8"

	// accessTokenUser1 is the access token for user with ID 2 and an expiry date of 100 years from now
	accessTokenUser2 = "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjQ0NDUiLCJzdWIiOiIyIiwiYXVkIjpbImh0dHA6Ly9sb2NhbGhvc3Q6MzAwMCJdLCJleHAiOjQ4MjA3MjU5MjMuMzY5NzMyLCJuYmYiOjE2NjUwNDg3MjMuMzY5NzMyMSwiaWF0IjoxNjY1MDQ4NzIzLjM2OTczMjF9.U6_jPV1K79TZ8ksawF0-8QS7TKYrm-jyc-2g_x0uKNZIV2NH3LEIwVTj2zOHewUIvulH5l-if-XKSSx3LZ8AMepsS6jj6F5Tw8zIoegTQTMIM6pjKZftzLJ_uexfyS6wWYI4xMyr5bvCpvQu4r98MbcIvwTgWwsPqrMbjNqdSU_bQLVwWNP9DpADuf3iV6mE9g2tlNVYVeFc5TMUa6QjAomr9P2gN6yRJc43A_stVJPWxY0CWn7dM_cqf_RiOOJGmVpxePcoZ9OShPd0QqdYnU2QbIwUXAfgUyi8PFiTZNGYxSHOgb3ul-nj55l9H_dB2lE98fTcZmSRbKhmTRR5H-ohlRTER5C7ijhmsUBHdLUl_5mhqCh0-tMZsbaTDaCalzhaDVN-dHr49Hl3JBiiIg3IAx6KNxIWF0DiCvgNT_cgJqfiHf6fFYVZGkQnsoUglCHoPupsEPNXELxE2Iy5M-WGPxlsHh3sCUwPkPE52_XkkAdwPQEWd8pfPuMNcYweGZe9SwcGAXygCYwg6x5-VlVfeb47brEEgmTkCPtwcdmeTC9DqcRI4nZb6EUQon28_CWo5XIbqtD-tevZn0FZWrR8Fc7LAeAZkxKwIHd5JtTzJVlP0hLCxuNCqHqogUOhagDPa2hK1Z9PK_23e4GEid68d0oXauRpbBNzW_nhf2U"

	// accessTokenUser1 is the access token for user with ID 3 and an expiry date of 100 years from now
	accessTokenUser3 = "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjQ0NDUiLCJzdWIiOiIzIiwiYXVkIjpbImh0dHA6Ly9sb2NhbGhvc3Q6MzAwMCJdLCJleHAiOjQ4MjA3MjYwMTUuODExNTI1LCJuYmYiOjE2NjUwNDg4MTUuODExNTI1MywiaWF0IjoxNjY1MDQ4ODE1LjgxMTUyNTN9.OCCRdwpdoj3PP4XmxeDFldsMSd-bSblvTvI0pNJwRbN-zg5xQWWCvYtflxah_-KHgCsyfVVRY-5MGG6sNZLPQB50Z1aSglEvcxyncdz3uANALHZDy_HGUp79dEMCeIQZvQTTSobrSZJLXFByhFp5YLwM4yUR_kW3WIoewHjdjzz0PmbvXzs9Gh8Co_wbsPlWQHqax-3KrqV28GV9mlO0xJqwJ0wHvn93kGJzU9vA6giXkcU976GRxkcZgAzVR1MRshQ4wT0Dv1MlN5P5mipg6yS92TLoT6xjBavJhFM4Nm_TerBHISF0_SEKtZ7NnO5b1NkkPl5acozzm13FLJsj_rsc3xGzliUa9nOSFyJhukkGE4sV2hX7M7FKGxxhnct77pnrnT5amcabtTnc1CqjJb17coKTVFZVy7Ubojp-7Lx6SIZOT1d6QSLtEkK6bCrqJEWDbJsClQeVn-X1Z8ybNG3yn5qs-aOOGBZySk8MbIGczltr-feTdn4rr-rwpvLwavJJEqPPQ0TDUm8WLe8f_Sj4wqJUTK1oUo_A0FUC9GoLAoaZuuWwhbvWTRXoXTQUfDETBW_1hIql35mtulCiOfBPYrhURhkZMZkLfuapLl1BaWaUtGpyYmL3w3Qs8MKCp5JhB55mvb5dS9zhmZOF4WmBaEU3AI5jL8FnEg4sOtM"

	// invalidAccessToken is a JWT access token NOT generated by the Identity microservice
	invalidAccessToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	// TestDatabase is a constant that defines the name of the database we use when we run tests
	TestDatabase = constants.Database + "_test"
)

var tracerProvider = opentelemetry.SetupTracer(true)

// Create a newTestApplication helper which returns an instance of our
// application struct with some modifications for tests
func newTestApplication(t *testing.T) (*Application, func(), []primitive.ObjectID) {
	// Setup logger
	var output io.Writer
	logFlag := os.Getenv("TEST_LOG")

	if logFlag != "" {
		output = os.Stdout
	} else {
		output = io.Discard
	}

	logger := logger.New(output, logger.LevelInfo)

	// Read configuration
	config, err := configuration.LoadConfig("../../config/dev.json")
	if err != nil {
		t.Fatal(err, nil)
	}

	// Start MongoDB
	mongoClient, err := database.NewMongoClient(config)

	// Create "catalog_items" collection in test database
	err = data.CreateCatalogItemsCollection(mongoClient, TestDatabase)
	if err != nil {
		t.Fatal(err, nil)
	}

	// Create "inventory_items" collection in test database
	err = data.CreateInventoryItemsCollection(mongoClient, TestDatabase)
	if err != nil {
		t.Fatal(err, nil)
	}

	// Create "users" collection
	err = database.CreateUsersCollection(mongoClient, TestDatabase)
	if err != nil {
		logger.Fatal(err, nil)
	}

	// Create users and catalog items repositories
	usersRepository := database.NewMongoRepository[int64, database.User](mongoClient, TestDatabase, database.UsersCollection)
	catalogItemsRepository := database.NewMongoRepository[primitive.ObjectID, data.CatalogItem](mongoClient, TestDatabase, constants.CatalogItemsCollection)

	// Seed users
	seedUsersCollection(t, usersRepository)

	// Seed catalog items
	catalogItemIDs := seedCatalogItemsCollection(t, catalogItemsRepository)

	// Database cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Delete test database and disconnect from mongo
		mongoClient.Database(TestDatabase).Drop(ctx)

		if err = mongoClient.Disconnect(ctx); err != nil {
			t.Fatal(err, nil)
		}

		// Shutdown opentelemetry tracer
		if err := tracerProvider.Shutdown(ctx); err != nil {
			t.Error(err, nil)
		}
	}

	return &Application{
		App: common.App{
			Config: config,
			Logger: logger,
			Tracer: tracerProvider.Tracer(config.ServiceName),
		},
		InventoryItemsRepository: database.NewMongoRepository[primitive.ObjectID, data.InventoryItem](mongoClient, TestDatabase, constants.InventoryItemsCollection),
		CatalogItemsRepository:   catalogItemsRepository,
		UsersRepository:          usersRepository,
	}, cleanup, catalogItemIDs
}

// Define a custom testServer type which anonymously embeds a httptest.Server
// instance.
type testServer struct {
	*httptest.Server
}

// Create a newTestServer helper which initializes and returns a new instance
// of our custom testServer type
func newTestServer(t *testing.T, router http.Handler) *testServer {
	ts := httptest.NewServer(router)

	return &testServer{ts}
}

// makeRequest is a helper method that creates a request with the given method and body
func (ts *testServer) makeRequest(t *testing.T, method string, urlPath string, body map[string]any, useAuthHeader bool, accessToken string) (int, http.Header, []byte) {
	var requestBody io.Reader

	if len(body) != 0 {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}

		requestBody = bytes.NewBuffer(jsonBody)
	} else {
		requestBody = nil
	}

	// Create HTTP request
	req, err := http.NewRequest(method, ts.URL+urlPath, requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set Authorization header if `useAuthHeader` is true
	if useAuthHeader {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	}

	// Make PUT request to given route
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	// Read the response body
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Return the response status, headers and body
	return res.StatusCode, res.Header, resBody
}

// get is a helper method for sending GET requests to the test server
func (ts *testServer) get(t *testing.T, urlPath string, useAuthHeader bool, accessToken string) (int, http.Header, []byte) {
	return ts.makeRequest(t, "GET", urlPath, map[string]any{}, useAuthHeader, accessToken)
}

// post is a helper method for sending POST requests to the test server
func (ts *testServer) post(t *testing.T, urlPath string, body map[string]any, useAuthHeader bool, accessToken string) (int, http.Header, []byte) {
	return ts.makeRequest(t, "POST", urlPath, body, useAuthHeader, accessToken)
}

// seedCatalogItemsCollection inserts some catalog items into the database
func seedCatalogItemsCollection(t *testing.T, repository types.MongoRepository[primitive.ObjectID, data.CatalogItem]) []primitive.ObjectID {
	// Check if items are already in the database
	fetchedItems, _, err := repository.GetAll(context.Background(), bson.M{}, filters.Filters{Page: 1, PageSize: 20, Sort: "_id", SortSafelist: []string{"_id", "name"}})
	if err != nil {
		t.Fatal(err)
	}

	var itemIDS []primitive.ObjectID

	// Items are already in the database
	if len(fetchedItems) == 5 {
		return itemIDS
	}

	items := []data.CatalogItem{
		{Name: "Potion", Description: "Restores a small amount of health", Version: 1},
		{Name: "Ether", Description: "Restores a small amount of MP", Version: 1},
		{Name: "Antidote", Description: "Cures poison", Version: 1},
		{Name: "Hi-Potion", Description: "Restores a small moderate of health", Version: 1},
		{Name: "Mega Potion", Description: "Restores a small big of health", Version: 1},
	}

	for i := range items {
		id, err := repository.Create(context.Background(), items[i])
		if err != nil {
			t.Fatal(err)
		}

		itemIDS = append(itemIDS, *id)
	}

	return itemIDS
}

// seedInventoryItemsCollection inserts some inventory items into the database
func seedInventoryItemsCollection(t *testing.T, ts *testServer, repository types.MongoRepository[primitive.ObjectID, data.InventoryItem], catalogItemIDs []primitive.ObjectID) {
	body := map[string]any{}
	body["userID"] = 1
	body["catalogItemID"] = catalogItemIDs[0]
	body["quantity"] = 2

	ts.post(t, "/items", body, true, accessTokenUser1)

	body = map[string]any{}
	body["userID"] = 1
	body["catalogItemID"] = catalogItemIDs[1]
	body["quantity"] = 3

	ts.post(t, "/items", body, true, accessTokenUser1)

	body = map[string]any{}
	body["userID"] = 1
	body["catalogItemID"] = catalogItemIDs[2]
	body["quantity"] = 5

	ts.post(t, "/items", body, true, accessTokenUser1)
}

// seedUsersCollection inserts some users into the database
func seedUsersCollection(t *testing.T, repository types.MongoRepository[int64, database.User]) {
	// Check if users are already in the database
	fetchedUsers, _, err := repository.GetAll(context.Background(), bson.M{}, filters.Filters{Page: 1, PageSize: 20, Sort: "_id", SortSafelist: []string{"_id"}})
	if err != nil {
		t.Fatal(err)
	}

	// Users are already in the database
	if len(fetchedUsers) == 3 {
		return
	}

	users := []database.User{
		{ID: 1, Permissions: permissions.Permissions{"inventory:read", "inventory:write"}, Activated: true, Version: 2},
		{ID: 2, Permissions: permissions.Permissions{"inventory:read"}, Activated: true, Version: 2},
		{ID: 3, Permissions: permissions.Permissions{"catalog:read"}, Activated: true, Version: 2},
	}

	for i := range users {
		_, err := repository.Create(context.Background(), users[i])
		if err != nil {
			t.Fatal(err)
		}
	}
}
