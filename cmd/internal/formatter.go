package internal

import (
	"fmt"
	"strings"
)

type Formatter struct {
	indentUnit int
	indent     int
	builder    strings.Builder
}

func (f *Formatter) SetIndentUnit(n int) {
	f.indentUnit = n
}
func (f *Formatter) GetIndentUnit() int {
	if f.indentUnit <= 0 {
		return 2
	}
	return f.indentUnit
}
func (f *Formatter) Indent() {
	f.indent += 1
}
func (f *Formatter) Deindent() {
	f.indent -= 1
}
func (f *Formatter) P(format string, a ...interface{}) {
	spaces := strings.Repeat(" ", f.indent*f.GetIndentUnit())
	f.builder.WriteString(spaces + fmt.Sprintf(format, a...) + "\n")
}
func (f *Formatter) PI(format string, a ...interface{}) {
	f.P(format, a...)
	f.Indent()
}
func (f *Formatter) PD(format string, a ...interface{}) {
	f.Deindent()
	f.P(format, a...)
}
func (f *Formatter) PDI(format string, a ...interface{}) {
	f.Deindent()
	f.P(format, a...)
	f.Indent()
}
func (f *Formatter) String() string {
	return f.builder.String()
}
