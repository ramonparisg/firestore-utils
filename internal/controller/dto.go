package controller

type QueryRequest struct {
	Collection string            `json:"collection"`
	Select     map[string]string `json:"select"`
	Filters    []struct {
		Field     string      `json:"field"`
		Operation string      `json:"operation"`
		Value     interface{} `json:"value"`
	} `json:"filters"`
	Limit int `json:"limit"`
}
