package gocat

import (
	"fmt"
	"sort"
	"strings"
)

type Node interface {
	IsNode() bool
}

type Arg struct {
	Name string
	Type Type
}

var VoidArg Arg = Arg{}

type Type interface {
	IsType() bool
	String() string
}

type VoidType struct {
}

func (*VoidType) IsType() bool {
	return true
}

func (*VoidType) String() string {
	return "void"
}

type UnionType struct {
	Types []Type
}

func (*UnionType) IsType() bool {
	return true
}

func (ut *UnionType) String() string {
	s := make([]string, 0)
	for _, typ := range ut.Types {
		s = append(s, typ.String())
	}

	return "{" + strings.Join(s, " ") + "}"
}

func NewUnionType(types []Type) (*UnionType, error) {
	sort.Slice(types, func(i, j int) bool {
		return TypeCmp(types[i], types[j]) < 0
	})

	for i := 0; i < len(types); i++ {
		if i != 0 && TypeEqual(types[i], types[i-1]) {
			return nil, fmt.Errorf("Duplicate type `%s` in union type `%s`.", types[i], &UnionType{Types: types})
		}
	}

	return &UnionType{
		Types: types,
	}, nil
}

type PrimType struct {
	Type string
}

func (*PrimType) IsType() bool {
	return true
}

func (pt *PrimType) String() string {
	return pt.Type
}

type ContractType struct {
	Funcs map[string]*FuncType
}

func (*ContractType) IsType() bool {
	return true
}

type FuncType struct {
	ArgTypes []Type
	RetTypes []Type
}

func (*FuncType) IsType() bool {
	return true
}

func (ft *FuncType) String() string {
	args := make([]string, 0)

	for _, argType := range ft.ArgTypes {
		args = append(args, argType.String())
	}

	rets := make([]string, 0)

	for _, retType := range ft.RetTypes {
		rets = append(rets, retType.String())
	}

	return "func{" + strings.Join(args, " ") + " : " + strings.Join(rets, " ") + "}"
}

var InvalidType Type = nil

type RootNode struct {
	Funcs     []*FuncNode
	TypeDecls map[string]*TypeDeclNode
}

func (*RootNode) IsNode() bool {
	return true
}

type TypeDeclNode struct {
	Name string
	Type Type
}

func (*TypeDeclNode) IsNode() bool {
	return true
}

type FuncNode struct {
	Name     string
	Args     []Arg
	Body     []Node
	RetTypes []Type
	Token    *Token
}

func (*FuncNode) IsNode() bool {
	return true
}

type LitIntNode struct {
	Value int64
	Token *Token
}

func (*LitIntNode) IsNode() bool {
	return true
}

type QuotNode struct {
	Ident string
	Token *Token
}

func (*QuotNode) IsNode() bool {
	return true
}

type LitFloatNode struct {
	Value float64
	Token *Token
}

func (*LitFloatNode) IsNode() bool {
	return true
}

type ReadVarNode struct {
	Name  string
	Token *Token
}

func (*ReadVarNode) IsNode() bool {
	return true
}

type IfElseNode struct {
	Condition Node
	ThenBlock []Node
	ElseBlock []Node
	Token     *Token
}

type VerbNode struct {
	Verb  string
	Token *Token
}

func (*VerbNode) IsNode() bool {
	return true
}

type ExpNode struct {
	Exps  []Node
	Token *Token
}

func (*ExpNode) IsNode() bool {
	return true
}

func TypesEqual(ts1 []Type, ts2 []Type) bool {
	if len(ts1) != len(ts2) {
		return false
	}

	for i := 0; i < len(ts1); i++ {
		if !TypeEqual(ts1[i], ts2[i]) {
			return false
		}
	}

	return true
}

func TypeEqual(t1 Type, t2 Type) bool {
	return TypeCmp(t1, t2) == 0
}

