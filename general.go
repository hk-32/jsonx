package jsonx

import (
	"errors"
	"strconv"
	"unsafe"
)

// pow10tab - Stores the pre-computed values upto 10^(31) or 1 with 31 zeros
var pow10tab = [...]float64{
	1e00, 1e01, 1e02, 1e03, 1e04, 1e05, 1e06, 1e07, 1e08, 1e09,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
	1e20, 1e21, 1e22, 1e23, 1e24, 1e25, 1e26, 1e27, 1e28, 1e29,
	1e30, 1e31,
}

// ErrDefault - Default error for Decode()
var ErrDefault = errors.New("JSON could not be parsed")

// state - Internal structure for keeping track of state
type state struct {
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
	iStar byte = '*'
	// Slash - Used internally to support jsonc: Json Comments
	iFSlash byte = '/'

	// Dot - Used internally for reading floats
	iDot byte = '.'
	// Quotation - Used internally for reading strings
	iQuotation byte = '"'

	// Hyphen - Syntax literal negative
	iHyphen byte = '-'

	// Comma - Syntax literal comma
	iComma byte = ','
	// Colon - Syntax literal colon
	iColon byte = ':'

	// LeftBrace - Syntax literal to start an object
	iLeftBrace byte = '{'
	// RightBrace - Syntax literal to end an object
	iRightBrace byte = '}'

	// LeftBracket - Syntax literal to start a list
	iLeftBracket byte = '['
	// RightBracket - Syntax literal to end a list
	iRightBracket byte = ']'
)

// ByteDigitsToInt - Option to change ByteDigitsToNumber incase you want int values
func ByteDigitsToInt(ByteNums []byte, isNegative bool) interface{} {
	// We are 99.9% sure this will always be successful, the parser made sure the digits and format was valid
	if num, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&ByteNums)), 10, 0); err == nil {
		if isNegative {
			return -int(num)
		}
		return int(num)
	}
	return 0
}

// get - Returns the value of a given index in source
func (state *state) get(n int) byte {
	if n < state.len {
		return state.source[n]
	}
	return iEOS
}

// peek - Returns the first non-space without advancing the position
func (state *state) peek() byte {
	for i := state.pos; i < state.len; i++ {
		if isSpace(state.source[i]) {
			return state.source[i]
		}
	}
	return iEOS
}

// swallow - Returns the first non-space and advances the position
func (state *state) swallow() byte {
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
	//if n >= 0 && n <= 31 { // n >= 0 && n <= 308
	// return pow10postab32[uint(n)/32] * pow10tab[uint(n)%32]
	if n > 31 {
		n = 31
	}
	return pow10tab[n]
	//}
	//panic("out of range number encountered")
}
