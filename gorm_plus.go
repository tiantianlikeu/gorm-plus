package gorm_plus

import (
	"database/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
	"reflect"
)

var gormPlus *gorm.DB

func Init(db *gorm.DB) {
	gormPlus = db
}

type Page[T any] struct {
	Current int
	Size    int
	Total   int64
	Records []*T
}

type BaseDao[T any] struct{}

func (dao BaseDao[T]) NewQuery() (*Query[T], *T) {
	q := &Query[T]{}
	return q, q.buildColumnNameMap()
}

func NewPage[T any](current, size int) *Page[T] {
	return &Page[T]{Current: current, Size: size}
}

func Save[T any](entity *T, dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	resultDb := db.Create(entity)
	return resultDb
}

func SaveOrUpdate[T any](entity *T, dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	resultDb := UpdateById(entity)
	affected := resultDb.RowsAffected
	if affected == 0 {
		resultDb = db.Create(entity)
	}
	return resultDb
}

func SaveBatch[T any](entities []*T, dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	if len(entities) == 0 {
		return db
	}
	resultDb := db.CreateInBatches(entities, BatchSize)
	return resultDb
}

func SaveBatchSize[T any](entities []*T, batchSize int, dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	if len(entities) == 0 {
		return db
	}
	if batchSize <= 0 {
		batchSize = BatchSize
	}
	resultDb := db.CreateInBatches(entities, batchSize)
	return resultDb
}

func DeleteById[T any](id any, dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	var entity T
	resultDb := db.Where(getPkColumnName[T](), id).Delete(&entity)
	return resultDb
}

func DeleteByIds[T any](ids any, dbs ...*gorm.DB) *gorm.DB {
	q, _ := NewQuery[T]()
	q.In(getPkColumnName[T](), ids)
	resultDb := Delete[T](q, dbs...)
	return resultDb
}

func Delete[T any](q *Query[T], dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	var entity T
	resultDb := db.Where(q.QueryBuilder.String(), q.QueryArgs...).Delete(&entity)
	return resultDb
}

func DeleteByMap[T any](q *Query[T], dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	for k, v := range q.ConditionMap {
		columnName := q.getColumnName(k)
		q.Eq(columnName, v)
	}
	var entity T
	resultDb := db.Where(q.QueryBuilder.String(), q.QueryArgs...).Delete(&entity)
	return resultDb
}

func UpdateById[T any](entity *T, dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	resultDb := db.Model(entity).Updates(entity)
	return resultDb
}

func Update[T any](q *Query[T], dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	resultDb := db.Model(new(T)).Where(q.QueryBuilder.String(), q.QueryArgs...).Updates(&q.UpdateMap)
	return resultDb
}

func SelectById[T any](id any, dbs ...*gorm.DB) (*T, *gorm.DB) {
	q, _ := NewQuery[T]()
	q.Eq(getPkColumnName[T](), id)
	var entity T
	resultDb := buildCondition(q, dbs...)
	return &entity, resultDb.First(&entity)
}

func SelectByIds[T any](ids any, dbs ...*gorm.DB) ([]*T, *gorm.DB) {
	q, _ := NewQuery[T]()
	q.In(getPkColumnName[T](), ids)
	return SelectList[T](q, dbs...)
}

func SelectOne[T any](q *Query[T], dbs ...*gorm.DB) (*T, *gorm.DB) {
	var entity T
	resultDb := buildCondition(q, dbs...)
	return &entity, resultDb.First(&entity)
}

func Exists[T any](q *Query[T], dbs ...*gorm.DB) (bool, error) {
	_, dbRes := SelectOne[T](q, dbs...)
	return dbRes.RowsAffected > 0, dbRes.Error
}

func SelectList[T any](q *Query[T], dbs ...*gorm.DB) ([]*T, *gorm.DB) {
	resultDb := buildCondition(q, dbs...)
	var results []*T
	resultDb.Find(&results)
	return results, resultDb
}

func SelectListModel[T any, R any](q *Query[T], dbs ...*gorm.DB) ([]*R, *gorm.DB) {
	resultDb := buildCondition(q, dbs...)
	var results []*R
	resultDb.Scan(&results)
	return results, resultDb
}

func SelectListByMap[T any](q *Query[T], dbs ...*gorm.DB) ([]*T, *gorm.DB) {
	resultDb := buildCondition(q, dbs...)
	var results []*T
	resultDb.Find(&results)
	return results, resultDb
}

func SelectListMaps[T any](q *Query[T], dbs ...*gorm.DB) ([]map[string]any, *gorm.DB) {
	resultDb := buildCondition(q, dbs...)
	var results []map[string]any
	resultDb.Find(&results)
	return results, resultDb
}

