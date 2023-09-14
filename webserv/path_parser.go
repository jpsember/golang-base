package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type PathParseStruct struct {
	text   string
	parts  []string
	cursor int
}

type PathParse = *PathParseStruct

func NewPathParse(text string) PathParse {
	t := &PathParseStruct{
		text: text,
	}
	t.parse()
	return t
}

func (p PathParse) String() string { return p.JSMap().String() }

func (p PathParse) JSMap() JSMap {
	x := NewJSMap()
	x.Put("text", p.text)
	x.Put("parts", JSListWith(p.Parts()))
	return x
}

func (p PathParse) Parts() []string {
	return p.parts
}

func (p PathParse) HasNext() bool {
	return p.cursor < len(p.parts)
}

func (p PathParse) Peek() string {
	if p.HasNext() {
		return p.parts[p.cursor]
	}
	return ""
}

func (p PathParse) PeekInt() int {
	x := p.Peek()
	if x != "" {
		val, err := ParseInt(x)
		if err == nil {
			return int(val)
		}
	}
	return -1
}

func (p PathParse) ReadInt() int {
	x := p.PeekInt()
	p.advance()
	return x
}

func (p PathParse) advance() {
	if p.HasNext() {
		p.cursor++
	}
}

func (p PathParse) Read() string {
	x := p.Peek()
	p.advance()
	return x
}

func (p PathParse) parse() {
	if p.parts != nil {
		return
	}
	c := strings.TrimSpace(p.text)
	strings.TrimSuffix(c, "/")
	substr := strings.Split(c, "/")
	var f []string
	for _, x := range substr {
		x := strings.TrimSpace(x)
		if x == "" || x == "/" {
			continue
		}
		f = append(f, x)
	}
	p.parts = f
}
