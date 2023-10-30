package config

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
	"os"
	"sync"
)

var firestoreClient *firestore.Client
var once sync.Once

func GetFirestoreClient() *firestore.Client {
	once.Do(func() {
		firestoreClient = createClient(context.Background())
	})
	return firestoreClient
}

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := os.Getenv("PROJECT_ID")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}
