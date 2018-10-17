package gocat

import (
	"testing"
)

func TestParseType(t *testing.T) {
	checkParseType("int", &PrimType{Type: "int"}, t)
	ut, _ := NewUnionType(
		[]Type{
			&PrimType{Type: "int"},
			&PrimType{Type: "float"},
		})
	checkParseType("{int float}", ut, t)
	mustErrorParseType("{int {foo bar} float}", t)
}

func TestParseExp(t *testing.T) {
	checkASTExp(
		"5 6 foo;",
		&ExpNode{
			Exps: []Node{
				&LitIntNode{
					Value: 5,
				},
				&LitIntNode{
					Value: 6,
				},
				&VerbNode{
					Verb: "foo",
				},
			},
		}, t)
}

func TestParseFunc(t *testing.T) {

	checkASTFunc(
		"func main [(a int)] [float] {}",
		&FuncNode{
			Name:     "main",
			RetTypes: []Type{&PrimType{Type: "float"}},
			Body:     []Node{},
			Args: []Arg{
				Arg{
					Type: &PrimType{Type: "int"},
					Name: "a",
				},
			},
		}, t)

	checkASTFunc(
		"func main [(a int) (b float)] [float string] {}",
		&FuncNode{
			Name:     "main",
			RetTypes: []Type{&PrimType{Type: "float"}, &PrimType{Type: "string"}},
			Body:     []Node{},
			Args: []Arg{
				Arg{
					Type: &PrimType{Type: "int"},
					Name: "a",
				},
				Arg{
					Type: &PrimType{Type: "float"},
					Name: "b",
				},
			},
		}, t)
}

func checkASTFunc(code string, exp Node, t *testing.T) {
	p := NewParser(NewTokenizerString(code))

	n, err := p.parseFunc()

	if err != nil {
		t.Fatalf("Unexpected error for %s: %s.", code, err.Error())
	}

	if !ASTEqual(n, exp) {
		t.Fatalf("ASTs do not match for %s! %+v %+v", code, n, exp)
	}
}

func checkASTExp(code string, exp Node, t *testing.T) {
	p := NewParser(NewTokenizerString(code))

	n, err := p.parseExp()

	if err != nil {
		t.Fatalf("Unexpected error for %s: %s.", code, err.Error())
	}

	if !ASTEqual(n, exp) {
		t.Fatalf("ASTs do not match for %s! %+v %+v", code, n, exp)
	}
}

func checkParseType(code string, exp Type, t *testing.T) {
	p := NewParser(NewTokenizerString(code))

	n, err := p.parseType()

	if err != nil {
		t.Fatalf("Unexpected error for %s: %s.", code, err.Error())
		return
	}

	if !TypeEqual(n, exp) {
		t.Fatalf("Got type %s but wanted %s for %s.", n, exp, code)
		return
	}
}

func mustErrorParseType(code string, t *testing.T) {
	p := NewParser(NewTokenizerString(code))

	_, err := p.parseType()

	if err == nil {
		t.Fatalf("Expected error but got none for: %s", code)
		return
	}
}
