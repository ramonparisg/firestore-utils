package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"sync"
)

type GenericRepository struct {
	firestore *firestore.Client
}

var repositorySingleton *GenericRepository
var once sync.Once
var ctx = context.Background()

func NewRepository(client *firestore.Client) *GenericRepository {
	once.Do(func() {
		repositorySingleton = &GenericRepository{
			firestore: client,
		}
	})
	return repositorySingleton
}

func (r GenericRepository) Query(collection string, filters []Filter) ([]interface{}, error) {
	coll := r.firestore.Collection(collection)
	var query firestore.Query
	for i, filter := range filters {
		if i == 0 {
			query = coll.Where(filter.Field, filter.Operation, filter.Value)
		} else {
			query = query.Where(filter.Field, filter.Operation, filter.Value)
		}
	}

	results, err := query.
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	var data []interface{}
	for _, result := range results {
		data = append(data, result.Data())
	}

	return data, nil

}
