package storage

// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/storage/example_test.go
// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/golang/snappy"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/logger"
)

var (
	bucket = viper.GetString("GOOGLE_BUCKET")
	client *storage.Client
)

var (
	PathGames       = func(playerID int64) (string) { return "/player-games/" + strconv.FormatInt(playerID, 10) + ".json" }
	PathBadges      = func(playerID int64) (string) { return "/player-badges/" + strconv.FormatInt(playerID, 10) + ".json" }
	PathFriends     = func(playerID int64) (string) { return "/player-friends/" + strconv.FormatInt(playerID, 10) + ".json" }
	PathRecentGames = func(playerID int64) (string) { return "/player-recent-games/" + strconv.FormatInt(playerID, 10) + ".json" }
)

func getClient() (c *storage.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if client == nil {

		client, err = storage.NewClient(ctx)
		if err != nil {
			return client, ctx, nil
		}
	}

	return client, ctx, nil
}

func Upload(path string, data []byte, public bool) (err error) {

	// Encode
	data = snappy.Encode(nil, data)

	// Get client
	client, ctx, err := getClient()
	if err != nil {
		return err
	}

	//
	object := client.Bucket(bucket).Object(path)

	// Upload bytes
	wc := object.NewWriter(ctx)
	if _, err = io.Copy(wc, bytes.NewReader(data)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	// Make public
	if public {
		if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func Download(path string) (bytes []byte, err error) {

	// Get client
	client, ctx, err := getClient()
	if err != nil {
		return bytes, err
	}

	// Download
	rc, err := client.Bucket(bucket).Object(path).NewReader(ctx)
	if err != nil {
		return bytes, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return bytes, err
	}

	// Decode
	bytes, err = snappy.Decode(nil, data)
	if err != nil {
		logger.Error(err)
		// data is not encoded? Return as is.
		bytes = data
	}

	return bytes, nil
}
