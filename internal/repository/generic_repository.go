package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
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

func (r GenericRepository) Query(collection string, filters []Filter, limit int) ([]interface{}, error) {
	coll := r.firestore.Collection(collection)
	var query firestore.Query
	for i, filter := range filters {
		if i == 0 {
			query = coll.Where(filter.Field, filter.Operation, filter.Value)
		} else {
			query = query.Where(filter.Field, filter.Operation, filter.Value)
		}
	}

	l := DEFAULT_LIMIT
	if limit > 0 {
		l = limit
	}

	results, err := query.
		Limit(l).
		Documents(ctx).
		GetAll()

	if err != nil {
		fmt.Println("Error getting documents", err)
		return nil, err
	}

	var data []interface{}
	for _, result := range results {
		data = append(data, result.Data())
	}

	return data, nil

}
