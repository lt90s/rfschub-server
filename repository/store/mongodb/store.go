package mongodb

import (
	"context"
	"github.com/lt90s/rfschub-server/common/store/mongodb"
	"github.com/lt90s/rfschub-server/gits/proto"
	"github.com/lt90s/rfschub-server/repository/config"
	"github.com/lt90s/rfschub-server/repository/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"path"
	"sync"
)

const (
	repositoryCollection = "repositories"
	fileCollection       = "files"
)

type mongodbStore struct {
	client         *mongo.Client
	name           string
	mutex          sync.Mutex
	setDirectories map[string]struct{}
}

func NewMongodbStore() store.Store {
	conf := config.DefaultConfig
	client := mongodb.NewClient(conf.Mongodb.Uri)
	return &mongodbStore{
		name:           conf.Mongodb.Database,
		client:         client,
		setDirectories: make(map[string]struct{}),
	}
}

func (m *mongodbStore) repositoryCollection() *mongo.Collection {
	return m.client.Database(m.name).Collection(repositoryCollection)
}

func (m *mongodbStore) fileCollection() *mongo.Collection {
	return m.client.Database(m.name).Collection(fileCollection)
}

//repository collection structure:
//{
//	_id: primitive.ObjectId
//	url: "url",
//	commits: [{
//		name: "master",
//		hash: "xxxxx",
//		branch: true,
//	}, ...]
//}

func (m *mongodbStore) AddRepository(ctx context.Context, url string, commits []store.NamedCommit) error {
	filter := bson.M{
		"url": url,
	}
	update := bson.M{
		"$set": bson.M{
			"commits": commits,
		},
	}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := m.repositoryCollection().UpdateOne(ctx, filter, update, option)
	return err
}

func (m *mongodbStore) GetRepository(ctx context.Context, url string) (commits []store.NamedCommit, err error) {
	filter := bson.M{
		"url": url,
	}
	result := m.repositoryCollection().FindOne(ctx, filter)
	if err = result.Err(); err != nil {
		return
	}
	var tmp struct {
		Commits []store.NamedCommit `bson:"commits"`
	}
	err = result.Decode(&tmp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = store.ErrorRepositoryNotFound
		}
		return
	}
	commits = tmp.Commits
	return
}

func (m *mongodbStore) GetCommitByName(ctx context.Context, url string, name string) (string, error) {
	filter := bson.M{
		"url": url,
		"commits": bson.M{
			"$elemMatch": bson.M{
				"name": name,
			},
		},
	}
	option := &options.FindOneOptions{
		Projection: bson.M{
			"commits": 1,
		},
	}
	sr := m.repositoryCollection().FindOne(ctx, filter, option)
	if sr.Err() != nil {
		return "", sr.Err()
	}

	var tmp struct {
		Commits []store.NamedCommit `bson:"commits"`
	}
	err := sr.Decode(&tmp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = store.ErrorRepositoryNotFound
		}
		return "", err
	}

	return tmp.Commits[0].Hash, nil
}

func (m *mongodbStore) RepositoryExist(ctx context.Context, url string, hash string) (bool, error) {
	filter := bson.M{
		"url": url,
		"commits": bson.M{
			"$elemMatch": bson.M{
				"hash": hash,
			},
		},
	}
	count, err := m.repositoryCollection().CountDocuments(ctx, filter)
	return count > 0, err
}

type dbEntry struct {
	UrlCommit string `bson:"urlCommit"`
	File      string `bson:"file"`
	ParentDir string `bson:"parentDir"`
	Dir       bool   `bson:"dir"`
	Content   string `bson:"content"`
	Synced    bool   `bson:"synced"`
	Plain     bool   `bson:"plain"`
}

// mutex is used to avoid simultaneously setting directory entries
// for each repository's commit, this should be called only once
func (m *mongodbStore) SetDirectories(ctx context.Context, url, commit string, entries []*gits.FileEntry) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	key := url + "@" + commit
	if _, ok := m.setDirectories[key]; ok {
		return nil
	}

	dbEntries := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		dbEntries = append(dbEntries, dbEntry{
			UrlCommit: key,
			File:      entry.File,
			ParentDir: path.Dir(entry.File),
			Dir:       entry.Dir,
		})
	}

	dbEntries = append(dbEntries, dbEntry{
		UrlCommit: key,
		File:      "/",
		Dir:       true,
	})

	_, err := m.fileCollection().InsertMany(ctx, dbEntries)
	return err
}

func (m *mongodbStore) GetDirectoryEntries(ctx context.Context, url, commit, path string) (synced bool, entries []store.DirectoryEntry, err error) {
	filter := bson.M{
		"urlCommit": url + "@" + commit,
		"parentDir": path,
	}
	option := &options.FindOptions{
		Projection: bson.M{
			"file": 1,
			"dir":  1,
		},
	}
	cursor, err := m.fileCollection().Find(ctx, filter, option)
	if err != nil {
		return
	}

	var entry store.DirectoryEntry
	for cursor.Next(ctx) {
		if err = cursor.Err(); err != nil {
			return
		}
		if err = cursor.Decode(&entry); err != nil {
			return
		}
		entries = append(entries, entry)
	}
	if len(entries) > 0 {
		synced = true
		return
	}

	// check if synced
	filter = bson.M{
		"urlCommit": url + "@" + commit,
		"file":      "/",
	}
	count, err := m.fileCollection().CountDocuments(ctx, filter)
	synced = count > 0
	return
}

func (m *mongodbStore) SetBlob(ctx context.Context, url, commit, path, content string, plain bool) error {
	filter := bson.M{
		"urlCommit": url + "@" + commit,
		"file":      path,
	}
	update := bson.M{
		"$set": bson.M{
			"synced":  true,
			"plain":   plain,
			"content": content,
		},
	}

	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := m.fileCollection().UpdateOne(ctx, filter, update, option)
	return err
}

func (m *mongodbStore) GetBlob(ctx context.Context, url, commit, path string) (blob store.Blob, err error) {
	filter := bson.M{
		"urlCommit": url + "@" + commit,
		"file":      path,
	}
	option := &options.FindOneOptions{
		Projection: bson.M{
			"synced":  1,
			"plain":   1,
			"content": 1,
		},
	}

	result := m.fileCollection().FindOne(ctx, filter, option)
	if err = result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			err = store.ErrorBlobNotFound
		}
		return
	}
	err = result.Decode(&blob)
	return
}