func TypeCmp(t1 Type, t2 Type) int {
	// The order of types is:
	// - void type
	// - prim type
	//   - sorted alphabetically
	// - union type

	switch t1.(type) {
	case *VoidType:
		switch t2.(type) {
		case *VoidType:
			return 0
		default:
			return -1 // void type comes before any other type
		}
	case *PrimType:
		switch t2.(type) {
		case *PrimType:
			return strings.Compare(t1.(*PrimType).Type, t2.(*PrimType).Type)
		case *VoidType:
			return 1 // PrimType comes after VoidType
		case *UnionType:
			return -1 // but it comes before UnionType
		}
	case *UnionType:
		switch t2.(type) {
		case *PrimType:
			return 1 // UnionType comes after PrimType
		case *VoidType:
			return 1 // and after VoidType
		case *UnionType:
			// fewer types first / more types second
			ut1 := t1.(*UnionType)
			ut2 := t2.(*UnionType)

			if len(ut1.Types) < len(ut1.Types) {
				return -1
			} else if len(ut1.Types) > len(ut2.Types) {
				return 1
			}

			for i := 0; i < len(ut1.Types); i++ {
				c := TypeCmp(ut1.Types[i], ut2.Types[i])

				if c != 0 {
					return c
				}
			}

			return 0
		}
	}

	panic("BUG: can't compare these types?")
}

func ArgEqual(a1 Arg, a2 Arg) bool {
	return a1.Name == a2.Name && TypeEqual(a1.Type, a2.Type)
}

func ASTEqual(n1 Node, n2 Node) bool {
	switch n1.(type) {
	case *FuncNode:
		switch n2.(type) {
		case *FuncNode:
			fn1 := n1.(*FuncNode)
			fn2 := n2.(*FuncNode)

			if fn1.Name != fn2.Name {
				return false
			}

			if len(fn1.Args) != len(fn2.Args) {
				return false
			}

			if len(fn1.Body) != len(fn2.Body) {
				return false
			}

			if len(fn1.RetTypes) != len(fn2.RetTypes) {
				return false
			}

			for i := 0; i < len(fn1.RetTypes); i++ {
				if !TypeEqual(fn1.RetTypes[i], fn2.RetTypes[i]) {
					return false
				}
			}

			for i := 0; i < len(fn1.Args); i++ {
				if !ArgEqual(fn1.Args[i], fn2.Args[i]) {
					return false
				}
			}

			for i := 0; i < len(fn1.Body); i++ {
				if !ASTEqual(fn1.Body[i], fn2.Body[i]) {
					return false
				}
			}

			return true
		default:
			return false
		}
	case *LitFloatNode:
		switch n2.(type) {
		case *LitFloatNode:
			return n1.(*LitFloatNode).Value == n2.(*LitFloatNode).Value
		default:
			return false
		}
	case *LitIntNode:
		switch n2.(type) {
		case *LitIntNode:
			return n1.(*LitIntNode).Value == n2.(*LitIntNode).Value
		default:
			return false
		}
	case *VerbNode:
		switch n2.(type) {
		case *VerbNode:
			return n1.(*VerbNode).Verb == n2.(*VerbNode).Verb
		default:
			return false
		}
	case *QuotNode:
		switch n2.(type) {
		case *QuotNode:
			return n1.(*QuotNode).Ident == n2.(*QuotNode).Ident
		default:
			return false
		}
	case *ExpNode:
		switch n2.(type) {
		case *ExpNode:
			n1_ := n1.(*ExpNode)
			n2_ := n2.(*ExpNode)

			if len(n1_.Exps) != len(n2_.Exps) {
				return false
			}

			sz := len(n1_.Exps)

			for i := 0; i < sz; i++ {
				if !ASTEqual(n1_.Exps[i], n2_.Exps[i]) {
					return false
				}
			}

			return true
		default:
			return false
		}
	}

	panic("BUG: ASTEqual")
}
