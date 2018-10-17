package gocat

import (
	"testing"
)

func TestInferType(t *testing.T) {
	checkInferedTypeExp("5 square.i;", []Type{&PrimType{Type: "int"}}, t)
	mustErrorInferedTypeExp("5.0 square.i;",
		&PrimType{Type: "int"},
		&PrimType{Type: "float"}, t)
}

func mustErrorInferedTypeExp(code string, wanted, got Type, t *testing.T) {
	p := NewParser(NewTokenizerString(code))
	n, err := p.parseExp()

	if err != nil {
		t.Fatalf("Unexpected error for %s: %s", code, err.Error())
		return
	}

	typ, err := InferTypes(n, nil, NewTypeWorlds(builtins))

	if err == nil {
		t.Fatalf("Expected error but got none for: %s. {%s}", code, typ)
		return
	}

	switch err.(type) {
	case *TypeError:
		te := err.(*TypeError)
		if !TypeEqual(te.Wanted, wanted) {
			t.Fatalf("Wanted mismatch %s != %s for %s.", te.Wanted, wanted, code)
			return
		}

		if !TypeEqual(te.Got, got) {
			t.Fatalf("Got mismatch %s != %s for %s.", te.Got, got, code)
			return
		}
	default:
		t.Fatalf("Unexpected error for %s: %s", code, err.Error())
		return
	}
}

func checkInferedTypeExp(code string, exp []Type, t *testing.T) {
	p := NewParser(NewTokenizerString(code))
	n, err := p.parseExp()

	if err != nil {
		t.Fatalf("Unexpected error for %s: %s", code, err.Error())
		return
	}

	types, err := InferTypes(n, nil, NewTypeWorlds(builtins))

	if err != nil {
		t.Fatalf("Unexpected error for %s: %s", code, err.Error())
		return
	}

	if !TypesEqual(types, exp) {
		t.Fatalf("Expected types %s but got %s for %s.", exp, types, code)
		return
	}
}
