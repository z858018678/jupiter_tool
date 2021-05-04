package gorm2

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/douyu/jupiter/pkg/metric"
	"github.com/douyu/jupiter/pkg/trace"
	"github.com/douyu/jupiter/pkg/util/xcolor"
	"github.com/douyu/jupiter/pkg/xlog"
)

// Handler ...
type Handler func(*DB)

// Interceptor ...
type Interceptor func(*DSN, string, *Config) func(next Handler) Handler

func logSQL(sql string, args []interface{}, containArgs bool) string {
	if containArgs {
		return bindSQL(sql, args)
	}
	return sql
}

func debugInterceptor(dsn *DSN, op string, options *Config) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *DB) {
			var now = time.Now()
			next(db)
			var sql = db.Statement.SQL.String()
			fmt.Printf("%-50s[%s] => %s\n", xcolor.Green(dsn.Addr+"/"+dsn.DBName), now.Format("04:05.000"), xcolor.Green("Send: "+logSQL(sql, db.Statement.Vars, true)))
			if db.Error != nil {
				fmt.Printf("%-50s[%s] => %s\n", xcolor.Red(dsn.Addr+"/"+dsn.DBName), time.Now().Format("04:05.000"), xcolor.Red("Erro: "+db.Error.Error()))
			} else {
				fmt.Printf("%-50s[%s] => %s\n", xcolor.Green(dsn.Addr+"/"+dsn.DBName), time.Now().Format("04:05.000"), xcolor.Green("Affected: "+strconv.Itoa(int(db.RowsAffected))))
			}
		}
	}
}

func metricInterceptor(dsn *DSN, op string, options *Config) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *DB) {
			beg := time.Now()
			next(db)
			cost := time.Since(beg)

			// error metric
			if db.Error != nil {
				metric.LibHandleCounter.WithLabelValues(metric.TypeGorm, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "ERR").Inc()
				// todo sql语句，需要转换成脱密状态才能记录到日志
				if db.Error != ErrRecordNotFound {
					options.logger.Error("mysql err", xlog.FieldErr(db.Error), xlog.FieldName(dsn.DBName+"."+db.Statement.Table), xlog.FieldMethod(op))
				} else if options.MetricLevel > 1 {
					options.logger.Warn("record not found", xlog.FieldErr(db.Error), xlog.FieldName(dsn.DBName+"."+db.Statement.Table), xlog.FieldMethod(op))
				}
			} else {
				metric.LibHandleCounter.Inc(metric.TypeGorm, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "OK")
			}

			metric.LibHandleHistogram.WithLabelValues(metric.TypeGorm, dsn.DBName+"."+db.Statement.Table, dsn.Addr).Observe(cost.Seconds())

			if options.SlowThreshold > time.Duration(0) && options.SlowThreshold < cost {
				options.logger.Error(
					"slow",
					xlog.FieldErr(errSlowCommand),
					xlog.FieldMethod(op),
					xlog.FieldExtMessage(logSQL(db.Statement.SQL.String(), db.Statement.Vars, options.DetailSQL)),
					xlog.FieldAddr(dsn.Addr),
					xlog.FieldName(dsn.DBName+"."+db.Statement.Table),
					xlog.FieldCost(cost),
				)
			}
		}
	}
}

func traceInterceptor(dsn *DSN, op string, options *Config) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *DB) {
			if val, ok := db.Get("_context"); ok {
				if ctx, ok := val.(context.Context); ok {
					span, _ := trace.StartSpanFromContext(
						ctx,
						"GORM", // TODO this op value is op or GORM
						trace.TagComponent("mysql"),
						trace.TagSpanKind("client"),
					)
					defer span.Finish()

					// 延迟执行 scope.CombinedConditionSql() 避免sqlVar被重复追加
					next(db)

					span.SetTag("sql.inner", dsn.DBName)
					span.SetTag("sql.addr", dsn.Addr)
					span.SetTag("span.kind", "client")
					span.SetTag("peer.service", "mysql")
					span.SetTag("db.instance", dsn.DBName)
					span.SetTag("peer.address", dsn.Addr)
					span.SetTag("peer.statement", logSQL(db.Statement.SQL.String(), db.Statement.Vars, options.DetailSQL))
					return
				}
			}

			next(db)
		}
	}
}
