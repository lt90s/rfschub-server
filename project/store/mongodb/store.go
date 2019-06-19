package mongodb

import (
	"context"
	"github.com/lt90s/rfschub-server/common/store/mongodb"
	"github.com/lt90s/rfschub-server/project/config"
	"github.com/lt90s/rfschub-server/project/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type mongodbStore struct {
	client       *mongo.Client
	databaseName string
	taskTimeout  int64
}

const (
	projectCollection          = "projects"
	annotationCollection       = "annotations"
	latestAnnotationCollection = "latest_annotations"
)

func NewMongodbStore() store.Store {
	client := mongodb.NewClient(config.DefaultConfig.Mongodb.Uri)
	ms := &mongodbStore{
		client:       client,
		databaseName: config.DefaultConfig.Mongodb.Database,
	}
	return ms
}

func (ms *mongodbStore) database() *mongo.Database {
	return ms.client.Database(ms.databaseName)
}

func (ms *mongodbStore) projectCollection() *mongo.Collection {
	return ms.database().Collection(projectCollection)
}

func (ms *mongodbStore) annotationCollection() *mongo.Collection {
	return ms.database().Collection(annotationCollection)
}

func (ms *mongodbStore) latestAnnotationCollection() *mongo.Collection {
	return ms.database().Collection(latestAnnotationCollection)
}

func (ms *mongodbStore) NewProject(ctx context.Context, uid, url, hash, name string, branch bool) error {
	filter := bson.M{
		"uid":  uid,
		"url":  url,
		"hash": hash,
	}
	update := bson.M{
		"$set": bson.M{
			"name":   name,
			"branch": branch,
		},
		"$setOnInsert": bson.M{
			"createdAt": time.Now().Unix(),
			"indexed":   false,
		},
	}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}

	us, err := ms.projectCollection().UpdateOne(ctx, filter, update, option)
	if err != nil {
		return err
	}

	if us.UpsertedCount == 0 {
		return store.ErrProjectExist
	}
	return nil
}
func (ms *mongodbStore) GetProjectInfo(ctx context.Context, uid, url, name string) (info store.ProjectInfo, err error) {
	filter := bson.M{
		"uid":  uid,
		"url":  url,
		"name": name,
	}
	option := &options.FindOneOptions{
		Projection: bson.M{
			"_id":     1,
			"hash":    1,
			"branch":  1,
			"indexed": 1,
		},
	}
	sr := ms.projectCollection().FindOne(ctx, filter, option)
	if err = sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			err = store.ErrProjectNotExist
		}
		return
	}

	var tmp struct {
		Id      primitive.ObjectID `bson:"_id"`
		Hash    string             `bson:"hash"`
		Branch  bool               `bson:"branch"`
		Indexed bool               `bson:"indexed"`
	}
	if err = sr.Decode(&tmp); err != nil {
		return
	}

	info.Id = tmp.Id.Hex()
	info.Hash = tmp.Hash
	info.Branch = tmp.Branch
	info.Indexed = tmp.Indexed
	return
}

func (ms *mongodbStore) SetProjectIndexed(ctx context.Context, uid, url, hash string) error {
	filter := bson.M{
		"uid":  uid,
		"url":  url,
		"hash": hash,
	}
	update := bson.M{
		"$set": bson.M{
			"indexed": true,
		},
	}
	_, err := ms.projectCollection().UpdateOne(ctx, filter, update)
	return err
}

func (ms *mongodbStore) ProjectExists(ctx context.Context, pid string) bool {
	id, err := primitive.ObjectIDFromHex(pid)
	if err != nil {
		return false
	}

	count, _ := ms.projectCollection().CountDocuments(ctx, bson.M{"_id": id})
	return count > 0
}

func (ms *mongodbStore) GetUserProjects(ctx context.Context, uid string) (projects []store.ProjectInfo, err error) {
	filter := bson.M{
		"uid": uid,
	}
	cursor, err := ms.projectCollection().Find(ctx, filter)
	if err != nil {
		return
	}
	var tmp store.ProjectInfo
	for cursor.Next(ctx) {
		if err = cursor.Decode(&tmp); err != nil {
			return
		}
		projects = append(projects, tmp)
	}
	return
}

func (ms *mongodbStore) AddAnnotation(ctx context.Context, pid, uid, file, annotation string, lineNumber int) error {
	_, err := ms.annotationCollection().InsertOne(ctx, bson.M{
		"pid":        pid,
		"uid":        uid,
		"file":       file,
		"annotation": annotation,
		"lineNumber": lineNumber,
		"createdAt":  time.Now().Unix(),
	})
	return err
}

func (ms *mongodbStore) GetAnnotations(ctx context.Context, pid, file string, lineNumber int) (records []store.AnnotationRecord, err error) {
	filter := bson.M{
		"pid":        pid,
		"file":       file,
		"lineNumber": lineNumber,
	}
	option := &options.FindOptions{
		Projection: bson.M{
			"uid":        1,
			"annotation": 1,
			"createdAt":  1,
		},
	}
	cursor, err := ms.annotationCollection().Find(ctx, filter, option)
	if err != nil {
		return
	}

	records = make([]store.AnnotationRecord, 0, 4)
	var tmp store.AnnotationRecord
	for cursor.Next(ctx) {
		if err = cursor.Decode(&tmp); err != nil {
			return
		}
		records = append(records, tmp)
	}
	return
}

func (ms *mongodbStore) GetAnnotationLines(ctx context.Context, pid, file string) (lines []int32, err error) {
	filter := bson.M{
		"pid":  pid,
		"file": file,
	}
	option := &options.FindOptions{
		Projection: bson.M{
			"lineNumber": 1,
		},
	}
	cursor, err := ms.annotationCollection().Find(ctx, filter, option)
	if err != nil {
		return
	}
	var tmp struct {
		LineNumber int32 `bson:"lineNumber"`
	}
	// TODO: use set to filter duplicates
	lines = make([]int32, 0, 32)
	for cursor.Next(ctx) {
		if err = cursor.Decode(&tmp); err != nil {
			return
		}
		lines = append(lines, tmp.LineNumber)
	}
	return
}

func (ms *mongodbStore) UpdateLatestAnnotation(ctx context.Context, pid, parent, sub, file, brief string, lineNumber int) error {
	filter := bson.M{
		"pid":    pid,
		"parent": parent,
		"sub":    sub,
	}
	update := bson.M{
		"$set": bson.M{
			"file":       file,
			"brief":      brief,
			"lineNumber": lineNumber,
			"timestamp":  time.Now().Unix(),
		},
	}
	upsert := true

	option := &options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := ms.latestAnnotationCollection().UpdateOne(ctx, filter, update, option)
	return err
}

func (ms *mongodbStore) GetLatestAnnotations(ctx context.Context, pid, parent string) (annotations []store.LatestAnnotation, err error) {
	filter := bson.M{
		"pid":    pid,
		"parent": parent,
	}
	cursor, err := ms.latestAnnotationCollection().Find(ctx, filter)
	if err != nil {
		return
	}
	var annotation store.LatestAnnotation
	for cursor.Next(ctx) {
		if err = cursor.Decode(&annotation); err != nil {
			return
		}
		annotations = append(annotations, annotation)
	}
	return
}
