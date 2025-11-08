package db

import (
	"gorm.io/gorm/clause"
)

type null struct {
	Column any
}

func (n null) Build(builder clause.Builder) {
	builder.WriteQuoted(n.Column)
	builder.WriteString(" IS NULL") //nolint: errcheck
}

func (n null) NegationBuild(builder clause.Builder) {
	builder.WriteQuoted(n.Column)
	builder.WriteString(" IS NOT NULL") //nolint: errcheck
}

type between struct {
	Column any
	Less   any
	More   any
}

func (b between) Build(builder clause.Builder) {
	b.build(builder, " BETWEEN ")
}

func (b between) NegationBuild(builder clause.Builder) {
	b.build(builder, " NOT BETWEEN ")
}

// nolint: errcheck
func (b between) build(builder clause.Builder, op string) {
	builder.WriteQuoted(b.Column)
	builder.WriteString(op)
	builder.AddVar(builder, b.Less)
	builder.WriteString(" And ")
	builder.AddVar(builder, b.More)
}

type exists struct {
	expr any
}

func (e exists) Build(builder clause.Builder) {
	e.build(builder, "EXISTS")
}

func (e exists) NegationBuild(builder clause.Builder) {
	e.build(builder, "NOT EXISTS")
}

// nolint: errcheck
func (e exists) build(builder clause.Builder, op string) {
	builder.WriteString(op)
	builder.WriteString(" (")
	builder.AddVar(builder, e.expr)
	builder.WriteString(")")
}

func Eq(col string, val any) clause.Expression {
	return clause.Eq{Column: col, Value: val}
}

func Neq(col string, val any) clause.Expression {
	return clause.Neq{Column: col, Value: val}
}

func Lt(col string, val any) clause.Expression {
	return clause.Lt{Column: col, Value: val}
}

func Lte(col string, val any) clause.Expression {
	return clause.Lte{Column: col, Value: val}
}

func Gt(col string, val any) clause.Expression {
	return clause.Gt{Column: col, Value: val}
}

func Gte(col string, val any) clause.Expression {
	return clause.Gte{Column: col, Value: val}
}

func Between(col string, less, more any) clause.Expression {
	return between{col, less, more}
}

func NotBetween(col string, less, more any) clause.Expression {
	return clause.Not(between{col, less, more})
}

func IsNull(col string) clause.Expression {
	return null{col}
}

func NotNull(col string) clause.Expression {
	return clause.Not(null{col})
}

func Like(col, tpl string) clause.Expression {
	return clause.Like{Column: col, Value: "%" + tpl + "%"}
}

func NotLike(col, tpl string) clause.Expression {
	return clause.Not(clause.Like{Column: col, Value: "%" + tpl + "%"})
}

func Contains(col, tpl string) clause.Expression {
	return Like(col, tpl)
}

func NotContains(col, tpl string) clause.Expression {
	return NotLike(col, tpl)
}

func HasPrefix(col, prefix string) clause.Expression {
	return clause.Like{Column: col, Value: prefix + "%"}
}

func NotPrefix(col, prefix string) clause.Expression {
	return clause.Not(HasPrefix(col, prefix))
}

func HasSuffix(col, suffix string) clause.Expression {
	return clause.Like{Column: col, Value: "%" + suffix}
}

func NotSuffix(col, suffix string) clause.Expression {
	return clause.Not(HasSuffix(col, suffix))
}

func In[T any](col string, values []T) clause.Expression {
	return clause.IN{Column: col, Values: toSlices(values)}
}

func NotIn[T any](col string, values []T) clause.Expression {
	return clause.Not(clause.IN{Column: col, Values: toSlices(values)})
}

func Exists(expr any) clause.Expression {
	return exists{expr}
}

func toSlices[T any](values []T) []any {
	args := make([]any, len(values))
	for i, v := range values {
		args[i] = v
	}
	return args
}

func NotExists(expr any) clause.Expression {
	return clause.Not(exists{expr})
}

func Or(exprs ...clause.Expression) clause.Expression {
	return clause.Or(exprs...)
}

func And(exprs ...clause.Expression) clause.Expression {
	return clause.And(exprs...)
}

func Not(exprs ...clause.Expression) clause.Expression {
	return clause.Not(exprs...)
}
