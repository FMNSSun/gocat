package gocat

import (
	"fmt"
)

type TypeError struct {
	Wanted Type
	Got    Type
	Token  *Token
	Extra  string
}

type TypeWorld map[string]Type
type TypeWorlds []TypeWorld

func (tws TypeWorlds) Lookup(val string) Type {
	for i := len(tws) - 1; i >= 0; i-- {
		it := tws[i][val]

		if it != nil {
			return it
		}
	}

	return nil
}

func NewTypeWorlds(typeWorlds ...TypeWorld) TypeWorlds {
	return typeWorlds
}

func (te *TypeError) Error() string {
	if te.Extra == "" {
		return fmt.Sprintf("Type error %s: Wanted type `%s` but got type `%s`.",
			te.Token.Pos, te.Wanted, te.Got)
	} else {
		return fmt.Sprintf("Type error %s %s: Wanted type `%s` but got type `%s`.",
			te.Extra, te.Token.Pos, te.Wanted, te.Got)
	}
}

var builtins map[string]Type = map[string]Type{
	"square.i": &FuncType{
		ArgTypes: []Type{
			&PrimType{
				Type: "int",
			},
		},
		RetTypes: []Type{
			&PrimType{
				Type: "int",
			},
		},
	},
}

func TypeCompatibleWith(a Type, b Type) bool {
	switch a.(type) {
	case *VoidType:
		switch b.(type) {
		case *VoidType:
			return true
		default:
			return false
		}
	case *PrimType:
		switch b.(type) {
		case *PrimType:
			return TypeEqual(a, b)
		case *UnionType:
			ut := b.(*UnionType)

			for _, typ := range ut.Types {
				if TypeEqual(a, typ) {
					return true
				}
			}

			return false
		default:
			return false
		}
	case *UnionType:
		switch b.(type) {
		case *UnionType:
			// all types of a must be types of b as well.
			ut_a := a.(*UnionType)
			ut_b := b.(*UnionType)

			for _, typ_a := range ut_a.Types {
				found := false
				for _, typ_b := range ut_b.Types {
					if TypeEqual(typ_a, typ_b) {
						found = true
						break
					}
				}

				if !found {
					return false
				}
			}

			return true
		default:
			return false
		}
	}

	panic("BUG: Can't tell if compatible or not?")
}

func InferTypes(node Node, stack []Type, typeWorlds TypeWorlds) ([]Type, error) {
	switch node.(type) {
	// Literals are easy to infer the type of.
	case *LitFloatNode:
		return append(stack, &PrimType{Type: "float"}), nil
	case *LitIntNode:
		return append(stack, &PrimType{Type: "int"}), nil

	case *ExpNode:
		exp := node.(*ExpNode)

		for _, v := range exp.Exps {
			switch v.(type) {
			// If the expression contains a literal just push the type
			// of the literal to the stack.
			case *LitIntNode:
				stack = append(stack, &PrimType{Type: "int"})
			case *LitFloatNode:
				stack = append(stack, &PrimType{Type: "float"})

			// If it's a verb we need to look up what argument types it expects
			// and what return types it has.
			case *VerbNode:
				verb := v.(*VerbNode).Verb
				ft := typeWorlds.Lookup(verb)

				if ft == nil {
					return nil, fmt.Errorf("Function `%s` does not exist!", verb)
				}
				funcType, ok := ft.(*FuncType)

				if !ok {
					return nil, fmt.Errorf("`%s` is not of type function.", verb)
				}

				if len(stack) < len(funcType.ArgTypes) {
					return nil, fmt.Errorf("Not enough arguments.")
				}

				m := len(funcType.ArgTypes)

				// On top of the stack is the last argument type so the first argument
				// type according to funcType.ArgTypes is offset by minus the amount of
				// arguments the function expects.
				for i := 0; i < m; i++ {
					got := stack[len(stack)-m+i]
					wanted := funcType.ArgTypes[i]
					if !TypeCompatibleWith(got, wanted) {
						return nil, &TypeError{
							Wanted: wanted,
							Got:    got,
							Token:  exp.Token,
							Extra:  fmt.Sprintf("in a call to `%s`.", verb),
						}
					}
				}

				// Pop the argument types from the stack
				stack = stack[:len(stack)-m]

				// And push the return types
				for _, rettyp := range funcType.RetTypes {
					stack = append(stack, rettyp)
				}
			}
		}

		return stack, nil
	}

	return nil, fmt.Errorf("Can't infer types.")
}

func TypeCheck(modules map[string]*Module) error {
	modulesTypeWorld := make(TypeWorld)

	// Loop through all the modules to compute the
	// type world of all the modules by adding each function
	// using it's fully qualified name.
	for k, v := range modules {
		if k != v.Name {
			panic("BUG: names don't match?")
		}

		for _, fn := range v.Funcs {
			fqname := v.Name + ":" + fn.Name
			modulesTypeWorld[fqname] = fn.Type
		}
	}

	// The typeWorlds consists of the typeWorld of all the
	// builtins and the modulesTypeWorld where the
	// modulesTypeWorld can override builtins.
	typeWorlds := NewTypeWorlds(builtins, modulesTypeWorld)

	for k, v := range modules {
		if k != v.Name {
			panic("BUG: names don't match?")
		}

		for _, fn := range v.Funcs {

			types := make([]Type, 0)
			var err error

			for _, node := range fn.FuncNode.Body {
				types, err = InferTypes(node, types, typeWorlds)

				if err != nil {
					return err
				}
			}

			if len(types) != len(fn.Type.RetTypes) {
				return fmt.Errorf("Function `%s` does not return the right amount of values. Wanted %d but got %d.",
					fn.Name, len(fn.Type.RetTypes), len(types))
			}

			for i := 0; i < len(types); i++ {
				if !TypeCompatibleWith(types[i], fn.Type.RetTypes[i]) {
					return &TypeError{
						Wanted: fn.Type.RetTypes[i],
						Got:    types[i],
						Token:  fn.FuncNode.Token,
						Extra:  fmt.Sprintf("(in returned values of function `%s`)", fn.Name),
					}
				}
			}
		}
	}

	return nil
}
