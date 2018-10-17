package gocat

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	tz    Tokenizer
	tkbuf []*Token
}

type ParserError struct {
	Token *Token
	Msg   string
}

func (pe *ParserError) Error() string {
	return fmt.Sprintf("Parser error %s: %s",
		pe.Token.Pos,
		pe.Msg)
}

func NewParser(tz Tokenizer) *Parser {
	return &Parser{
		tz:    tz,
		tkbuf: make([]*Token, 0),
	}
}

func (p *Parser) readbuf() *Token {
	if len(p.tkbuf) == 0 {
		return nil
	}

	it := p.tkbuf[0]
	p.tkbuf = p.tkbuf[1:]

	return it
}

func (p *Parser) read() (*Token, error) {
	it := p.readbuf()

	if it != nil {
		return it, nil
	}

	tk, err := p.tz.Next()

	if err != nil {
		return nil, err
	}

	return tk, nil
}

func (p *Parser) unread(tk *Token) {
	p.tkbuf = append(p.tkbuf, tk)
}

func (p *Parser) parseData() (Node, error) {
	// Next token must be LITINT or LITFLOAT or IDENT.
	tk, err := p.read()

	if err != nil {
		return nil, err
	}

	switch tk.Type {
	case TT_LITINT:
		iv, err := strconv.ParseInt(tk.SVal, 10, 64)

		if err != nil {
			return nil, &ParserError{
				Token: tk,
				Msg:   fmt.Sprintf("`%s` is not a valid integer literal.", tk.SVal),
			}
		}

		return &LitIntNode{
			Value: iv,
			Token: tk,
		}, nil
	case TT_LITFLOAT:
		fv, err := strconv.ParseFloat(tk.SVal, 64)

		if err != nil {
			return nil, &ParserError{
				Token: tk,
				Msg:   fmt.Sprintf("`%s` is not a valid float literal.", tk.SVal),
			}
		}

		return &LitFloatNode{
			Value: fv,
			Token: tk,
		}, nil
	case TT_IDENT:
		return &VerbNode{
			Verb:  tk.SVal,
			Token: tk,
		}, nil
	case TT_QUOT:
		// Next token must be IDENT
		tk, err = p.read()

		if tk.Type != TT_IDENT {
			return nil, &ParserError{
				Token: tk,
				Msg:   fmt.Sprintf("Expected identifier but got `%s`.", tk.SVal),
			}
		}

		return &QuotNode{
			Ident: tk.SVal,
			Token: tk,
		}, nil
	default:
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected literal but got `%s`.", tk.SVal),
		}
	}
}

func (p *Parser) parseArg() (Arg, error) {
	tk, err := p.read()

	if err != nil {
		return VoidArg, nil
	}

	if tk.Type != TT_LPAREN {
		return VoidArg, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected `(` but got `%s`.", tk.SVal),
		}
	}

	tk, err = p.read()

	if err != nil {
		return VoidArg, nil
	}

	if tk.Type != TT_IDENT {
		return VoidArg, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected identifier but got `%s`.", tk.SVal),
		}
	}

	aname := tk.SVal

	tp, err := p.parseType()

	if err != nil {
		return VoidArg, err
	}

	tk, err = p.read()

	if err != nil {
		return VoidArg, err
	}

	if tk.Type != TT_RPAREN {
		return VoidArg, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected `)` but got `%s`.", tk.SVal),
		}
	}

	return Arg{
		Name: aname,
		Type: tp,
	}, nil
}

func (p *Parser) parseType() (Type, error) {
	tk, err := p.read()

	if err != nil {
		return InvalidType, err
	}

	switch tk.Type {
	case TT_IDENT:
		return &PrimType{
			Type: tk.SVal,
		}, nil
	case TT_LCBRACKET:
		types := make([]Type, 0)

		for {
			done := false

			tk, err := p.read()

			if err != nil {
				return InvalidType, err
			}

			switch tk.Type {
			case TT_IDENT:
				p.unread(tk)
				typ, err := p.parseType()

				if err != nil {
					return InvalidType, err
				}

				types = append(types, typ)
			case TT_LCBRACKET:
				return InvalidType, &ParserError{
					Token: tk,
					Msg:   fmt.Sprintf("Unexpected %s. Union types can not be nested.", tk.SVal),
				}
			case TT_RCBRACKET:
				done = true
			default:
				return InvalidType, &ParserError{
					Token: tk,
					Msg:   fmt.Sprintf("Expected identifier but got `%s`.", tk.SVal),
				}
			}

			if done {
				break
			}
		}

		ut, err := NewUnionType(types)

		if err != nil {
			return InvalidType, &ParserError{
				Token: tk,
				Msg:   err.Error(),
			}
		}

		return ut, nil
	}

	return InvalidType, &ParserError{
		Token: tk,
		Msg:   fmt.Sprintf("`%s` is not a type.", tk.SVal),
	}
}

func (p *Parser) Funcs() ([]*FuncNode, error) {
	return p.parseFuncs()
}

