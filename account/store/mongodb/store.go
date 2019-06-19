package mongodb

import (
	"context"
	"github.com/lt90s/rfschub-server/account/config"
	"github.com/lt90s/rfschub-server/account/store"
	"github.com/lt90s/rfschub-server/common/store/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type mongodbStore struct {
	client *mongo.Client
	name   string
}

const (
	accountCollection = "accounts"
)

func NewMongodbStore() store.Store {
	client := mongodb.NewClient(config.DefaultConfig.Mongodb.Uri)

	ms := &mongodbStore{
		client: client,
		name:   config.DefaultConfig.Mongodb.Database,
	}
	ms.setup()
	return ms
}

func (ms *mongodbStore) accountCollection() *mongo.Collection {
	return ms.client.Database(ms.name).Collection(accountCollection)
}

func (ms *mongodbStore) setup() {
	iv := ms.accountCollection().Indexes()
	unique := true
	models := []mongo.IndexModel{
		{
			Keys: bson.M{"name": 1},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		}, {
			Keys: bson.M{"email": 1},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	}
	_, err := iv.CreateMany(context.Background(), models)
	if err != nil && !strings.Contains(err.Error(), "IndexKeySpecsConflict") {
		panic(err)
	}
}

func (ms *mongodbStore) CreateAccount(ctx context.Context, name, email, password string, code []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// TODO: account should be activate before being valid
	is, err := ms.accountCollection().InsertOne(ctx, bson.M{
		"name":      name,
		"email":     email,
		"hash":      hash,
		"createdAt": time.Now().Unix(),
		"code":      code,
		"activated": true, // defaults to activated for now
	})

	if err != nil {
		return "", err
	}

	return is.InsertedID.(primitive.ObjectID).Hex(), err
}

func (ms *mongodbStore) LoginAccount(ctx context.Context, name, email, password string) (info store.AccountInfo, err error) {
	filter := bson.M{}
	if name != "" {
		filter["name"] = name
	} else {
		filter["email"] = email
	}

	option := &options.FindOneOptions{
		Projection: bson.M{
			"_id":       1,
			"name":      1,
			"hash":      1,
			"createdAt": 1,
			"activated": 1,
		},
	}
	sr := ms.accountCollection().FindOne(ctx, filter, option)
	if err = sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			err = store.ErrNoMatch
		}
		return
	}

	var tmp struct {
		Id        primitive.ObjectID `bson:"_id"`
		Name      string             `bson:"name"`
		Hash      []byte             `bson:"hash"`
		CreatedAt int64              `bson:"createdAt"`
		Activated bool               `bson:"activated"`
	}

	if err = sr.Decode(&tmp); err != nil {
		return
	}

	if err = bcrypt.CompareHashAndPassword(tmp.Hash, []byte(password)); err != nil {
		err = store.ErrNoMatch
		return
	}

	if !tmp.Activated {
		err = store.ErrNoAccount
		return
	}

	info.Id = tmp.Id.Hex()
	info.Name = tmp.Name
	info.CreatedAt = tmp.CreatedAt
	return
}

func (ms *mongodbStore) GetAccountId(ctx context.Context, name string) (string, error) {
	filter := bson.M{
		"name": name,
	}
	option := &options.FindOneOptions{
		Projection: bson.M{
			"_id": 1,
		},
	}
	sr := ms.accountCollection().FindOne(ctx, filter, option)

	if sr.Err() != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return "", store.ErrNoAccount
		}
		return "", sr.Err()
	}

	var tmp struct {
		Id primitive.ObjectID `bson:"_id"`
	}
	if err := sr.Decode(&tmp); err != nil {
		return "", err
	}

	return tmp.Id.Hex(), nil
}

func (ms *mongodbStore) GetAccountsBasicInfo(ctx context.Context, uids []string) (infos []store.BasicInfo, err error) {
	ids := make([]primitive.ObjectID, 0, len(uids))
	for _, uid := range uids {
		id, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	option := &options.FindOptions{
		Projection: bson.M{
			"_id":  1,
			"name": 1,
		},
	}
	cursor, err := ms.accountCollection().Find(ctx, filter, option)

	if err != nil {
		return
	}
	var tmp struct {
		Id   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
	}
	for cursor.Next(ctx) {
		if err = cursor.Decode(&tmp); err != nil {
			return
		}
		infos = append(infos, store.BasicInfo{Id: tmp.Id.Hex(), Name: tmp.Name})
	}
	return
}

func (ms *mongodbStore) GetAccountInfoByName(ctx context.Context, name string) (info store.AccountInfo, err error) {
	filter := bson.M{
		"name": name,
	}

	sr := ms.accountCollection().FindOne(ctx, filter)
	if err = sr.Err(); err != nil {
		return
	}

	var tmp struct {
		Id      primitive.ObjectID `bson:"_id"`
		Created int64              `bson:"createdAt"`
	}

	if err = sr.Decode(&tmp); err != nil {
		if err == mongo.ErrNoDocuments {
			err = store.ErrNoAccount
		}
		return
	}
	info.Id = tmp.Id.Hex()
	info.Name = name
	info.CreatedAt = tmp.Created
	return
}
