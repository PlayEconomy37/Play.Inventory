package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/PlayEconomy37/Play.Common/database"
	"github.com/PlayEconomy37/Play.Common/filters"
	"github.com/PlayEconomy37/Play.Common/types"
	"github.com/PlayEconomy37/Play.Common/validator"
	"github.com/PlayEconomy37/Play.Inventory/internal/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// healthCheckHandler is the handler for the "GET /healthcheck" endpoint
func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	env := types.Envelope{
		"status": "available",
	}

	err := app.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

// getInventoryItemsHandler is the handler for the "GET /items" endpoint
func (app *Application) getInventoryItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Create trace for the handler
	ctx, span := app.Tracer.Start(r.Context(), "Retrieving inventory items")
	defer span.End()

	// Anonymous struct used to hold the expected values from the request's query string
	var input struct {
		userID int64
		filters.Filters
	}

	// Instantiate validator
	v := validator.New()

	// Read query string
	queryString := r.URL.Query()

	// Extract values from query string if they exist
	input.userID = int64(app.ReadIntFromQueryString(queryString, "user_id", 0, v))
	input.Filters.Page = app.ReadIntFromQueryString(queryString, "page", 1, v)
	input.Filters.PageSize = app.ReadIntFromQueryString(queryString, "page_size", 20, v)
	input.Filters.Sort = app.ReadStringFromQueryString(queryString, "sort", "_id")

	// Add the supported sort values for this endpoint to the sort safelist
	input.Filters.SortSafelist = []string{"_id", "quantity", "acquiredDate", "-_id", "-quantity", "-acquiredDate"}

	// Validate user id and filters
	v.Check(input.userID > 0, "user_id", "must be greater than 0")
	filters.ValidateFilters(v, input.Filters)

	// Check the Validator instance for any errors
	if v.HasErrors() {
		span.SetStatus(codes.Error, "Validation failed")
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	span.SetAttributes(attribute.Int64("userID", input.userID))

	// Set filter
	filter := bson.M{}

	filter["user_id"] = bson.M{"$eq": input.userID}

	// Retrieve all inventory items
	inventoryItems, metadata, err := app.InventoryItemsRepository.GetAll(ctx, filter, input.Filters)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		app.ServerErrorResponse(w, r, err)
		return
	}

	// Collect catalog item ids from inventory items
	var itemIds []primitive.ObjectID

	for _, item := range inventoryItems {
		itemIds = append(itemIds, item.CatalogItemID)
	}

	// Set filter
	filter = bson.M{}

	filter["_id"] = bson.M{"$in": itemIds}

	// Retrieve all catalog items using collected item ids.
	// We use default filters otherwise this causes unexpected errors.
	catalogItems, _, err := app.CatalogItemsRepository.GetAll(ctx, filter, filters.Filters{Page: 1, PageSize: 20, Sort: "_id", SortSafelist: []string{"_id"}})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		app.ServerErrorResponse(w, r, err)
		return
	}

	type fullInventoryItem struct {
		ID            primitive.ObjectID `json:"id"`
		UserID        int64              `json:"userID"`
		CatalogItemID primitive.ObjectID `json:"catalogItemID"`
		Name          string             `json:"name"`
		Description   string             `json:"description"`
		Quantity      int64              `json:"quantity"`
	}

	var items []fullInventoryItem

	for _, inventoryItem := range inventoryItems {
		for _, catalogItem := range catalogItems {
			if catalogItem.ID == inventoryItem.CatalogItemID {
				item := fullInventoryItem{
					Name:          catalogItem.Name,
					Description:   catalogItem.Description,
					ID:            inventoryItem.ID,
					UserID:        inventoryItem.UserID,
					CatalogItemID: inventoryItem.CatalogItemID,
					Quantity:      inventoryItem.Quantity,
				}

				items = append(items, item)
			}
		}
	}

	env := types.Envelope{
		"items":    items,
		"metadata": metadata,
	}

	// Send back response
	err = app.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		app.ServerErrorResponse(w, r, err)
	}
}

// grantItemsHandler is the handler for the "POST /items" endpoint
func (app *Application) grantItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Create trace for the handler
	ctx, span := app.Tracer.Start(r.Context(), "Granting inventory items")
	defer span.End()

	// Declare an anonymous struct to hold the information that we expect to be in the
	// request body. This struct will be our *target decode destination*
	var input struct {
		UserID        int64              `json:"userID"`
		CatalogItemID primitive.ObjectID `json:"catalogItemID"`
		Quantity      int64              `json:"quantity"`
	}

	// Read request body and decode it into the input struct
	err := app.ReadJSON(w, r, &input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		app.BadRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new Item struct
	item := data.InventoryItem{
		UserID:        input.UserID,
		CatalogItemID: input.CatalogItemID,
		Quantity:      input.Quantity,
		Version:       1,
		AcquiredDate:  time.Now().UTC(),
		MessageIds:    []primitive.ObjectID{},
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Perform validation checks
	data.ValidateInventoryItem(v, item)

	if v.HasErrors() {
		span.SetStatus(codes.Error, "Validation failed")
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Record item attributes in trace
	span.SetAttributes(
		attribute.Int64("userID", item.UserID),
		attribute.String("catalogItemID", item.CatalogItemID.Hex()),
		attribute.Int64("quantity", item.Quantity),
	)

	// Set filters
	filter := bson.M{}

	filter["user_id"] = bson.M{"$eq": item.UserID}
	filter["catalog_item_id"] = bson.M{"$eq": item.CatalogItemID}

	inventoryItem, err := app.InventoryItemsRepository.GetByFilter(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			break
		default:
			app.ServerErrorResponse(w, r, err)
			return
		}
	}

	if inventoryItem.ID == primitive.NilObjectID {
		// Create a record in the database
		_, err := app.InventoryItemsRepository.Create(ctx, item)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			app.ServerErrorResponse(w, r, err)
			return
		}
	} else {
		// Update record in the database
		inventoryItem.Quantity = inventoryItem.Quantity + item.Quantity

		err = app.InventoryItemsRepository.Update(ctx, inventoryItem)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			app.ServerErrorResponse(w, r, err)
			return
		}
	}

	env := types.Envelope{
		"message": "Item granted successfully",
	}

	err = app.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		app.ServerErrorResponse(w, r, err)
	}
}
