package jsonx

import (
	"errors"
	"strconv"
	"unicode"
)

// ErrDefault - Default error for Decode()
var ErrDefault = errors.New("JSON could not be parsed")

// ErrParse - Only used internally
var ErrParse = errors.New("Error occured during parsing")

// lexer - internal type for keeping track
type lexer struct {
	source []byte // The whole input
	pos    int    // Current position in source
	len    int    // lenght of source
}

// Object - Represents a Json object
type Object = map[string]interface{}

// Array - Represents a Json array
type Array = []interface{}

// Literal Tokens
const (
	// Star - Used internally to support jsonc: Json Comments
	Star = byte('*')
	// Slash - Used internally to support jsonc: Json Comments
	FSlash = byte('/')

	// Dot - Used internally for reading floats
	Dot = byte('.')
	// Quotation - Used internally for reading strings
	Quotation = byte('"')

	// Comma - Syntax literal comma
	Comma = byte(',')
	// Colon - Syntax literal colon
	Colon = byte(':')

	// LeftBrace - Syntax literal to start an object
	LeftBrace = byte('{')
	// RightBrace - Syntax literal to end an object
	RightBrace = byte('}')

	// LeftBracket - Syntax literal to start a list
	LeftBracket = byte('[')
	// RightBracket - Syntax literal to end a list
	RightBracket = byte(']')
)

const eof byte = 0x04

// Decode - Decodes the input
func Decode(input []byte) (interface{}, error) {
	lexer := lexer{source: input, pos: -1, len: len(input)}
	return lexer.build()
}

// FormatNumber - Formats a number from []byte -> string -> your type
var FormatNumber = func(ByteNums []byte) interface{} {
	val, _ := strconv.ParseFloat(string(ByteNums), 64)
	return val
}

// BuildOne - Returns Objects, Arrays, Strings and Numbers
func (state *lexer) build() (interface{}, error) {
	Byte := state.swallow()

	// Parse arrays
	if Byte == LeftBracket {
		var holder Array
		for {
			if element, err := state.build(); err == nil {
				holder = append(holder, element)

				// Swallow next ... should be a comma or a RightBracket
				next := state.swallow()
				if next == Comma {
					// There is more to come...
					continue
				} else if next == RightBracket {
					// The end has been reached...
					break
				}
			} else if element == RightBracket && len(holder) == 0 {
				// Array had nothing...
				return holder, nil
			}
			// Error ... next can be colon or something that does not make sense here
			return nil, ErrDefault
		}
		return holder, nil
	}

	// Parse objects
	if Byte == LeftBrace {
		var holder = make(Object)
		for {
			if supposedKey, err := state.build(); err == nil {
				if key, isString := supposedKey.(string); isString {
					// Swallow a colon
					if state.swallow() == Colon {
						// Parse value
						if value, err := state.build(); err == nil {
							holder[key] = value

							// Swallow next ... should be a comma or a RightBrace
							next := state.swallow()
							if next == Comma {
								// There is more to come...
								continue
							} else if next == RightBrace {
								// The end has been reached...
								break
							}
						}
					}
				} else if supposedKey == RightBrace && len(holder) == 0 {
					// Object had nothing...
					return holder, nil
				}
			}
			// Error ... next can be colon or something that does not make sense here
			return nil, ErrDefault
		}
		return holder, nil
	}

	// Parse strings
	if Byte == Quotation { // Read until next quotation
		var str []byte
		for {
			// Manually parsing the input
			state.pos++
			if Byte, err := state.get(state.pos); err == nil {
				if Byte == Quotation {
					break
				}
				str = append(str, Byte)
				continue
			}
			// nil means invalid here
			return nil, ErrDefault
		}
		return string(str), nil
	}

	// If Byte is Digit
	if unicode.IsDigit(rune(Byte)) {
		digits := []byte{Byte}
		for {
			// Manually parsing the input
			state.pos++
			if Byte, err := state.get(state.pos); err == nil {
				if unicode.IsDigit(rune(Byte)) || Byte == Dot {
					// Add it to the currently read number
					digits = append(digits, Byte)

					continue // Continue to next number
				}
			}
			// Current byte was not a number, back off so next round can process it
			state.pos--
			// Numbers have finished so break
			break
		}
		return FormatNumber(digits), nil
	}

	// Lex booleans & null
	if Byte == 't' {
		Byte = state.source[state.pos+1]
		if Byte == 'r' {
			Byte = state.source[state.pos+2]
			if Byte == 'u' {
				Byte = state.source[state.pos+3]
				if Byte == 'e' {
					state.pos += 3
					return true, nil
				}
			}
		}
	} else if Byte == 'f' {
		Byte = state.source[state.pos+1]
		if Byte == 'a' {
			Byte = state.source[state.pos+2]
			if Byte == 'l' {
				Byte = state.source[state.pos+3]
				if Byte == 's' {
					Byte = state.source[state.pos+4]
					if Byte == 'e' {
						state.pos += 4
						return false, nil
					}
				}
			}
		}
	} else if Byte == 'n' {
		Byte = state.source[state.pos+1]
		if Byte == 'u' {
			Byte = state.source[state.pos+2]
			if Byte == 'l' {
				Byte = state.source[state.pos+3]
				if Byte == 'l' {
					state.pos += 3
					return nil, nil
				}
			}
		}
	}

	// Handle unknown byte... Return basically because we know that json has an error
	return nil, ErrDefault
}

// get - gets the value of a given index in source
func (state *lexer) get(n int) (byte, error) {
	if n < state.len {
		return state.source[n], nil
	}
	return 0, ErrParse
}

// peek - Returns the first non-space
func (state *lexer) peek() byte {
	for i := state.pos; i < state.len; i++ {
		if !unicode.IsSpace(rune(state.source[i])) {
			return state.source[i]
		}
	}
	return eof
}

// swallow - Returns the first non-space and advances the position
func (state *lexer) swallow() byte {
	for state.pos < state.len {
		state.pos++
		if !unicode.IsSpace(rune(state.source[state.pos])) {
			return state.source[state.pos]
		}
	}
	return eof
}
