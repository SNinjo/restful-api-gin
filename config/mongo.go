package config

import (
	"log"
	"os"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitializeMongoDB() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI is not set in the environment")
	}

	if err := mgm.SetDefaultConfig(nil, "restful-api", options.Client().ApplyURI(uri)); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
}
