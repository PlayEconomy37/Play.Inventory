package data

import (
	"context"
	"time"

	"github.com/PlayEconomy37/Play.Common/validator"
	"github.com/PlayEconomy37/Play.Inventory/internal/constants"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InventoryItem is a struct that defines an inventory item in our application
type InventoryItem struct {
	ID            primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	UserID        int64                `json:"userID" bson:"user_id"`
	CatalogItemID primitive.ObjectID   `json:"catalogItemID" bson:"catalog_item_id"`
	Quantity      int64                `json:"quantity" bson:"quantity"`
	Version       int32                `json:"version" bson:"version"`
	AcquiredDate  time.Time            `json:"-" bson:"acquired_date"`
	MessageIds    []primitive.ObjectID `json:"messageIDs" bson:"message_ids"`
}

// GetID returns the id of an inventory item.
// This method is necessary for our generic constraint of our mongo repository.
func (i InventoryItem) GetID() primitive.ObjectID {
	return i.ID
}

// GetVersion returns the version of an inventory item.
// This method is necessary for our generic constraint of our mongo repository.
func (i InventoryItem) GetVersion() int32 {
	return i.Version
}

// SetVersion sets the version of an inventory item to the given value and returns the inventory item.
// This method is necessary for our generic constraint of our mongo repository.
func (i InventoryItem) SetVersion(version int32) InventoryItem {
	i.Version = version

	return i
}

// ValidateInventoryItem runs validation checks on the `InventoryItem` struct
func ValidateInventoryItem(v *validator.Validator, item InventoryItem) {
	v.Check(item.UserID > 0, "userID", "must be greater than 0")
	v.Check(item.Quantity > 0, "quantity", "must be greater than 0")
}

// CreateInventoryItemsCollection creates inventory items collection in MongoDB database
func CreateInventoryItemsCollection(client *mongo.Client, databaseName string) error {
	db := client.Database(databaseName)

	// JSON validation schema
	jsonSchema := bson.M{
		"bsonType":             "object",
		"required":             []string{"user_id", "catalog_item_id", "quantity", "version", "acquired_date", "message_ids"},
		"additionalProperties": false,
		"properties": bson.M{
			"_id": bson.M{
				"bsonType":    "objectId",
				"description": "Document ID",
			},
			"user_id": bson.M{
				"bsonType":    "long",
				"description": "ID of user who owns the item",
			},
			"catalog_item_id": bson.M{
				"bsonType":    "objectId",
				"description": "ID of the catalog item",
			},
			"quantity": bson.M{
				"bsonType":    "long",
				"minimum":     1,
				"description": "Quantity of the inventory item",
			},
			"version": bson.M{
				"bsonType":    "int",
				"minimum":     1,
				"description": "Document version",
			},
			"acquired_date": bson.M{
				"bsonType":    "date",
				"description": "Date when item was acquired",
			},
			"message_ids": bson.M{
				"bsonType":    "array",
				"description": "Array of message broker message ids",
			},
		},
	}

	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}

	// Create collection
	opts := options.CreateCollection().SetValidator(validator)
	err := db.CreateCollection(context.Background(), constants.InventoryItemsCollection, opts)
	if err != nil {
		// Returns error if collection already exists so we ignore it
		return nil
	}

	return nil
}
