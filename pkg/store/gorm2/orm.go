package gorm2

import (
	"errors"

	"github.com/douyu/jupiter/pkg/util/xdebug"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLCommon ...
type (
	// SQLCommon alias of gorm.SQLCommon
	// SQLCommon = gorm.SQLCommon
	// Callback alias of gorm.Callback
	// Callback = gorm.Callback
	// CallbackProcessor alias of gorm.CallbackProcessor
	// CallbackProcessor = gorm.CallbackProcessor
	// Dialect alias of gorm.Dialect
	Dialect = gorm.Dialector
	// DB ...
	DB = gorm.DB
	// Model ...
	Model = gorm.Model
	// ModelStruct ...
	// ModelStruct = gorm.ModelStruct
	// Field ...
	// Field = gorm.Field
	// FieldStruct ...
	// StructField = gorm.StructField
	// RowQueryResult ...
	// RowQueryResult = gorm.RowQueryResult
	// RowsQueryResult ...
	// RowsQueryResult = gorm.RowsQueryResult
	// Association ...
	Association = gorm.Association
	// Errors ...
	// Errors = gorm.Errors
	// logger ...
	// Logger = gorm.Logger
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound

	errSlowCommand = errors.New("mysql slow command")
)

// Open ...
func Open(dialect string, options *Config) (*DB, error) {
	var inner, err = gorm.Open(mysql.Open(options.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	inner.Logger = inner.Logger.LogMode(logger.LogLevel(options.LogLevel))

	// 设置默认连接配置
	db, err := inner.DB()
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetMaxOpenConns(options.MaxOpenConns)

	if options.ConnMaxLifetime != 0 {
		db.SetConnMaxLifetime(options.ConnMaxLifetime)
	}

	if xdebug.IsDevelopmentMode() {
		inner.Logger.LogMode(logger.Info)
	}

	replace := func(
		processor interface {
			Get(name string) func(*DB)
			Replace(name string, handler func(*DB)) error
		},
		callbackName string,
		interceptors ...Interceptor,
	) {
		old := processor.Get(callbackName)
		var handler = old
		for _, inte := range interceptors {
			handler = inte(options.dsnCfg, callbackName, options)(handler)
		}

		processor.Replace(callbackName, handler)
	}

	replace(
		inner.Callback().Delete(),
		"gorm:delete",
		options.interceptors...,
	)
	replace(
		inner.Callback().Update(),
		"gorm:update",
		options.interceptors...,
	)
	replace(
		inner.Callback().Create(),
		"gorm:create",
		options.interceptors...,
	)
	replace(
		inner.Callback().Query(),
		"gorm:query",
		options.interceptors...,
	)
	replace(
		inner.Callback().Row(),
		"gorm:row",
		options.interceptors...,
	)

	return inner, err
}
