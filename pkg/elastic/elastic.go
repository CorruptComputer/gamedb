package elastic

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/olivere/elastic/v7"
)

const (
	IndexAchievements = "achievements"
	IndexGroups       = "groups"
	IndexApps         = "apps"
	IndexPlayers      = "players"
)

var (
	settings                = map[string]interface{}{"number_of_shards": 1, "number_of_replicas": 0}
	fieldTypeInteger        = map[string]interface{}{"type": "integer"}
	fieldTypeHalfFloat      = map[string]interface{}{"type": "half_float"}
	fieldTypeLong           = map[string]interface{}{"type": "long"}
	fieldTypeBool           = map[string]interface{}{"type": "boolean"}
	fieldTypeKeyword        = map[string]interface{}{"type": "keyword"}
	fieldTypeTextWithPrefix = map[string]interface{}{"type": "text", "index_prefixes": map[string]interface{}{"min_chars": 1, "max_chars": 10}}
	fieldTypeText           = map[string]interface{}{"type": "text"}
	fieldTypeDisabled       = map[string]interface{}{"enabled": false}
)

var (
	client *elastic.Client
	ctx    context.Context
	lock   sync.Mutex
)

func GetElastic() (*elastic.Client, context.Context, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error

	if client == nil {

		ctx = context.Background()

		ops := []elastic.ClientOptionFunc{
			elastic.SetURL(config.Config.ElasticAddress.Get()),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(true),
			elastic.SetBasicAuth(config.Config.ElasticUsername.Get(), config.Config.ElasticPassword.Get()),
		}

		client, err = elastic.NewClient(ops...)
	}

	return client, ctx, err
}

func indexDocument(index string, key string, doc interface{}) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	_, err = client.Index().Index(index).Id(key).BodyJson(doc).Do(ctx)
	return err
}

func indexDocuments(index string, docs map[string]interface{}) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	bulk := client.Bulk()
	for key, doc := range docs {
		bulk.Add(elastic.NewBulkIndexRequest().Index(index).Id(key).Doc(doc))
	}

	resp, err := bulk.Do(ctx)
	if err != nil {
		return err
	}

	failed := resp.Failed()
	if len(failed) > 0 {
		return errors.New(failed[0].Error.Reason)
	}

	return nil
}

func DeleteDocument(index string, key string) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	_, err = client.Delete().Index(index).Id(key).Do(ctx)
	return err
}

func rebuildIndex(index string, mapping map[string]interface{}) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	_, err = client.DeleteIndex(index).Do(ctx)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	createIndexResp, err := client.CreateIndex(index).BodyJson(mapping).Do(ctx)
	if err != nil {
		return err
	}

	if !createIndexResp.Acknowledged {
		return errors.New("not acknowledged")
	}

	return nil
}
