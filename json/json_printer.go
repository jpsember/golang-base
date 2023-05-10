package json

import (
	//. "js/base"
	"strings"
)

type JSONPrinter struct {
	Pretty        bool
	StringBuilder *strings.Builder
	indent        int
	indentStack   []int
}

func NewJSONPrinter(pretty bool) *JSONPrinter {
	var m = new(JSONPrinter)
	m.StringBuilder = new(strings.Builder)
	m.Pretty = pretty
	return m
}

func (p *JSONPrinter) PushIndentAdjust(amount int) {
	p.indentStack = append(p.indentStack, p.indent)
	p.indent += amount
}

func (p *JSONPrinter) PopIndent() {
	var s = p.indentStack
	var i = len(s) - 1
	p.indent = s[i]
	p.indentStack = s[:i]
}

func (p *JSONPrinter) Indent() int {
	return p.indent
}

func (p *JSONPrinter) GetPrintResult() string {
	var s = p.StringBuilder.String()
	p.StringBuilder.Reset()
	p.indent = 0
	p.indentStack = p.indentStack[:0]
	return s
}

func (p *JSONPrinter) WriteString(s string) {
	p.StringBuilder.WriteString(s)
}

func PrintJSEntity(jsEntity JSEntity, pretty bool) string {
	var printer = NewJSONPrinter(pretty)
	jsEntity.PrintTo(printer)
	return printer.GetPrintResult()
}

// func prettyPrintJSEntity(arg any) string {
// 	var jsEntity = arg.(JSEntity)
// 	return PrintJSEntity(jsEntity, true)
// }

// func init() {
// 	RegisterBasePrinterType(&JSMap{}, prettyPrintJSEntity)
// 	RegisterBasePrinterType(&JSList{}, prettyPrintJSEntity)
// }
