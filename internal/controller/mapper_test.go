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
			name: "Test 1",
			args: args{
				selected: map[string]string{
					"orderNumber": "data.order.orderNumber",
					"customer":    "data.order.customer.firstName",
					"product":     "data.order.orderLines[n].product.productName",
					//"customInfo":  "data.order.orderLines[n].customInfo[name=size].value[0]",
				},
				rows: getRowData(),
			},
			want: []interface{}{
				map[string]interface{}{
					"orderNumber": "123",
					"customer":    "John",
					"product":     []string{"Product_1", "Product_2"},
					//"customInfo":  []string{"big"},
				},
				map[string]interface{}{
					"orderNumber": "321",
					"customer":    "Anthony",
					"product":     []string{"Product_X", "Product_Y"},
					//"customInfo":  []string{"big"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := filterSelectedFields(tt.args.selected, tt.args.rows)
			for w := range tt.want {
				for k, v := range tt.want[w].(map[string]interface{}) {
					if !reflect.DeepEqual(fields[w].(map[string]interface{})[k], v.(interface{})) { // todo not working :(
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
					"orderNumber": "123",
					"customer": map[string]interface{}{
						"firstName": "John",
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
					"orderNumber": "321",
					"customer": map[string]interface{}{
						"firstName": "Anthony",
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
