package gocat

import (
	"testing"
)

func TestTokenizerSimple(t *testing.T) {
	checkTypes("func", []TokenType{TT_FUNC}, t)
	checkTypes("  func", []TokenType{TT_FUNC}, t)
	checkTypes(" func ", []TokenType{TT_FUNC}, t)
	checkTypes(" func\n func\n ", []TokenType{TT_FUNC, TT_FUNC}, t)
}

func TestTokenizerSpecials(t *testing.T) {
	checkTypes("{}", []TokenType{TT_LCBRACKET, TT_RCBRACKET}, t)
	checkTypes("func()", []TokenType{TT_FUNC, TT_LPAREN, TT_RPAREN}, t)
	checkTypes(" ; ", []TokenType{TT_SEMICOLON}, t)
}

func TestTokenizerLits(t *testing.T) {
	checkTypes("5", []TokenType{TT_LITINT}, t)
	checkTypes("5.0", []TokenType{TT_LITFLOAT}, t)
	checkTypes("1.", []TokenType{TT_LITFLOAT}, t)
	checkTypes("5func", []TokenType{TT_LITINT, TT_FUNC}, t)
	checkTypes("5.func", []TokenType{TT_LITFLOAT, TT_FUNC}, t)
	checkTypes("5.1func", []TokenType{TT_LITFLOAT, TT_FUNC}, t)
	checkTypes("-5.1", []TokenType{TT_LITFLOAT}, t)
	checkTypes("-399", []TokenType{TT_LITINT}, t)
	mustError("-", t)
	mustError("5.1.", t)
	mustError("5.1.2", t)
	mustError("5..1", t)
}

func mustError(str string, t *testing.T) {
	tz := NewTokenizerString(str)

	for {
		tk, err := tz.Next()

		if err != nil {
			return
		}

		if tk.Type == TT_EOF {
			break
		}
	}

	t.Fatalf("Expected error but got none: %q", str)
}

func checkTypes(str string, tts []TokenType, t *testing.T) {
	tz := NewTokenizerString(str)

	for _, v := range tts {
		tk, err := tz.Next()

		if err != nil {
			t.Fatalf("Unexpected error: %q", err.Error())
			return
		}

		if tk.Type != v {
			t.Fatalf("Expected TT %d but got %d: %q", v, tk.Type, str)
			return
		}
	}

	tk, err := tz.Next()

	if err != nil {
		t.Fatalf("Unexpected error: %q", err.Error())
		return
	}

	if tk.Type != TT_EOF {
		t.Fatalf("Expected EOF but still got a token")
		return
	}
}
