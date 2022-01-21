package table

import (
	"github.com/goal-web/contracts"
	"github.com/goal-web/querybuilder"
	"github.com/qbhy/goal/application"
	"github.com/qbhy/goal/exceptions"
)

type table struct {
	contracts.QueryBuilder
	connection contracts.DBConnection
	tx         contracts.DBTx

	primaryKey string
}

func getTable(name string) *table {
	builder := querybuilder.NewQuery(name)
	instance := &table{
		QueryBuilder: builder,
		primaryKey:   "id",
	}
	builder.Bind(instance)
	return instance
}

// Query 将使用默认 connection
func Query(name string) *table {
	return getTable(name).SetConnection(application.Get("db").(contracts.DBConnection))
}

// WithConnection 使用指定链接
func WithConnection(name string, connection interface{}) *table {
	return getTable(name).SetConnection(connection)
}

// WithTX 使用TX
func WithTX(name string, tx contracts.DBTx) contracts.QueryBuilder {
	return getTable(name).SetTX(tx)
}

// SetConnection 参数要么是 contracts.DBConnection 要么是 string
func (this *table) SetConnection(connection interface{}) *table {
	if conn, ok := connection.(contracts.DBConnection); ok {
		this.connection = conn
	} else {
		this.connection = application.Get("db.factory").(contracts.DBFactory).Connection(connection.(string))
	}
	return this
}

// SetPrimaryKey 设置主键
func (this *table) SetPrimaryKey(name string) *table {
	this.primaryKey = name
	return this
}

// getExecutor 获取 sql 语句的执行者
func (this *table) getExecutor() contracts.SqlExecutor {
	if this.tx != nil {
		return this.tx
	}
	return this.connection
}

// SetTX 参数必须是 contracts.DBTx 实例
func (this *table) SetTX(tx interface{}) contracts.QueryBuilder {
	this.tx = tx.(contracts.DBTx)
	return this
}
func (this *table) Delete() int64 {
	sql, bindings := this.DeleteSql()
	result, err := this.getExecutor().Exec(sql, bindings...)
	if err != nil {
		panic(DeleteException{exceptions.WithError(err, contracts.Fields{"sql": sql, "bindings": bindings})})
	}
	num, err := result.RowsAffected()
	if err != nil {
		panic(DeleteException{exceptions.WithError(err, contracts.Fields{"sql": sql, "bindings": bindings})})
	}
	return num
}
