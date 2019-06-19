package mongodb

import (
	"context"
	"github.com/lt90s/rfschub-server/common/store/mongodb"
	"github.com/lt90s/rfschub-server/index/config"
	"github.com/lt90s/rfschub-server/index/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type mongodbStore struct {
	client  *mongo.Client
	name    string
	timeout int64
}

func NewMongodbStore() store.Store {
	conf := config.DefaultConfig
	client := mongodb.NewClient(conf.Mongodb.Uri)

	ms := &mongodbStore{
		client:  client,
		name:    conf.Mongodb.Database,
		timeout: int64(conf.Timeout),
	}
	return ms
}

func (ms *mongodbStore) database() *mongo.Database {
	return ms.client.Database(ms.name)
}

func (ms *mongodbStore) taskCollection() *mongo.Collection {
	return ms.database().Collection("tasks")
}

func (ms *mongodbStore) fileIndexCollection() *mongo.Collection {
	return ms.database().Collection("file_indexes")
}

func (ms *mongodbStore) NewIndexTask(ctx context.Context, url, hash string) error {
	now := time.Now().Unix()
	filter := bson.M{
		"url":  url,
		"hash": hash,
	}
	update := bson.M{
		"$set": bson.M{
			"createdAt": now,
		},
	}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}

	rs, err := ms.taskCollection().UpdateOne(ctx, filter, update, option)
	if err != nil {
		return err
	}
	if rs.ModifiedCount != 1 && rs.UpsertedCount != 1 {
		return store.ErrIndexTaskExist
	}
	return nil
}

// set task state success or failure
func (ms *mongodbStore) SetTaskState(ctx context.Context, url, hash string, success bool) error {
	filter := bson.M{
		"url":  url,
		"hash": hash,
	}
	update := bson.M{}
	if success {
		update["$set"] = bson.M{
			"success": true,
		}
	} else {
		return nil
	}
	_, err := ms.taskCollection().UpdateOne(ctx, filter, update)
	return err
}

func (ms *mongodbStore) RepositoryIndexed(ctx context.Context, url, hash string) (bool, error) {
	filter := bson.M{
		"url":  url,
		"hash": hash,
	}
	option := &options.FindOneOptions{
		Projection: bson.M{
			"success": 1,
		},
	}
	sr := ms.taskCollection().FindOne(ctx, filter, option)
	if err := sr.Err(); err != nil {
		return false, sr.Err()
	}
	var tmp struct {
		Success bool `bson:"success"`
	}
	if err := sr.Decode(&tmp); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return tmp.Success, nil
}
func (ms *mongodbStore) RepositoryFileIndexed(ctx context.Context, url, hash, file string) (bool, error) {
	filter := bson.M{
		"url":  url,
		"hash": hash,
		"file": file,
	}
	count, err := ms.fileIndexCollection().CountDocuments(ctx, filter)
	return count > 0, err
}

// add all index symbols of a file
func (ms *mongodbStore) AddFileIndexEntries(ctx context.Context, entries []store.IndexEntry) error {
	documents := make([]interface{}, len(entries))
	for idx := range entries {
		documents[idx] = entries[idx]
	}
	_, err := ms.fileIndexCollection().InsertMany(ctx, documents)
	return err
}

func (ms *mongodbStore) FindSymbols(ctx context.Context, url, hash, name string) (symbols []store.Symbol, err error) {
	filter := bson.M{
		"url":  url,
		"hash": hash,
		"name": name,
	}
	option := &options.FindOptions{
		Projection: bson.M{
			"url":     0,
			"hash":    0,
			"name":    0,
			"pattern": 0,
		},
	}

	cursor, err := ms.fileIndexCollection().Find(ctx, filter, option)
	if err != nil {
		return
	}

	symbols = make([]store.Symbol, 0, 8)
	var symbol store.Symbol
	for cursor.Next(ctx) {
		if err = cursor.Err(); err != nil {
			return
		}
		if err = cursor.Decode(&symbol); err != nil {
			return
		}
		symbols = append(symbols, symbol)
	}
	return
}
