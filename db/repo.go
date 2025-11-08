package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repo Repository 的基本实现，方便快速组合
type Repo[T any] struct {
	db *gorm.DB
	pk string // 默认 id
}

// NewRepo 新建 Repository 实例
func NewRepo[T any](db *gorm.DB, pk ...string) Repo[T] {
	r := Repo[T]{db: db, pk: "id"}
	for _, v := range pk {
		if v != "" {
			r.pk = v
		}
	}
	return r
}

// DB 返回 GORM 实例
func (r *Repo[T]) DB() *gorm.DB {
	return r.db
}

// Transaction 自动事务
func (r *Repo[T]) Transaction(f func(repo *Repo[T]) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return f(&Repo[T]{db: tx, pk: r.pk})
	})
}

// NewQueryBuilder 返回查询构建器
func (r *Repo[T]) NewQueryBuilder() *QueryBuilder[T] {
	return NewQueryBuilder[T](r.db)
}

// Create 创建数据
func (r *Repo[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

// DeleteBy 根据参数 expr 删除数据
func (r *Repo[T]) DeleteBy(expr *Expression) (int64, error) {
	var entity T
	res := r.db.Model(&entity).Scopes(Scopes(expr)).Delete(&entity)
	return res.RowsAffected, res.Error
}

// DeleteByID 根据主键删除数据
func (r *Repo[T]) DeleteByID(id any) error {
	var entity T
	return r.db.Delete(&entity, r.pk, id).Error
}

// DeleteByAndReturn 根据参数 expr 删除数据，并且返回被删除的数据
func (r *Repo[T]) DeleteByAndReturn(expr *Expression) (int64, []T, error) {
	var entities []T
	res := r.db.Clauses(&clause.Returning{}).Scopes(Scopes(expr)).Delete(&entities)
	if err := res.Error; err != nil {
		return 0, nil, err
	}
	return res.RowsAffected, entities, nil
}

// DeleteByIDAndReturn 根据主键删除数据，并且返回被删除的数据
func (r *Repo[T]) DeleteByIDAndReturn(id any) (*T, error) {
	var entity T
	err := r.db.Clauses(&clause.Returning{}).Delete(&entity, r.pk, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateBy 根据参数 expr 更新数据
// 能够有效解决零值被忽略的问题
func (r *Repo[T]) UpdateBy(expr *Expression, values map[string]any) (int64, error) {
	var entity T
	res := r.db.Model(&entity).Scopes(Scopes(expr)).Updates(values)
	return res.RowsAffected, res.Error
}

// UpdateByID 根据主键更新数据
func (r *Repo[T]) UpdateByID(id any, values map[string]any) error {
	var entity T
	return r.db.Model(&entity).Where(r.pk, id).Updates(values).Error
}

// UpdateColumnByID 根据主键更新指定的列
func (r *Repo[T]) UpdateColumnByID(id any, column string, value any) error {
	var entity T
	return r.db.Model(&entity).Where(r.pk, id).UpdateColumn(column, value).Error
}

// UpdateByAndReturn 根据主键更新数据
func (r *Repo[T]) UpdateByAndReturn(expr *Expression, values map[string]any) (int64, []T, error) {
	var entities []T
	res := r.db.Model(&entities).Clauses(&clause.Returning{}).Scopes(Scopes(expr)).Updates(values)
	if res.Error != nil {
		return 0, nil, res.Error
	}
	return res.RowsAffected, entities, nil
}

func (r *Repo[T]) UpdateByIDAndReturn(id any, values map[string]any) (int64, *T, error) {
	var entity T
	res := r.db.Clauses(&clause.Returning{}).Model(&entity).Where(r.pk, id).Updates(values)
	if res.Error != nil {
		return 0, nil, res.Error
	}
	return res.RowsAffected, &entity, nil
}

// GetByID 根据主键查询数据
// 参数 ctx 可以传 nil
func (r *Repo[T]) GetByID(id any) (*T, error) {
	var entity T
	err := r.db.Where(r.pk, id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetBy 根据参数 expr 查询一条
// 参数 ctx 可以传 nil
func (r *Repo[T]) GetBy(expr ...*Expression) (*T, error) {
	var entity T
	err := r.db.Scopes(wraps(expr...)).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindBy 根据参数 expr 查询列表
// 参数 ctx 可以传 nil
func (r *Repo[T]) FindBy(expr ...*Expression) ([]T, error) {
	var items []T
	err := r.db.Scopes(wraps(expr...)).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repo[T]) Count(expr ...*Expression) (int64, error) {
	var entity T
	var total int64
	err := r.db.Model(&entity).Scopes(wraps(expr...)).Count(&total).Error
	return total, err
}
