package db

import (
	"cmp"
	"database/sql"
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// QueryBuilder 查询构造器
// TODO（hupeh）：实现 joins 和表别名
type QueryBuilder[T any] struct {
	db       *gorm.DB
	selects  []string
	omits    []string
	expr     *Expression
	orders   []string
	limit    int
	offset   int
	distinct []any
	preloads []preload
	having   []any
	groups   []string
}

func NewQueryBuilder[T any](db *gorm.DB) *QueryBuilder[T] {
	return &QueryBuilder[T]{expr: &Expression{}, db: db}
}

type preload struct {
	query string
	args  []any
}

type Pager[T any] struct {
	Total  int `json:"total" xml:"total"` // 数据总数
	Page   int `json:"page"  xml:"page"`  // 当前页码
	Limit  int `json:"limit" xml:"limit"` // 数据容量
	Offset int `json:"-"     xml:"-"`     // 偏移量
	Items  []T `json:"items" xml:"items"` // 数据列表
}

func NewPager[T any](page, limit int) *Pager[T] {
	var p Pager[T]
	page = max(cmp.Or(page, 1), 1)
	p.Limit = max(cmp.Or(limit, 10), 1)
	p.Offset = (page - 1) * p.Limit
	p.Page = int(math.Ceil(float64(p.Offset)/float64(p.Limit))) + 1
	p.Items = make([]T, 0, limit)
	return &p
}

func (q *QueryBuilder[T]) Select(columns ...string) *QueryBuilder[T] {
	q.selects = append(q.selects, columns...)
	return q
}

func (q *QueryBuilder[T]) Omit(columns ...string) *QueryBuilder[T] {
	q.omits = append(q.omits, columns...)
	return q
}

func (q *QueryBuilder[T]) Eq(col string, val any) *QueryBuilder[T] {
	q.expr.Eq(col, val)
	return q
}

func (q *QueryBuilder[T]) Neq(col string, val any) *QueryBuilder[T] {
	q.expr.Neq(col, val)
	return q
}

func (q *QueryBuilder[T]) Lt(col string, val any) *QueryBuilder[T] {
	q.expr.Lt(col, val)
	return q
}

func (q *QueryBuilder[T]) Lte(col string, val any) *QueryBuilder[T] {
	q.expr.Lte(col, val)
	return q
}

func (q *QueryBuilder[T]) Gt(col string, val any) *QueryBuilder[T] {
	q.expr.Gt(col, val)
	return q
}

func (q *QueryBuilder[T]) Gte(col string, val any) *QueryBuilder[T] {
	q.expr.Gte(col, val)
	return q
}

func (q *QueryBuilder[T]) Between(col string, less, more any) *QueryBuilder[T] {
	q.expr.Between(col, less, more)
	return q
}

func (q *QueryBuilder[T]) NotBetween(col string, less, more any) *QueryBuilder[T] {
	q.expr.NotBetween(col, less, more)
	return q
}

func (q *QueryBuilder[T]) IsNull(col string) *QueryBuilder[T] {
	q.expr.IsNull(col)
	return q
}

func (q *QueryBuilder[T]) NotNull(col string) *QueryBuilder[T] {
	q.expr.NotNull(col)
	return q
}

func (q *QueryBuilder[T]) Like(col, tpl string) *QueryBuilder[T] {
	q.expr.Like(col, tpl)
	return q
}

func (q *QueryBuilder[T]) NotLike(col, tpl string) *QueryBuilder[T] {
	q.expr.NotLike(col, tpl)
	return q
}

func (q *QueryBuilder[T]) Contains(col, tpl string) *QueryBuilder[T] {
	q.expr.Contains(col, tpl)
	return q
}

func (q *QueryBuilder[T]) NotContains(col, tpl string) *QueryBuilder[T] {
	q.expr.NotContains(col, tpl)
	return q
}

func (q *QueryBuilder[T]) HasPrefix(col, prefix string) *QueryBuilder[T] {
	q.expr.HasPrefix(col, prefix)
	return q
}

func (q *QueryBuilder[T]) NotPrefix(col, prefix string) *QueryBuilder[T] {
	q.expr.NotPrefix(col, prefix)
	return q
}

func (q *QueryBuilder[T]) HasSuffix(col, suffix string) *QueryBuilder[T] {
	q.expr.HasSuffix(col, suffix)
	return q
}

func (q *QueryBuilder[T]) NotSuffix(col, suffix string) *QueryBuilder[T] {
	q.expr.NotSuffix(col, suffix)
	return q
}

func (q *QueryBuilder[T]) In(col string, values []any) *QueryBuilder[T] {
	q.expr.In(col, values)
	return q
}

func (q *QueryBuilder[T]) NotIn(col string, values []any) *QueryBuilder[T] {
	q.expr.NotIn(col, values)
	return q
}

func (q *QueryBuilder[T]) When(condition bool, then func(e *Expression), els ...func(e *Expression)) *QueryBuilder[T] {
	q.expr.When(condition, then, els...)
	return q
}

func (q *QueryBuilder[T]) Or(expr ...clause.Expression) *QueryBuilder[T] {
	q.expr.Or(expr...)
	return q
}

func (q *QueryBuilder[T]) And(expr ...clause.Expression) *QueryBuilder[T] {
	q.expr.And(expr...)
	return q
}

func (q *QueryBuilder[T]) Not(expr ...clause.Expression) *QueryBuilder[T] {
	q.expr.Not(expr...)
	return q
}

func (q *QueryBuilder[T]) OrderBy(value ...string) *QueryBuilder[T] {
	q.orders = append(q.orders, value...)
	return q
}

func (q *QueryBuilder[T]) DescentBy(columns ...string) *QueryBuilder[T] {
	for _, col := range columns {
		q.orders = append(q.orders, col+" DESC")
	}
	return q
}

func (q *QueryBuilder[T]) AscentBy(columns ...string) *QueryBuilder[T] {
	for _, col := range columns {
		q.orders = append(q.orders, col+" ASC")
	}
	return q
}

func (q *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	q.limit = limit
	return q
}

func (q *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	q.offset = offset
	return q
}

func (q *QueryBuilder[T]) Distinct(columns ...any) *QueryBuilder[T] {
	q.distinct = append(q.distinct, columns...)
	return q
}

func (q *QueryBuilder[T]) Having(v any) *QueryBuilder[T] {
	q.having = append(q.having, v)
	return q
}

func (q *QueryBuilder[T]) Group(name string) *QueryBuilder[T] {
	q.groups = append(q.groups, name)
	return q
}

func (q *QueryBuilder[T]) Preload(query string, args ...any) *QueryBuilder[T] {
	q.preloads = append(q.preloads, preload{query, args})
	return q
}

func (q *QueryBuilder[T]) With(expr ...*Expression) *QueryBuilder[T] {
	for _, e := range expr {
		q.And(e.clauses...)
	}
	return q
}

func (q *QueryBuilder[T]) Scopes(tx *gorm.DB) *gorm.DB {
	tx = q.scopesWithoutEffect(tx)
	if q.orders != nil {
		for _, order := range q.orders {
			tx = tx.Order(order)
		}
	}
	if q.limit > 0 {
		tx = tx.Limit(q.limit)
	}
	if q.offset > 0 {
		tx = tx.Offset(q.offset)
	}
	if q.preloads != nil {
		for _, pl := range q.preloads {
			tx = tx.Preload(pl.query, pl.args...)
		}
	}
	return tx
}

func (q *QueryBuilder[T]) scopesWithoutEffect(tx *gorm.DB) *gorm.DB {
	var entity T
	tx = tx.Model(&entity)
	if q.selects != nil {
		tx = tx.Select(q.selects)
	}
	if q.omits != nil {
		tx = tx.Omit(q.omits...)
	}
	if len(q.distinct) > 0 {
		tx = tx.Distinct(q.distinct...)
	}
	for _, having := range q.having {
		tx = tx.Having(having)
	}
	for _, group := range q.groups {
		tx = tx.Group(group)
	}
	return scopes(tx, q.expr)
}

func (q *QueryBuilder[T]) Count() (int64, error) {
	var count int64
	err := q.db.Scopes(q.scopesWithoutEffect).Count(&count).Error
	return count, err
}

// First finds the first record ordered by primary key, matching given conditions conds
func (q *QueryBuilder[T]) First() (*T, error) {
	var entity T
	err := q.db.Scopes(q.Scopes).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Take finds the first record returned by the database in no specified order, matching given conditions conds
func (q *QueryBuilder[T]) Take() (*T, error) {
	var entity T
	err := q.db.Scopes(q.Scopes).Take(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Last finds the last record ordered by primary key, matching given conditions conds
func (q *QueryBuilder[T]) Last() (*T, error) {
	var entity T
	err := q.db.Scopes(q.Scopes).Last(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Find finds all records matching given conditions conds
func (q *QueryBuilder[T]) Find() ([]T, error) {
	var entities []T
	err := q.db.Scopes(q.Scopes).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (q *QueryBuilder[T]) Paginate(page, limit int) (*Pager[T], error) {
	pager := NewPager[T](page, limit)

	q.limit = pager.Limit
	q.offset = pager.Offset

	count, err := q.Count()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return pager, nil
	}

	pager.Items, err = q.Find()
	if err != nil {
		return nil, err
	}

	pager.Total = int(count)

	return pager, nil
}

// Rows 返回行数据迭代器
//
// 使用示例：
//
//	rows, err := q.Eq("name", "jack").Rows()
//	if err != nil {
//	  panic(err)
//	}
//	defer rows.Close()
//	for rows.Channel() {
//	  var user User
//	  db.ScanRows(rows, &user)
//	  // do something
//	}
func (q *QueryBuilder[T]) Rows() (*sql.Rows, error) {
	return q.db.Scopes(q.Scopes).Rows()
}

// Pluck 获取指定列的值
//
// 示例：
//
//	var names []string
//	q.Pluck("name", &names)
//
// 注意，由于 GORM 会清空选择的字段，所以该方法不要使用 Having 等。
func (q *QueryBuilder[T]) Pluck(column string, dest any) error {
	return q.db.Scopes(q.Scopes).Pluck(column, dest).Error
}
