package controller

import (
	"reflect"
	"testing"
)

// todo fix and do tests for: getting primitive string, integer, boolean, array, map, nested map, nested array, nested map with array, nested array with map, nested array with array
func Test_filterSelectedFields(t *testing.T) {
	type args struct {
		selected map[string]string
		rows     []interface{}
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "Test all fields",
			args: args{
				selected: map[string]string{
					"orderNumber":    "data.order.orderNumber",
					"isBool":         "data.order.customer.isBool",
					"product":        "data.order.orderLines[n].product.productName",
					"customInfoSize": "data.order.orderLines[n].customInfo[name=size].value[0]",
				},
				rows: getRowData(),
			},
			want: []interface{}{
				map[string]interface{}{
					"orderNumber":    123,
					"isBool":         true,
					"product":        []interface{}{"Product_1", "Product_2"},
					"customInfoSize": []interface{}{"big"},
				},
				map[string]interface{}{
					"orderNumber":    321,
					"isBool":         false,
					"product":        []interface{}{"Product_X", "Product_Y"},
					"customInfoSize": []interface{}{"grand"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := filterSelectedFields(tt.args.selected, tt.args.rows)
			for w := range tt.want {
				for k, v := range tt.want[w].(map[string]interface{}) {
					if !reflect.DeepEqual(fields[w].(map[string]interface{})[k], v.(interface{})) {
						t.Errorf("filterSelectedFields() = %v, want %v", fields[w].(map[string]interface{})[k], v)
					}
				}
			}
		})
	}
}

func getRowData() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"data": map[string]interface{}{
				"order": map[string]interface{}{
					"orderNumber": 123,
					"customer": map[string]interface{}{
						"isBool": true,
					},
					"orderLines": []interface{}{
						map[string]interface{}{
							"product": map[string]interface{}{
								"productName": "Product_1",
							},
							"customInfo": []interface{}{
								map[string]interface{}{
									"name":  "size",
									"value": []interface{}{"big", "small"},
								},
								map[string]interface{}{
									"name":  "example",
									"value": []interface{}{"value"},
								}},
						},
						map[string]interface{}{
							"product": map[string]interface{}{
								"productName": "Product_2",
							},
						},
					},
				},
			},
		},
		map[string]interface{}{
			"data": map[string]interface{}{
				"order": map[string]interface{}{
					"orderNumber": 321,
					"customer": map[string]interface{}{
						"isBool": false,
					},
					"orderLines": []interface{}{
						map[string]interface{}{
							"product": map[string]interface{}{
								"productName": "Product_X",
							},
							"customInfo": []interface{}{
								map[string]interface{}{
									"name":  "size",
									"value": []interface{}{"grand", "petit"},
								},
								map[string]interface{}{
									"name":  "example",
									"value": []interface{}{"value"},
								}},
						},
						map[string]interface{}{
							"product": map[string]interface{}{
								"productName": "Product_Y",
							},
						},
					},
				},
			},
		},
	}
}
