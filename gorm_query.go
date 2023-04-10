package gorm_plus

import (
	"fmt"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

var columnNameMapCache sync.Map
var modelInstanceCache sync.Map

type Query[T any] struct {
	SelectColumns     []string
	DistinctColumns   []string
	QueryBuilder      strings.Builder
	OrBracketBuilder  strings.Builder
	OrBracketArgs     []any
	AndBracketBuilder strings.Builder
	AndBracketArgs    []any
	QueryArgs         []any
	OrderBuilder      strings.Builder
	GroupBuilder      strings.Builder
	HavingBuilder     strings.Builder
	HavingArgs        []any
	LastCond          string
	UpdateMap         map[string]any
	ColumnNameMap     map[uintptr]string
	ConditionMap      map[any]any
}

func NewQuery[T any]() (*Query[T], *T) {
	q := &Query[T]{}
	return q, q.buildColumnNameMap()
}

func NewQueryMap[T any]() (*Query[T], *T) {
	q := &Query[T]{}
	q.ConditionMap = make(map[any]any)
	return q, q.buildColumnNameMap()
}

func (q *Query[T]) Eq(column any, val any) *Query[T] {
	q.addCond(column, val, Eq)
	return q
}

func (q *Query[T]) Ne(column any, val any) *Query[T] {
	q.addCond(column, val, Ne)
	return q
}

func (q *Query[T]) Gt(column any, val any) *Query[T] {
	q.addCond(column, val, Gt)
	return q
}

func (q *Query[T]) Ge(column any, val any) *Query[T] {
	q.addCond(column, val, Ge)
	return q
}

func (q *Query[T]) Lt(column any, val any) *Query[T] {
	q.addCond(column, val, Lt)
	return q
}

func (q *Query[T]) Le(column any, val any) *Query[T] {
	q.addCond(column, val, Le)
	return q
}

func (q *Query[T]) Like(column any, val any) *Query[T] {
	s := fmt.Sprintf("%v", val)
	q.addCond(column, "%"+s+"%", Like)
	return q
}

func (q *Query[T]) NotLike(column any, val any) *Query[T] {
	s := fmt.Sprintf("%v", val)
	q.addCond(column, "%"+s+"%", Not+" "+Like)
	return q
}

func (q *Query[T]) LikeLeft(column any, val any) *Query[T] {
	s := fmt.Sprintf("%v", val)
	q.addCond(column, "%"+s, Like)
	return q
}

func (q *Query[T]) LikeRight(column any, val any) *Query[T] {
	s := fmt.Sprintf("%v", val)
	q.addCond(column, s+"%", Like)
	return q
}

func (q *Query[T]) IsNull(column any) *Query[T] {
	columnName := q.getColumnName(column)
	q.buildAndIfNeed()
	cond := fmt.Sprintf("%s is null", columnName)
	q.QueryBuilder.WriteString(cond)
	return q
}

func (q *Query[T]) IsNotNull(column any) *Query[T] {
	columnName := q.getColumnName(column)
	q.buildAndIfNeed()
	cond := fmt.Sprintf("%s is not null", columnName)
	q.QueryBuilder.WriteString(cond)
	return q
}

func (q *Query[T]) In(column any, val any) *Query[T] {
	q.addCond(column, val, In)
	return q
}

func (q *Query[T]) NotIn(column any, val any) *Query[T] {
	q.addCond(column, val, Not+" "+In)
	return q
}

func (q *Query[T]) Between(column any, start, end any) *Query[T] {
	columnName := q.getColumnName(column)
	q.buildAndIfNeed()
	cond := fmt.Sprintf("%s %s ? and ? ", columnName, Between)
	q.QueryBuilder.WriteString(cond)
	q.QueryArgs = append(q.QueryArgs, start, end)
	return q
}

func (q *Query[T]) NotBetween(column any, start, end any) *Query[T] {
	columnName := q.getColumnName(column)
	q.buildAndIfNeed()
	cond := fmt.Sprintf("%s %s %s ? and ? ", columnName, Not, Between)
	q.QueryBuilder.WriteString(cond)
	q.QueryArgs = append(q.QueryArgs, start, end)
	return q
}

func (q *Query[T]) Distinct(columns ...any) *Query[T] {
	for _, v := range columns {
		columnName := q.ColumnNameMap[reflect.ValueOf(v).Pointer()]
		q.DistinctColumns = append(q.DistinctColumns, columnName)
	}
	return q
}

func (q *Query[T]) And() *Query[T] {
	q.QueryBuilder.WriteString(And)
	q.QueryBuilder.WriteString(" ")
	q.LastCond = And
	return q
}

func (q *Query[T]) AndBracket(bracketQuery *Query[T]) *Query[T] {
	q.AndBracketBuilder.WriteString(And + " " + LeftBracket + bracketQuery.QueryBuilder.String() + RightBracket + " ")
	q.AndBracketArgs = append(q.AndBracketArgs, bracketQuery.QueryArgs...)
	return q
}

func (q *Query[T]) Or() *Query[T] {
	q.QueryBuilder.WriteString(Or)
	q.QueryBuilder.WriteString(" ")
	q.LastCond = Or
	return q
}

func (q *Query[T]) OrBracket(bracketQuery *Query[T]) *Query[T] {
	q.OrBracketBuilder.WriteString(Or + " " + LeftBracket + bracketQuery.QueryBuilder.String() + RightBracket + " ")
	q.OrBracketArgs = append(q.OrBracketArgs, bracketQuery.QueryArgs...)
	return q
}

func (q *Query[T]) Select(columns ...any) *Query[T] {
	for _, v := range columns {
		columnName := q.getColumnName(v)
		q.SelectColumns = append(q.SelectColumns, columnName)
	}
	return q
}

func (q *Query[T]) OrderByDesc(columns ...any) *Query[T] {
	var columnNames []string
	for _, v := range columns {
		columnName := q.getColumnName(v)
		columnNames = append(columnNames, columnName)
	}
	q.buildOrder(Desc, columnNames...)
	return q
}

func (q *Query[T]) OrderByAsc(columns ...any) *Query[T] {
	var columnNames []string
	for _, v := range columns {
		columnName := q.getColumnName(v)
		columnNames = append(columnNames, columnName)
	}
	q.buildOrder(Asc, columnNames...)
	return q
}

func (q *Query[T]) Group(columns ...any) *Query[T] {
	for _, v := range columns {
		columnName := q.getColumnName(v)
		if q.GroupBuilder.Len() > 0 {
			q.GroupBuilder.WriteString(Comma)
		}
		q.GroupBuilder.WriteString(columnName)
	}
	return q
}

func (q *Query[T]) Having(having string, args ...any) *Query[T] {
	q.HavingBuilder.WriteString(having)
	q.HavingArgs = append(q.HavingArgs, args)
	return q
}

func (q *Query[T]) Set(column any, val any) *Query[T] {
	columnName := q.getColumnName(column)
	if q.UpdateMap == nil {
		q.UpdateMap = make(map[string]any)
	}
	q.UpdateMap[columnName] = val
	return q
}

func (q *Query[T]) addCond(column any, val any, condType string) {
	columnName := q.getColumnName(column)
	q.buildAndIfNeed()
	cond := fmt.Sprintf("%s %s ?", columnName, condType)
	q.QueryBuilder.WriteString(cond)
	q.QueryBuilder.WriteString(" ")
	q.LastCond = ""
	q.QueryArgs = append(q.QueryArgs, val)
}

func (q *Query[T]) buildAndIfNeed() {
	if q.LastCond != And && q.LastCond != Or && q.QueryBuilder.Len() > 0 {
		q.QueryBuilder.WriteString(And)
		q.QueryBuilder.WriteString(" ")
	}
}

func (q *Query[T]) buildOrder(orderType string, columns ...string) {
	for _, v := range columns {
		if q.OrderBuilder.Len() > 0 {
			q.OrderBuilder.WriteString(Comma)
		}
		q.OrderBuilder.WriteString(v)
		q.OrderBuilder.WriteString(" ")
		q.OrderBuilder.WriteString(orderType)
	}
}

func (q *Query[T]) buildColumnNameMap() *T {
	// first try to load from cache
	modelType := reflect.TypeOf((*T)(nil)).Elem().String()
	if model, ok := modelInstanceCache.Load(modelType); ok {
		if cachedColumnNameMap, ok := columnNameMapCache.Load(modelType); ok {
			q.ColumnNameMap = cachedColumnNameMap.(map[uintptr]string)
			return model.(*T)
		}
	}

	q.ColumnNameMap = make(map[uintptr]string)
	model := new(T)
	valueOf := reflect.ValueOf(model)
	typeOf := reflect.TypeOf(model)
	for i := 0; i < valueOf.Elem().NumField(); i++ {
		pointer := valueOf.Elem().Field(i).Addr().Pointer()
		field := typeOf.Elem().Field(i)
		tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		name, ok := tagSetting["COLUMN"]
		if ok {
			q.ColumnNameMap[pointer] = name
		} else {
			namingStrategy := schema.NamingStrategy{}
			name = namingStrategy.ColumnName("", field.Name)
			q.ColumnNameMap[pointer] = name
		}
	}

	// store to cache
	modelInstanceCache.Store(modelType, model)
	columnNameMapCache.Store(modelType, q.ColumnNameMap)

	return model
}

func (q *Query[T]) getColumnName(v any) string {
	var columnName string
	valueOf := reflect.ValueOf(v)
	switch valueOf.Kind() {
	case reflect.String:
		columnName = v.(string)
	case reflect.Pointer:
		columnName = q.ColumnNameMap[valueOf.Pointer()]
	}
	return columnName
}
