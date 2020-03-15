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

// build - Returns Objects, Arrays, Strings and Numbers in json grammar
func build(thing interface{}) ([]byte, error) {
	// Handle strings & numbers
	if str, ok := thing.(string); ok {
		return append(append([]byte{'"'}, []byte(str)...), '"'), nil
	} else if boolean, ok := thing.(bool); ok {
		if boolean {
			return []byte{'t', 'r', 'u', 'e'}, nil
		}
		return []byte{'f', 'a', 'l', 's', 'e'}, nil
	} else if thing == nil {
		return []byte{'n', 'u', 'l', 'l'}, nil
	}

	// Handle Arrays
	if reflect.TypeOf(thing).Kind() == reflect.Slice {
		array := reflect.ValueOf(thing)
		len := array.Len()

		var builder = []byte{'['}
		for i := 0; i < len; i++ {
			if element, err := build(array.Index(i).Interface()); err == nil {
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
	if reflect.TypeOf(thing).Kind() == reflect.Map {
		object := reflect.ValueOf(thing)
		len := object.Len()

		var builder = []byte{'{'}
		for i, keyV := range object.MapKeys() {
			if key, ok := keyV.Interface().(string); ok {
				builder = append(append(append(builder, '"'), []byte(key)...), '"', ':')

				if element, err := build(object.MapIndex(keyV).Interface()); err == nil {
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

	// Handle Numbers - 1894 KB
	switch reflect.TypeOf(thing).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(reflect.ValueOf(thing).Int(), 10)), nil
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.FormatFloat(reflect.ValueOf(thing).Float(), 'f', -1, 64)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(reflect.ValueOf(thing).Uint(), 10)), nil
	}

	// Handle Numbers - 2085 KB
	/*switch v := thing.(type) {
	case int, int16, int32, int64, float32, float64:
		return []byte(fmt.Sprintf("%v", v)), nil
	}*/

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
