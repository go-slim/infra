package db

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Expression)(nil)

type Expression struct {
	clauses []clause.Expression
}

// Build 实现 clause.Expression 接口
func (e *Expression) Build(builder clause.Builder) {
	if and := clause.And(e.clauses...); and != nil {
		and.Build(builder)
	}
}

func (e *Expression) add(expr clause.Expression) *Expression {
	e.clauses = append(e.clauses, expr)
	return e
}

func (e *Expression) Eq(col string, val any) *Expression {
	return e.add(clause.Eq{Column: col, Value: val})
}

func (e *Expression) Neq(col string, val any) *Expression {
	return e.add(clause.Neq{Column: col, Value: val})
}

func (e *Expression) Lt(col string, val any) *Expression {
	return e.add(clause.Lt{Column: col, Value: val})
}

func (e *Expression) Lte(col string, val any) *Expression {
	return e.add(clause.Lte{Column: col, Value: val})
}

func (e *Expression) Gt(col string, val any) *Expression {
	return e.add(clause.Gt{Column: col, Value: val})
}

func (e *Expression) Gte(col string, val any) *Expression {
	return e.add(clause.Gte{Column: col, Value: val})
}

func (e *Expression) Between(col string, less, more any) *Expression {
	return e.add(between{col, less, more})
}

func (e *Expression) NotBetween(col string, less, more any) *Expression {
	return e.add(clause.Not(between{col, less, more}))
}

func (e *Expression) IsNull(col string) *Expression {
	return e.add(null{col})
}

func (e *Expression) NotNull(col string) *Expression {
	return e.add(clause.Not(null{col}))
}

func (e *Expression) Like(col, tpl string) *Expression {
	return e.add(Like(col, tpl))
}

func (e *Expression) NotLike(col, tpl string) *Expression {
	return e.add(NotLike(col, tpl))
}

func (e *Expression) Contains(col, tpl string) *Expression {
	return e.add(Contains(col, tpl))
}

func (e *Expression) NotContains(col, tpl string) *Expression {
	return e.add(NotContains(col, tpl))
}

func (e *Expression) HasPrefix(col, prefix string) *Expression {
	return e.add(HasPrefix(col, prefix))
}

func (e *Expression) NotPrefix(col, prefix string) *Expression {
	return e.add(NotPrefix(col, prefix))
}

func (e *Expression) HasSuffix(col, suffix string) *Expression {
	return e.add(HasSuffix(col, suffix))
}

func (e *Expression) NotSuffix(col, suffix string) *Expression {
	return e.add(NotSuffix(col, suffix))
}

func (e *Expression) In(col string, values []any) *Expression {
	return e.add(clause.IN{Column: col, Values: values})
}

func (e *Expression) NotIn(col string, values []any) *Expression {
	return e.add(clause.Not(clause.IN{Column: col, Values: values}))
}

func (e *Expression) Exists(expr any) *Expression {
	return e.add(exists{expr})
}

func (e *Expression) NotExists(expr any) *Expression {
	return e.add(clause.Not(exists{expr}))
}

// TODO 更好的方式实现
func (e *Expression) When(condition bool, then func(e *Expression), elses ...func(e *Expression)) *Expression {
	if condition {
		then(e)
	} else {
		for _, els := range elses {
			els(e)
		}
	}
	return e
}

func (e *Expression) Or(expr ...clause.Expression) *Expression {
	if len(expr) == 0 {
		return e
	}
	if len(e.clauses) == 0 {
		e.clauses = expr
		return e
	}
	e.clauses = []clause.Expression{
		clause.Or(
			clause.And(e.clauses...),
			clause.And(expr...),
		),
	}
	return e
}

func (e *Expression) And(expr ...clause.Expression) *Expression {
	if len(expr) == 0 {
		return e
	}
	if len(e.clauses) == 0 {
		e.clauses = expr
		return e
	}
	e.clauses = []clause.Expression{
		clause.And(
			clause.And(e.clauses...),
			clause.And(expr...),
		),
	}
	return e
}

func (e *Expression) Not(expr ...clause.Expression) *Expression {
	if len(expr) == 0 {
		return e
	}
	return e.add(clause.Not(expr...))
}

func (e *Expression) Where(expr ...clause.Expression) *Expression {
	return e.And(expr...)
}

func (e *Expression) AlwaysTrue() *Expression {
	// 使用原生 SQL 片段，避免被加引号
	return e.add(clause.Expr{SQL: "1=1"})
}

// Scopes 将 expr 自动转换成数据库查询方法
//
// 参数 expr 支持下面几种类型：
//
//   - Expression
//   - clause.Expression
//   - func(*gorm.DB) *gorm.DB
//
// 其它情况在执行查询时报错。
func Scopes(expr any) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return scopes(tx, expr)
	}
}

func wraps(expr ...*Expression) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		for _, e := range expr {
			tx = scopes(tx, e)
		}
		return tx
	}
}

func scopes(tx *gorm.DB, expr any) *gorm.DB {
	if expr == nil {
		return tx
	}

	switch v := expr.(type) {
	case *Expression:
		for _, express := range v.clauses {
			tx = tx.Where(express)
		}
		return tx
	case clause.Expression:
		return tx.Where(v)
	case func(db2 *gorm.DB) *gorm.DB:
		return v(tx)
	case string:
		return tx.Where(v)
	default:
		_ = tx.AddError(fmt.Errorf("unsupported expression type: %T", expr))
		return tx
	}
}
