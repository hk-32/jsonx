package jsonx

import (
	"reflect"
	"strconv"
)

// Encode - Encodes the input
func Encode(input interface{}) ([]byte, error) {
	out, err := build(input)
	// Just to make sure that out is nothing except for nil on failure
	if err != nil {
		return nil, err
	}
	return out, nil
}

// numberToDigits - Checks if num is a number and formats number type to []byte
func numberToDigits(value interface{}, refv reflect.Value) ([]byte, bool) {
	switch refv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(refv.Int(), 10)), true
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.FormatFloat(refv.Float(), 'f', -1, 64)), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(refv.Uint(), 10)), true
	}
	return nil, false
}

// build - Returns Objects, Arrays, Strings and Numbers in json grammar
func build(node interface{}) ([]byte, error) {
	// Handle strings & numbers
	if str, ok := node.(string); ok {
		return append(append([]byte{'"'}, []byte(str)...), '"'), nil
	} else if boolean, ok := node.(bool); ok {
		if boolean {
			return []byte{'t', 'r', 'u', 'e'}, nil
		}
		return []byte{'f', 'a', 'l', 's', 'e'}, nil
	} else if node == nil {
		return []byte{'n', 'u', 'l', 'l'}, nil
	}
	container := reflect.ValueOf(node)
	kind := container.Kind()

	// Handle Arrays
	if kind == reflect.Slice {
		len := container.Len()

		var builder = []byte{'['}
		for i := 0; i < len; i++ {
			if element, err := build(container.Index(i).Interface()); err == nil {
				builder = append(builder, element...)
				// Add comma if i was not last
				if i+1 < len {
					builder = append(builder, ',')
				}
				continue
			}
			return nil, ErrDefault
		}
		builder = append(builder, ']')
		return builder, nil
	}

	// Handle Objects
	if kind == reflect.Map {
		len := container.Len()

		var builder = []byte{'{'}
		for i, keyV := range container.MapKeys() {
			if key, ok := keyV.Interface().(string); ok {
				builder = append(append(append(builder, '"'), []byte(key)...), '"', ':')

				if element, err := build(container.MapIndex(keyV).Interface()); err == nil {
					builder = append(builder, element...)
					// Add comma if i was not last
					if i+1 < len {
						builder = append(builder, ',')
					}
					continue
				}
			}
			return nil, ErrDefault
		}
		builder = append(builder, '}')
		return builder, nil
	}

	// Handle Numbers - 1891 KB
	if num, ok := numberToDigits(node, container); ok {
		return num, nil
	}

	return nil, ErrDefault
}

/*if array, ok := thing.(Array); ok {
	var builder = []byte{'['}
	for i, v := range array {
		if element, err := build(v); err == nil {
			builder = append(builder, element...)
			// Add comma if i was not last
			if i+1 < len(array) {
				builder = append(builder, ',')
			}
			continue
		}
		return nil, ErrDefault
	}
	builder = append(builder, ']')
	return builder, nil
}*/

// Handle Objects
/*if object, ok := thing.(Object); ok {
	var builder = []byte{'{'}
	var i int
	for key, val := range object {
		builder = append(append(append(builder, '"'), []byte(key)...), '"', ':')
		if element, err := build(val); err == nil {
			builder = append(builder, element...)
			// Add comma if i was not last
			if i+1 < len(object) {
				builder = append(builder, ',')
			}
			i++
			continue
		}
		return nil, ErrDefault
	}
	builder = append(builder, '}')
	return builder, nil
}*/
