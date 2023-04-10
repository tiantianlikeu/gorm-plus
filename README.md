

# Gorm-plus

fork 基于 Gorm-plus，详见[https://github.com/acmestack/gorm-plus]()
Gorm-plus是基于Gorm的增强版，类似Mybatis-plus语法。

## 特性

- [x] 无侵入，只做增强不做改变
- [x] 强大的CRUD 操作，内置通用查询，不需要任何配置，即可构建复杂条件查询
- [x] 支持指针字段形式查询，方便编写各类查询条件，无需再担心字段写错
- [x] 支持主键自动生成
- [x] 内置分页插件

## 事务
- 在需要事务的场景，可以使用gplus.Begin()开启事务，获取*gorm.DB
- 所有的dao方法，均支持传入*gorm.DB，后续的操作均以传入的为准

## 用法
- 更多详细用法，详见[https://github.com/acmestack/gorm-plus]


```go

type Student struct {
    ID        int
    Name      string
    Age       uint8
    Email     string
    Birthday  time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}

var gormDb *gorm.DB

func init() {
    dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
    var err error
    gormDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })

	if err != nil {
        log.Println(err)
    }
    gorm_plus.Init(gormDb)
}

func main() {
    var student Student
    // 创建表
    gormDb.AutoMigrate(student)

    // 插入数据
    studentItem := Student{Name: "zhangsan", Age: 18, Email: "123@11.com", Birthday: time.Now()}
    gplus.Insert(&studentItem)

    // 根据Id查询数据
    studentResult, resultDb := gplus.SelectById[Student](studentItem.ID)
    log.Printf("error:%v\n", resultDb.Error)
    log.Printf("RowsAffected:%v\n", resultDb.RowsAffected)
    log.Printf("studentResult:%+v\n", studentResult)

    // 根据条件查询
    query, model := gplus.NewQuery[Student]()
    query.Eq(&model.Name, "zhangsan")
    studentResult, resultDb = gplus.SelectOne(query)
    log.Printf("error:%v\n", resultDb.Error)
    log.Printf("RowsAffected:%v\n", resultDb.RowsAffected)
    log.Printf("studentResult:%+v\n", studentResult)

    // 根据Id更新
    studentItem.Name = "lisi"
    resultDb = gplus.UpdateById[Student](&studentItem)
    log.Printf("error:%v\n", resultDb.Error)
    log.Printf("RowsAffected:%v\n", resultDb.RowsAffected)

    // 根据条件更新
    query, model = gplus.NewQuery[Student]()
    query.Eq(&model.Name, "lisi").Set(&model.Age, 35)
    resultDb = gplus.Update[Student](query)
    log.Printf("error:%v\n", resultDb.Error)
    log.Printf("RowsAffected:%v\n", resultDb.RowsAffected)

    // 根据Id删除
    resultDb = gplus.DeleteById[Student](studentItem.ID)
    log.Printf("error:%v\n", resultDb.Error)
    log.Printf("RowsAffected:%v\n", resultDb.RowsAffected)
    
    // 根据条件删除
    query, model = gplus.NewQuery[Student]()
    query.Eq(&model.Name, "zhangsan")
    resultDb = gplus.Delete[Student](query)
    log.Printf("error:%v\n", resultDb.Error)
    log.Printf("RowsAffected:%v\n", resultDb.RowsAffected)
}
```

## golang代码详细用法示例 表转为结构体

```go
package main

import (
	"fmt"
	"github.com/tiantianlikeu/converter"
)

func main() {
	t2t := converter.NewTable2Struct()
	// 个性化配置
	t2t.Config(&converter.T2tConfig{
		StructNameToHump:  true,
		RmTagIfUcFirsted:  false, // 如果字段首字母本来就是大写, 就不添加tag, 默认false添加, true不添加
		TagToLower:        false, // tag的字段名字是否转换为小写, 如果本身有大写字母的话, 默认false不转
		UcFirstOnly:       false, // 字段首字母大写的同时, 是否要把其他字母转换为小写,默认false不转换
		JsonTagToHump:     true,  // 驼峰处理
		JsonTagFirstLower: true,  // json  tag 驼峰首字母小写
	})
	err := t2t.
		SavePath("./model.go").
		//SavePath("./wx_user.go").
		Dsn("root:1234@tcp(127.0.0.1:3306)/database?charset=utf8mb4").
		TagKey("gorm").
		PackageName("packageName").
		RealNameMethod("RealNameMethod").
		EnableJsonTag(true).
		EnableFormTag(true).
		Table("Table").
		Run()
	fmt.Println(err)
}

```