func SelectPage[T any](page *Page[T], q *Query[T], dbs ...*gorm.DB) (*Page[T], *gorm.DB) {
	total, countDb := SelectCount[T](q, dbs...)
	if countDb.Error != nil {
		return page, countDb
	}
	page.Total = total
	resultDb := buildCondition(q, dbs...)
	var results []*T
	resultDb.Scopes(paginate(page)).Find(&results)
	page.Records = results
	return page, resultDb
}

func SelectPageModel[T any, R any](page *Page[R], q *Query[T], dbs ...*gorm.DB) (*Page[R], *gorm.DB) {
	total, countDb := SelectCount[T](q, dbs...)
	if countDb.Error != nil {
		return page, countDb
	}
	page.Total = total
	resultDb := buildCondition(q, dbs...)
	var results []*R
	resultDb.Scopes(paginate(page)).Scan(&results)
	page.Records = results
	return page, resultDb
}

func SelectPageMaps[T any](page *Page[map[string]any], q *Query[T], dbs ...*gorm.DB) (*Page[map[string]any], *gorm.DB) {
	total, countDb := SelectCount[T](q, dbs...)
	if countDb.Error != nil {
		return page, countDb
	}
	page.Total = total
	resultDb := buildCondition(q, dbs...)
	var results []map[string]any
	resultDb.Scopes(paginate(page)).Find(&results)
	for _, m := range results {
		page.Records = append(page.Records, &m)
	}
	return page, resultDb
}

func SelectCount[T any](q *Query[T], dbs ...*gorm.DB) (int64, *gorm.DB) {
	var count int64
	resultDb := buildCondition(q, dbs...)
	resultDb.Count(&count)
	return count, resultDb
}

func paginate[T any](p *Page[T]) func(db *gorm.DB) *gorm.DB {
	page := p.Current
	pageSize := p.Size
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func buildCondition[T any](q *Query[T], dbs ...*gorm.DB) *gorm.DB {
	db := DB(dbs...)
	resultDb := db.Model(new(T))
	if q != nil {
		if len(q.DistinctColumns) > 0 {
			resultDb.Distinct(q.DistinctColumns)
		}

		if len(q.SelectColumns) > 0 {
			resultDb.Select(q.SelectColumns)
		}

		if q.QueryBuilder.Len() > 0 {

			if q.AndBracketBuilder.Len() > 0 {
				q.QueryArgs = append(q.QueryArgs, q.AndBracketArgs...)
				q.QueryBuilder.WriteString(q.AndBracketBuilder.String())
			}

			if q.OrBracketBuilder.Len() > 0 {
				q.QueryArgs = append(q.QueryArgs, q.OrBracketArgs...)
				q.QueryBuilder.WriteString(q.OrBracketBuilder.String())
			}

			resultDb.Where(q.QueryBuilder.String(), q.QueryArgs...)
		}

		if len(q.ConditionMap) > 0 {
			var condMap = make(map[string]any)
			for k, v := range q.ConditionMap {
				columnName := q.getColumnName(k)
				condMap[columnName] = v
			}
			resultDb.Where(condMap)
		}

		if q.OrderBuilder.Len() > 0 {
			resultDb.Order(q.OrderBuilder.String())
		}

		if q.GroupBuilder.Len() > 0 {
			resultDb.Group(q.GroupBuilder.String())
		}

		if q.HavingBuilder.Len() > 0 {
			resultDb.Having(q.HavingBuilder.String(), q.HavingArgs...)
		}
	}
	return resultDb
}

func getPkColumnName[T any]() string {
	var entity T
	entityType := reflect.TypeOf(entity)
	numField := entityType.NumField()
	var columnName string
	for i := 0; i < numField; i++ {
		field := entityType.Field(i)
		tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		isPrimaryKey := utils.CheckTruth(tagSetting["PRIMARYKEY"], tagSetting["PRIMARY_KEY"])
		if isPrimaryKey {
			name, ok := tagSetting["COLUMN"]
			if !ok {
				namingStrategy := schema.NamingStrategy{}
				name = namingStrategy.ColumnName("", field.Name)
			}
			columnName = name
			break
		}
	}
	if columnName == "" {
		return DefaultPrimaryName
	}
	return columnName
}

func DB(dbs ...*gorm.DB) *gorm.DB {
	if len(dbs) > 0 {
		db := dbs[0]
		return db
	}
	return gormPlus
}

func Begin(opts ...*sql.TxOptions) *gorm.DB {
	db := DB()
	return db.Begin(opts...)
}
