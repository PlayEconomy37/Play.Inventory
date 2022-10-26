package data

import (
	"context"

	"github.com/PlayEconomy37/Play.Inventory/internal/constants"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CatalogItem is a struct that defines a catalog item in our application
type CatalogItem struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Version     int32              `json:"version" bson:"version"`
}

// GetID returns the id of a catalog item.
// This method is necessary for our generic constraint of our mongo repository.
func (i CatalogItem) GetID() primitive.ObjectID {
	return i.ID
}

// GetVersion returns the version of a catalog item.
// This method is necessary for our generic constraint of our mongo repository.
func (i CatalogItem) GetVersion() int32 {
	return i.Version
}

// SetVersion sets the version of a catalog item to the given value and returns the catalog item.
// This method is necessary for our generic constraint of our mongo repository.
func (i CatalogItem) SetVersion(version int32) CatalogItem {
	i.Version = version

	return i
}

// CreateCatalogItemsCollection creates catalog items collection in MongoDB database
func CreateCatalogItemsCollection(client *mongo.Client, databaseName string) error {
	db := client.Database(databaseName)

	// JSON validation schema
	jsonSchema := bson.M{
		"bsonType":             "object",
		"required":             []string{"name", "description", "version"},
		"additionalProperties": false,
		"properties": bson.M{
			"_id": bson.M{
				"bsonType":    "objectId",
				"description": "Document ID",
			},
			"name": bson.M{
				"bsonType":    "string",
				"description": "Name of the item",
			},
			"description": bson.M{
				"bsonType":    "string",
				"description": "Description of the item",
			},
			"version": bson.M{
				"bsonType":    "int",
				"minimum":     1,
				"description": "Document version",
			},
		},
	}

	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}

	// Create collection
	opts := options.CreateCollection().SetValidator(validator)
	err := db.CreateCollection(context.Background(), constants.CatalogItemsCollection, opts)
	if err != nil {
		// Returns error if collection already exists so we ignore it
		return nil
	}

	// Create unique and text indexes
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.M{"description": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"name": "text"},
		},
	}

	_, err = db.Collection(constants.CatalogItemsCollection).Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		return err
	}

	return nil
}
