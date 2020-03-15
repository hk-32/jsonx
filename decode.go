package jsonx

import "errors"

// pow10tab - Stores the pre-computed values upto 10^(31) or 1 with 31 zeros
var pow10tab = [...]float64{
	1e00, 1e01, 1e02, 1e03, 1e04, 1e05, 1e06, 1e07, 1e08, 1e09,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
	1e20, 1e21, 1e22, 1e23, 1e24, 1e25, 1e26, 1e27, 1e28, 1e29,
	1e30, 1e31,
}

// ErrDefault - Default error for Decode()
var ErrDefault = errors.New("JSON could not be parsed")

// lexer - Internal structure for keeping track of state
type lexer struct {
	source []byte // The whole input
	pos    int    // Current position in source
	len    int    // Lenght of source
}

// Object - Represents a Json Object
type Object = map[string]interface{}

// Array - Represents a Json Array
type Array = []interface{}

// Identifiers
const (
	// EOS - Used internally to signify end of stream
	iEOS byte = 0x03 // 0x03 = End of Text, 0x04 = End of Transmission
	// Star - Used internally to support jsonc: Json Comments
	iStar = byte('*')
	// Slash - Used internally to support jsonc: Json Comments
	iFSlash = byte('/')

	// Dot - Used internally for reading floats
	iDot = byte('.')
	// Quotation - Used internally for reading strings
	iQuotation = byte('"')

	// Hyphen - Syntax literal negative
	iHyphen = byte('-')

	// Comma - Syntax literal comma
	iComma = byte(',')
	// Colon - Syntax literal colon
	iColon = byte(':')

	// LeftBrace - Syntax literal to start an object
	iLeftBrace = byte('{')
	// RightBrace - Syntax literal to end an object
	iRightBrace = byte('}')

	// LeftBracket - Syntax literal to start a list
	iLeftBracket = byte('[')
	// RightBracket - Syntax literal to end a list
	iRightBracket = byte(']')
)

// Decode - Decodes the input
func Decode(input []byte) (interface{}, error) {
	lexer := lexer{source: input, pos: -1, len: len(input)}
	out, err := lexer.build()
	// Just to make sure that out is nothing except for nil on failure
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FormatNumber - Formats digits from []byte -> your type
var FormatNumber = func(ByteNums []byte, isNegative bool) interface{} {
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

// build - Returns Objects, Arrays, Strings and Numbers
func (state *lexer) build() (interface{}, error) {
	Byte := state.swallow()

	// Parse Arrays
	if Byte == iLeftBracket {
		var holder Array
		for {
			if element, err := state.build(); err == nil {
				holder = append(holder, element)

				// Swallow next ... should be a comma or a RightBracket
				next := state.swallow()
				if next == iComma {
					// There is more to come...
					continue
				} else if next == iRightBracket {
					// The end has been reached...
					break
				}
			} else if element == iRightBracket && len(holder) == 0 {
				// Array had nothing...
				return holder, nil
			}
			// Error ... next can be EOS or something that does not make sense here
			return nil, ErrDefault
		}
		return holder, nil
	}

	// Parse Objects
	if Byte == iLeftBrace {
		var holder = make(Object)
		for {
			if supposedKey, err := state.build(); err == nil {
				if key, isString := supposedKey.(string); isString {
					// Swallow a colon
					if state.swallow() == iColon {
						// Build value
						if value, err := state.build(); err == nil {
							holder[key] = value

							// Swallow next ... should be a comma or a RightBrace
							next := state.swallow()
							if next == iComma {
								// There is more to come...
								continue
							} else if next == iRightBrace {
								// The end has been reached...
								break
							}
						}
					}
				} else if supposedKey == iRightBrace && len(holder) == 0 {
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
		var str []byte
		// Basically read till next quotation
		for {
			state.pos++
			if Byte := state.get(state.pos); Byte != iEOS {
				if Byte == iQuotation {
					break
				}
				str = append(str, Byte)
				continue
			}
			// json can't end with an unfinished string so error out
			return nil, ErrDefault
		}
		return string(str), nil
	}

	// Parse Numbers
	if isDigit(Byte) || Byte == iHyphen {
		isNegative := Byte == iHyphen
		isDecimalIncluded := false

		var digits []byte
		if !isNegative {
			digits = append(digits, Byte)
		}
		for {
			state.pos++
			if Byte := state.get(state.pos); Byte != iEOS && Byte != iHyphen {
				if isDigit(Byte) {
					// Add it to the currently read number
					digits = append(digits, Byte)
					continue
				} else if Byte == iDot {
					if !isDecimalIncluded {
						digits = append(digits, Byte)
						isDecimalIncluded = true
						continue
					}
					return nil, ErrDefault
				}
				// Current byte was not a digit, back off so next round can process it
				state.pos--
				break
			}
			return nil, ErrDefault
		}
		return FormatNumber(digits, isNegative), nil
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

// get - Returns the value of a given index in source
func (state *lexer) get(n int) byte {
	if n < state.len {
		return state.source[n]
	}
	return iEOS
}

// peek - Returns the first non-space without advancing the position
func (state *lexer) peek() byte {
	for i := state.pos; i < state.len; i++ {
		if isSpace(state.source[i]) {
			return state.source[i]
		}
	}
	return iEOS
}

// swallow - Returns the first non-space and advances the position
func (state *lexer) swallow() byte {
	for state.pos+1 < state.len {
		state.pos++
		if !isSpace(state.source[state.pos]) {
			return state.source[state.pos]
		}
	}
	return iEOS
}

// isDigit - Checks if the byte is a digit ie, 0 - 9
func isDigit(r byte) bool {
	// return '0' <= r && r <= '9'
	return r >= '0' && r <= '9'
}

// isSpace - Checks if the byte is empty space
func isSpace(r byte) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

// pow10 - Raises 10 by n
func pow10(n int) float64 {
	if n >= 0 && n <= 31 { // n >= 0 && n <= 308
		// return pow10postab32[uint(n)/32] * pow10tab[uint(n)%32]
		return pow10tab[n]
	}
	panic("out of range number encountered")
}
