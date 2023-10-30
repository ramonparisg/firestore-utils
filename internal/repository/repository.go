package repository

type Filter struct {
	Field     string      `json:"field"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value"`
}

type Repository interface {
	Query(collection string, filters []Filter, limit int) ([]interface{}, error)
}

var DEFAULT_LIMIT = 100
