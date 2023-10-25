package repository

type Filter struct {
	Field     string      `json:"field"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value"`
}

type Repository interface {
	Query(collection string, filters []Filter) ([]interface{}, error)
}