func (p *Parser) parseFuncs() ([]*FuncNode, error) {
	funcs := make([]*FuncNode, 8) // TODO: resize later
	fj := 0

	for {
		tk, err := p.read()

		if err != nil {
			return nil, err
		}

		if tk.Type == TT_EOF {
			break
		}

		if tk.Type != TT_LPAREN {
			return nil, &ParserError{
				Token: tk,
				Msg:   fmt.Sprintf("Expected `(` but got `%s`.", tk.SVal),
			}
		}

		p.unread(tk)

		fn, err := p.parseFunc()

		if err != nil {
			return nil, err
		}

		fn_, ok := fn.(*FuncNode)

		if !ok {
			panic("BUG: didn't get *FuncNode")
		}

		funcs[fj] = fn_
		fj++
	}

	return funcs[:fj], nil
}

func (p *Parser) parseFunc() (Node, error) {
	// next token must be FUNC

	tk, err := p.read()

	firsttk := tk

	if err != nil {
		return nil, err
	}

	if tk.Type != TT_FUNC {
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected `func` but got `%s`.", tk.SVal),
		}
	}

	// then the next token must be IDENT

	tk, err = p.read()

	if err != nil {
		return nil, err
	}

	if tk.Type != TT_IDENT {
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected identifier but got `%s`.", tk.SVal),
		}
	}

	funcname := tk.SVal

	if strings.ContainsRune(funcname, ':') {
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("`:` is not allowed in identifiers in this context. Offending identifier is `%s`.", tk.SVal),
		}
	}

	args := make([]Arg, 8) // TODO: resize later
	aj := 0

	// then the arguments follow. which is at least one LPAREN then until RPAREN
	tk, err = p.read()

	if err != nil {
		return nil, err
	}

	if tk.Type != TT_LBRACKET {
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected `[` but got `%s`.", tk.SVal),
		}
	}

	for {
		done := false

		tk, err = p.read()

		if err != nil {
			return nil, err
		}

		switch tk.Type {
		case TT_RBRACKET:
			done = true
		case TT_LPAREN:
			p.unread(tk)

			arg, err := p.parseArg()

			if err != nil {
				return nil, err
			}

			args[aj] = arg
			aj++
		default:
			return nil, &ParserError{
				Token: tk,
				Msg:   fmt.Sprintf("Expected `(` or `)` but got `%s`", tk.SVal),
			}
		}

		if done {
			break
		}
	}

	// Then return types...
	tk, err = p.read()

	if err != nil {
		return nil, err
	}

	if tk.Type != TT_LBRACKET {
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected `[` but got `%s`.", tk.SVal),
		}
	}

	rets := make([]Type, 0, 1)

	for {
		done := false

		tk, err = p.read()

		if err != nil {
			return nil, err
		}

		switch tk.Type {
		case TT_RBRACKET:
			done = true

		default:
			p.unread(tk)

			ret, err := p.parseType()

			if err != nil {
				return nil, err
			}

			rets = append(rets, ret)
		}

		if done {
			break
		}
	}

	tk, err = p.read()

	if tk.Type != TT_LCBRACKET {
		return nil, &ParserError{
			Token: tk,
			Msg:   fmt.Sprintf("Expected `{` but got `%s`.", tk.SVal),
		}
	}

	bodies := make([]Node, 8) // TODO: resize later
	bj := 0

	for {
		done := false

		tk, err = p.read()

		if err != nil {
			return nil, err
		}

		switch tk.Type {
		case TT_RCBRACKET:
			done = true
		case TT_IF:
			p.unread(tk)

			ifn, err := p.parseIf()

			if err != nil {
				return nil, err
			}

			bodies[bj] = ifn
			bj++
		default:
			p.unread(tk)

			sexp, err := p.parseExp()

			if err != nil {
				return nil, err
			}

			bodies[bj] = sexp
			bj++
		}

		if done {
			break
		}
	}

	return &FuncNode{
		Args:     args[:aj],
		RetTypes: rets,
		Body:     bodies[:bj],
		Token:    firsttk,
		Name:     funcname,
	}, nil
}

func (p *Parser) parseIf() (Node, error) {
	panic("BUG")
}

func (p *Parser) parseExp() (Node, error) {
	var firsttk *Token = nil

	nodes := make([]Node, 8) //TODO: resize this later
	nj := 0

	for {
		tk, err := p.read()

		if firsttk == nil {
			firsttk = tk
		}

		if err != nil {
			return nil, err
		}

		switch tk.Type {
		case TT_LITINT, TT_LITFLOAT, TT_IDENT, TT_QUOT:
			p.unread(tk)
			node, err := p.parseData()

			if err != nil {
				return nil, err
			}

			nodes[nj] = node
			nj++
		case TT_SEMICOLON:
			return &ExpNode{
				Exps:  nodes[:nj],
				Token: firsttk,
			}, nil
		default:
			return nil, &ParserError{
				Token: tk,
				Msg:   fmt.Sprintf("Expected literal, identifier, `;` or `'` but got `%s`.", tk.SVal),
			}
		}
	}

	return nil, nil
}
