// Code generated by "stringer -type Token"; DO NOT EDIT.

package xsqlparser

import "strconv"

const _Token_name = "SQLKeywordNumberCharSingleQuotedStringNationalStringLiteralCommaWhitespaceEqNeqLtGtLtEqGtEqPlusMinusMultDivModLParenRParenPeriodColonDoubleColonSemicolonBackslashLBracketRBracketAmpersandLBraceRBraceILLEGAL"

var _Token_index = [...]uint8{0, 10, 16, 20, 38, 59, 64, 74, 76, 79, 81, 83, 87, 91, 95, 100, 104, 107, 110, 116, 122, 128, 133, 144, 153, 162, 170, 178, 187, 193, 199, 206}

func (i Token) String() string {
	if i < 0 || i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}
