// Package expr は集合演算式（-, +, &, ()）の構文解析および評価機能を提供します。
package expr

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// Set は部屋番号等の整数の集合を表します。
type Set map[int]struct{}

// NewSet はスライスから Set を生成します。
func NewSet(items ...int) Set {
	s := make(Set, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

// Clone は集合の浅いコピーを作成します。
func (s Set) Clone() Set {
	if s == nil {
		return make(Set)
	}
	dst := make(Set, len(s))
	for k := range s {
		dst[k] = struct{}{}
	}
	return dst
}

// ToSlice は集合の全要素を昇順ソートしたスライスとして返します。
func (s Set) ToSlice() []int {
	res := make([]int, 0, len(s))
	for k := range s {
		res = append(res, k)
	}
	slices.Sort(res)
	return res
}

// Union は2つの集合の和集合（a ∪ b）を返します。
func Union(a, b Set) Set {
	res := make(Set, len(a)+len(b))
	for k := range a {
		res[k] = struct{}{}
	}
	for k := range b {
		res[k] = struct{}{}
	}
	return res
}

// Difference は集合 a から b の要素を取り除いた差集合（a \ b）を返します。
func Difference(a, b Set) Set {
	res := make(Set, len(a))
	for k := range a {
		if _, ok := b[k]; !ok {
			res[k] = struct{}{}
		}
	}
	return res
}

// Intersect は2つの集合の積集合（a ∩ b）を返します。
func Intersect(a, b Set) Set {
	if len(a) > len(b) {
		a, b = b, a
	}
	res := make(Set)
	for k := range a {
		if _, ok := b[k]; ok {
			res[k] = struct{}{}
		}
	}
	return res
}

// LookupFunc は識別子（例: "all_rooms.1F"）から部屋集合を取得する関数です。
type LookupFunc func(ident string) (Set, error)

type tokenType int

const (
	tokEOF tokenType = iota
	tokIdent
	tokPlus
	tokMinus
	tokAnd
	tokLParen
	tokRParen
)

type token struct {
	typ tokenType
	val string
	pos int
}

type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input), pos: 0}
}

func (l *lexer) nextToken() (token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return token{typ: tokEOF, pos: l.pos}, nil
	}

	ch := l.input[l.pos]
	pos := l.pos

	switch ch {
	case '+':
		l.pos++
		return token{typ: tokPlus, val: "+", pos: pos}, nil
	case '-':
		l.pos++
		return token{typ: tokMinus, val: "-", pos: pos}, nil
	case '&':
		l.pos++
		return token{typ: tokAnd, val: "&", pos: pos}, nil
	case '(':
		l.pos++
		return token{typ: tokLParen, val: "(", pos: pos}, nil
	case ')':
		l.pos++
		return token{typ: tokRParen, val: ")", pos: pos}, nil
	}

	if isIdentChar(ch) {
		start := l.pos
		for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
			l.pos++
		}
		val := string(l.input[start:l.pos])
		return token{typ: tokIdent, val: val, pos: start}, nil
	}

	return token{}, fmt.Errorf("unexpected character '%c' at position %d", ch, pos)
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func isIdentChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.'
}

type parser struct {
	lex     *lexer
	currTok token
	lookup  LookupFunc
}

func newParser(input string, lookup LookupFunc) (*parser, error) {
	p := &parser{
		lex:    newLexer(input),
		lookup: lookup,
	}
	if err := p.nextToken(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *parser) nextToken() error {
	tok, err := p.lex.nextToken()
	if err != nil {
		return err
	}
	p.currTok = tok
	return nil
}

// Expression = Term ( ("+" | "-") Term )*
func (p *parser) parseExpression() (Set, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.currTok.typ == tokPlus || p.currTok.typ == tokMinus {
		op := p.currTok.typ
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}

		if op == tokPlus {
			left = Union(left, right)
		} else {
			left = Difference(left, right)
		}
	}

	return left, nil
}

// Term = Factor ( "&" Factor )*
func (p *parser) parseTerm() (Set, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.currTok.typ == tokAnd {
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}

		left = Intersect(left, right)
	}

	return left, nil
}

// Factor = IDENT | "(" Expression ")"
func (p *parser) parseFactor() (Set, error) {
	switch p.currTok.typ {
	case tokIdent:
		name := p.currTok.val
		if err := p.nextToken(); err != nil {
			return nil, err
		}

		if p.lookup == nil {
			return nil, fmt.Errorf("lookup function is nil (resolving %q)", name)
		}
		set, err := p.lookup(name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %q: %w", name, err)
		}
		return set, nil

	case tokLParen:
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.currTok.typ != tokRParen {
			return nil, fmt.Errorf("expected ')' at position %d, got %q", p.currTok.pos, p.currTok.val)
		}
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		return expr, nil

	case tokEOF:
		return nil, fmt.Errorf("unexpected end of expression at position %d", p.currTok.pos)

	default:
		return nil, fmt.Errorf("unexpected token %q at position %d", p.currTok.val, p.currTok.pos)
	}
}

// EvalToSet は集合演算式をパース・評価し、結果の Set を返します。
func EvalToSet(expression string, lookup LookupFunc) (Set, error) {
	trimmed := strings.TrimSpace(expression)
	if trimmed == "" {
		return nil, fmt.Errorf("empty expression")
	}

	p, err := newParser(trimmed, lookup)
	if err != nil {
		return nil, err
	}

	resSet, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.currTok.typ != tokEOF {
		return nil, fmt.Errorf("unexpected token %q after expression at position %d", p.currTok.val, p.currTok.pos)
	}

	return resSet, nil
}

// Eval は集合演算式をパース・評価し、昇順ソートされた部屋番号スライスを返します。
func Eval(expression string, lookup LookupFunc) ([]int, error) {
	resSet, err := EvalToSet(expression, lookup)
	if err != nil {
		return nil, err
	}
	return resSet.ToSlice(), nil
}
