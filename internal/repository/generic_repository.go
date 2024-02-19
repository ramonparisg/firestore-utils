package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"strings"
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

const chunkSize = 30

func (r GenericRepository) Query(collection string, filters []Filter, limit int) ([]interface{}, error) {
	coll := r.getCollection(collection)

	hasInCondition, filterInCondition := getFilterWithInCondition(filters)
	if hasInCondition {
		results, err := r.queryWithChunks(filters, filterInCondition, coll, limit)
		if err != nil {
			return nil, err
		}
		return results, nil
	}
	return r.runQuery(filters, coll, limit)
}

func (r GenericRepository) getCollection(collection string) *firestore.CollectionRef {
	// todo refactor and make it more dynamic
	split := strings.Split(collection, "/")
	var coll *firestore.CollectionRef
	coll = r.firestore.Collection(split[0])
	if len(split) == 3 {
		coll = coll.Doc(split[1]).Collection(split[2])
	}
	return coll
}

func (r GenericRepository) queryWithChunks(filters []Filter, filterInCondition *Filter, coll *firestore.CollectionRef, limit int) ([]interface{}, error) {
	valueArray := filterInCondition.Value.([]interface{})
	chunks := getChunks(valueArray)

	var allResults []interface{}
	for _, chunk := range chunks {
		filtersChunk := duplicateWithChangedFirstValue(filters, chunk)
		results, err := r.runQuery(filtersChunk, coll, limit)
		if err != nil {
			return nil, err
		}
		if len(results) > 0 {
			allResults = append(allResults, results...)
		}
	}

	return allResults, nil
}

func getChunks(values []interface{}) [][]interface{} {
	var chunks [][]interface{}
	for i := 0; i < len(values); i += chunkSize {
		end := i + chunkSize
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[i:end])
	}
	return chunks
}

func duplicateWithChangedFirstValue(filters []Filter, newValue interface{}) []Filter {
	duplicate := make([]Filter, len(filters))
	copy(duplicate, filters)
	duplicate[0].Value = newValue
	return duplicate
}

func (r GenericRepository) runQuery(filters []Filter, coll *firestore.CollectionRef, limit int) ([]interface{}, error) {
	query := coll.Limit(DEFAULT_LIMIT)
	for i, filter := range filters {
		value := filter.Value
		if filter.Operation == "in" {
			value = filter.Value.([]interface{})
		}
		if i == 0 {
			query = coll.Where(filter.Field, filter.Operation, value)
		} else {
			query = query.Where(filter.Field, filter.Operation, value)
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

func getFilterWithInCondition(filters []Filter) (bool, *Filter) {
	for _, filter := range filters {
		if filter.Operation == "in" {
			return true, &filter
		}
	}
	return false, nil
}
