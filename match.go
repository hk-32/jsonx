package jsonx

import "reflect"

// Match - Used to make sure types of y conforms to that of schema, use it to validate incoming json
// After this use recieved structure with type assertions without fear ie no error checking
func Match(schema, y interface{}) bool {
	// Arrays, Strings, Numbers(same type, ie float64 by default).
	if reflect.TypeOf(schema).Kind() == reflect.TypeOf(y).Kind() {
		return true
	} else if x, xok := schema.(Object); xok {
		if y, yok := y.(Object); yok {
			if len(x) == len(y) {
				for k, v1 := range x {
					if v2, exists := y[k]; exists {
						if Match(v1, v2) {
							continue
						}
					}
					return false
				}
				// Loop successfully ran on all values
				return true
			}
		}
		return false
	}
	return false
}
