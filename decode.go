package jsonx

import "unsafe"

// ByteDigitsToNumber - Formats digits from []byte -> your type (default float64)
var ByteDigitsToNumber = ByteDigitsToFloat64

// Decode - Decodes the input
func Decode(input []byte) (interface{}, error) {
	dec := decoder{source: input, pos: -1, len: len(input)}
	out, err := dec.compose()

	// Just to make sure that out is nothing except for nil on failure
	if err != nil {
		return nil, err
	}
	return out, nil
}

// build - Returns Objects, Arrays, Strings and Numbers with err nil, rest as byte but err
func (state *decoder) compose() (interface{}, error) {
	Byte := state.swallow()

	// Parse Arrays
	if Byte == iLeftBracket {
		var structure Array

		var element interface{}
		var err error
		for {
			if element, err = state.compose(); err == nil {
				structure = append(structure, element)

				// Swallow next... should be an iComma or an iRightBracket
				switch state.swallow() {
				case iComma:
					// There is more to come...
					continue
				case iRightBracket:
					// The end has been reached.
					return structure, nil
				}
			} else if element == iRightBracket && len(structure) == 0 {
				// Array had nothing...
				return structure, nil
			}
			// Error ... element can be iEOS or something that does not make sense here
			return nil, ErrDefault
		}
	}

	// Parse Objects
	if Byte == iLeftBrace {
		var holder = make(Object)

		var value interface{}
		var err error
		var key string
		var isString bool
		for {
			if value, err = state.compose(); err == nil {
				if key, isString = value.(string); isString {
					// Swallow a colon
					if state.swallow() == iColon {
						// Build value
						if value, err = state.compose(); err == nil {
							holder[key] = value

							// Swallow next... should be an iComma or an iRightBrace
							next := state.swallow()
							if next == iComma {
								// There is more to come...
								continue
							} else if next == iRightBrace {
								// The end has been reached.
								break
							}
						}
					}
				} else if value == iRightBrace && len(holder) == 0 {
					// Object had nothing...
					return holder, nil
				}
			}
			// Error ... next can be EOS or something that does not make sense here
			return nil, ErrDefault
		}
		return holder, nil
	}

	// Parse Strings
	if Byte == iQuotation {
		state.pos++
		var l int
		// Get string lenght
		for Byte = state.get(state.pos); Byte != iQuotation; Byte = state.get(state.pos + l) {
			if Byte == iEOS {
				// Json strings can't end unfinished, so error out
				return nil, ErrDefault
			}
			l++
		}

		var str = make([]byte, l)
		l = 0
		// Read everything till next iQuotation
		for Byte = state.get(state.pos); Byte != iQuotation; Byte = state.get(state.pos + l) {
			str[l] = Byte
			l++
		}
		// Success
		state.pos += l
		return *(*string)(unsafe.Pointer(&str)), nil
	}

	// Parse Numbers
	if isDigit(Byte) || Byte == iHyphen {
		isNegative := Byte == iHyphen
		hasDecimal := false

		var digits []byte
		if !isNegative {
			digits = append(digits, Byte)
		}

		for Byte = state.get(state.pos + 1); Byte != iEOS; Byte = state.get(state.pos + 1) {
			if isDigit(Byte) {
				// Add it to the currently read number
				digits = append(digits, Byte)
				state.pos++ // Confirms that i was indeed a digit
				continue
			} else if Byte == iDot {
				if !hasDecimal {
					digits = append(digits, Byte)
					hasDecimal = true
					state.pos++
					continue
				}
				return nil, ErrDefault
			}
			// Some other Byte found most probably
			break
		}
		return ByteDigitsToNumber(digits, isNegative), nil
	}

	// Booleans & null
	if Byte == 't' {
		if state.get(state.pos+1) == 'r' {
			if state.get(state.pos+2) == 'u' {
				if state.get(state.pos+3) == 'e' {
					state.pos += 3
					return true, nil
				}
			}
		}
	} else if Byte == 'f' {
		if state.get(state.pos+1) == 'a' {
			if state.get(state.pos+2) == 'l' {
				if state.get(state.pos+3) == 's' {
					if state.get(state.pos+4) == 'e' {
						state.pos += 4
						return false, nil
					}
				}
			}
		}
	} else if Byte == 'n' {
		if state.get(state.pos+1) == 'u' {
			if state.get(state.pos+2) == 'l' {
				if state.get(state.pos+3) == 'l' {
					state.pos += 3
					return nil, nil
				}
			}
		}
	}

	// Handle unknown byte... Return basically because we know that json has an error
	return Byte, ErrDefault
}

// ByteDigitsToFloat64 - Default number decoding formatter
func ByteDigitsToFloat64(ByteNums []byte, isNegative bool) interface{} {
	// Formats the digits according to base 10
	var exponent int
	var number float64

	// Reverse loop to start from the right & add digit adjusted for power to total
	for i := len(ByteNums) - 1; i >= 0; i-- {
		digit := ByteNums[i]
		if digit != iDot {
			number += (float64(digit - '0')) * pow10(exponent)
			exponent++
		} else {
			// Divide number by its upper length & reset exponent
			number /= pow10tab[exponent]
			exponent = 0
		}
	}
	if isNegative {
		return -number
	}
	return number
}
