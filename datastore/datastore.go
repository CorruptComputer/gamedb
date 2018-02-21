// Docs: https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/example_test.go

package datastore

import (
	"context"
	"os"

	"cloud.google.com/go/datastore"
)

const (
	// CHANGE datastore kind
	CHANGE = "Change"

	// APP datastore kind
	//APP = "App"

	// PACKAGE datastore kind
	//PACKAGE = "Package"

	// ARTICLE is a news article for an app
	ARTICLE = "Article"

	// PLAYER datastore kind
	PLAYER = "Player"

	// RANK datastore kind
	RANK = "Rank"
)

func getDSClient() (client *datastore.Client, ctx context.Context, err error) {

	ctx = context.Background()
	client, err = datastore.NewClient(ctx, os.Getenv("STEAM_GOOGLE_PROJECT"))
	if err != nil {
		return client, ctx, err
	}

	return client, ctx, nil
}

func SaveKind(key *datastore.Key, data interface{}) (newKey *datastore.Key, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return nil, err
	}

	newKey, err = client.Put(ctx, key, data)
	if err != nil {
		return newKey, err
	}

	return newKey, nil
}
