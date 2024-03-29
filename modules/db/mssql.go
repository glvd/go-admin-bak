// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"fmt"

	"strconv"
	"strings"

	"github.com/glvd/go-admin/modules/config"
	"github.com/gogf/gf/text/gregex"
)

// Mssql is a Connection of mssql.
type Mssql struct {
	Base
}

// GetMssqlDB return the global mssql connection.
func GetMssqlDB() *Mssql {
	return &Mssql{
		Base: Base{
			DbList: make(map[string]*sql.DB),
		},
	}
}

// GetDelimiter implements the method Connection.GetDelimiter.
func (db *Mssql) GetDelimiter() string {
	return "["
}

// Name implements the method Connection.Name.
func (db *Mssql) Name() string {
	return "mssql"
}

// 从Gf框架复制
// 在执行sql之前对sql进行进一步处理
func (db *Mssql) handleSqlBeforeExec(query string) string {
	index := 0
	str, _ := gregex.ReplaceStringFunc("\\?", query, func(s string) string {
		index++
		return fmt.Sprintf("@p%d", index)
	})

	str, _ = gregex.ReplaceString("\"", "", str)

	return db.parseSql(str)
}

//将MYSQL的SQL语法转换为MSSQL的语法
//1.由于mssql不支持limit写法所以需要对mysql中的limit用法做转换
func (db *Mssql) parseSql(sql string) string {
	//下面的正则表达式匹配出SELECT和INSERT的关键字后分别做不同的处理，如有LIMIT则将LIMIT的关键字也匹配出
	patten := `^\s*(?i)(SELECT)|(LIMIT\s*(\d+)\s*,\s*(\d+))`
	if gregex.IsMatchString(patten, sql) == false {
		//fmt.Println("not matched..")
		return sql
	}

	res, err := gregex.MatchAllString(patten, sql)
	if err != nil {
		//fmt.Println("MatchString error.", err)
		return ""
	}

	index := 0
	keyword := strings.TrimSpace(res[index][0])
	keyword = strings.ToUpper(keyword)

	index++
	switch keyword {
	case "SELECT":
		//不含LIMIT关键字则不处理
		if len(res) < 2 || (strings.HasPrefix(res[index][0], "LIMIT") == false && strings.HasPrefix(res[index][0], "limit") == false) {
			break
		}

		//不含LIMIT则不处理
		if gregex.IsMatchString("((?i)SELECT)(.+)((?i)LIMIT)", sql) == false {
			break
		}

		//判断SQL中是否含有order by
		selectStr := ""
		orderbyStr := ""
		haveOrderby := gregex.IsMatchString("((?i)SELECT)(.+)((?i)ORDER BY)", sql)
		if haveOrderby {
			//取order by 前面的字符串
			queryExpr, _ := gregex.MatchString("((?i)SELECT)(.+)((?i)ORDER BY)", sql)

			if len(queryExpr) != 4 || strings.EqualFold(queryExpr[1], "SELECT") == false || strings.EqualFold(queryExpr[3], "ORDER BY") == false {
				break
			}
			selectStr = queryExpr[2]

			//取order by表达式的值
			orderbyExpr, _ := gregex.MatchString("((?i)ORDER BY)(.+)((?i)LIMIT)", sql)
			if len(orderbyExpr) != 4 || strings.EqualFold(orderbyExpr[1], "ORDER BY") == false || strings.EqualFold(orderbyExpr[3], "LIMIT") == false {
				break
			}
			orderbyStr = orderbyExpr[2]
		} else {
			queryExpr, _ := gregex.MatchString("((?i)SELECT)(.+)((?i)LIMIT)", sql)
			if len(queryExpr) != 4 || strings.EqualFold(queryExpr[1], "SELECT") == false || strings.EqualFold(queryExpr[3], "LIMIT") == false {
				break
			}
			selectStr = queryExpr[2]
		}

		//取limit后面的取值范围
		first, limit := 0, 0
		for i := 1; i < len(res[index]); i++ {
			if len(strings.TrimSpace(res[index][i])) == 0 {
				continue
			}

			if strings.HasPrefix(res[index][i], "LIMIT") || strings.HasPrefix(res[index][i], "limit") {
				first, _ = strconv.Atoi(res[index][i+1])
				limit, _ = strconv.Atoi(res[index][i+2])
				break
			}
		}

		if haveOrderby {
			sql = fmt.Sprintf("SELECT * FROM (SELECT ROW_NUMBER() OVER (ORDER BY %s) as ROWNUMBER_, %s   ) as TMP_ WHERE TMP_.ROWNUMBER_ > %d AND TMP_.ROWNUMBER_ <= %d", orderbyStr, selectStr, first, limit)
		} else {
			if first == 0 {
				first = limit
			} else {
				first = limit - first
			}
			sql = fmt.Sprintf("SELECT * FROM (SELECT TOP %d * FROM (SELECT TOP %d %s) as TMP1_ ) as TMP2_ ", first, limit, selectStr)
		}
	default:
	}
	return sql
}

// QueryWithConnection implements the method Connection.QueryWithConnection.
func (db *Mssql) QueryWithConnection(con string, query string, args ...interface{}) ([]map[string]interface{}, error) {
	query = db.handleSqlBeforeExec(query)
	return CommonQuery(db.DbList[con], query, args...)
}

// ExecWithConnection implements the method Connection.ExecWithConnection.
func (db *Mssql) ExecWithConnection(con string, query string, args ...interface{}) (sql.Result, error) {
	query = db.handleSqlBeforeExec(query)
	return CommonExec(db.DbList[con], query, args...)
}

// Query implements the method Connection.Query.
func (db *Mssql) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	query = db.handleSqlBeforeExec(query)
	return CommonQuery(db.DbList["default"], query, args...)
}

// Exec implements the method Connection.Exec.
func (db *Mssql) Exec(query string, args ...interface{}) (sql.Result, error) {
	query = db.handleSqlBeforeExec(query)
	return CommonExec(db.DbList["default"], query, args...)
}

// InitDB implements the method Connection.InitDB.
func (db *Mssql) InitDB(cfglist map[string]config.Database) Connection {
	db.Once.Do(func() {
		for conn, cfg := range cfglist {

			if cfg.Dsn == "" {

				cfg.Dsn = fmt.Sprintf("user id=%s;password=%s;server=%s;port=%s;database=%s;encrypt=disable",
					cfg.User, cfg.Pwd, cfg.Host, cfg.Port, cfg.Name)
			}

			sqlDB, err := sql.Open("sqlserver", cfg.Dsn)

			if sqlDB == nil {
				panic("invalid connection")
			}

			if err != nil {
				_ = sqlDB.Close()
				panic(err.Error())
			} else {
				sqlDB.SetMaxIdleConns(cfg.MaxIdleCon)
				sqlDB.SetMaxOpenConns(cfg.MaxOpenCon)

				db.DbList[conn] = sqlDB
			}
		}
	})
	return db
}

// BeginTxWithReadUncommitted starts a transaction with level LevelReadUncommitted.
func (db *Mssql) BeginTxWithReadUncommitted() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelReadUncommitted)
}

// BeginTxWithReadCommitted starts a transaction with level LevelReadCommitted.
func (db *Mssql) BeginTxWithReadCommitted() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelReadCommitted)
}

// BeginTxWithRepeatableRead starts a transaction with level LevelRepeatableRead.
func (db *Mssql) BeginTxWithRepeatableRead() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelRepeatableRead)
}

// BeginTx starts a transaction with level LevelDefault.
func (db *Mssql) BeginTx() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelDefault)
}

// BeginTxWithLevel starts a transaction with given transaction isolation level.
func (db *Mssql) BeginTxWithLevel(level sql.IsolationLevel) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], level)
}

// BeginTxWithReadUncommittedAndConnection starts a transaction with level LevelReadUncommitted and connection.
func (db *Mssql) BeginTxWithReadUncommittedAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelReadUncommitted)
}

// BeginTxWithReadCommittedAndConnection starts a transaction with level LevelReadCommitted and connection.
func (db *Mssql) BeginTxWithReadCommittedAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelReadCommitted)
}

// BeginTxWithRepeatableReadAndConnection starts a transaction with level LevelRepeatableRead and connection.
func (db *Mssql) BeginTxWithRepeatableReadAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelRepeatableRead)
}

// BeginTxAndConnection starts a transaction with level LevelDefault and connection.
func (db *Mssql) BeginTxAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelDefault)
}

// BeginTxWithLevelAndConnection starts a transaction with given transaction isolation level and connection.
func (db *Mssql) BeginTxWithLevelAndConnection(conn string, level sql.IsolationLevel) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], level)
}

// QueryWithTx is query method within the transaction.
func (db *Mssql) QueryWithTx(tx *sql.Tx, query string, args ...interface{}) ([]map[string]interface{}, error) {
	query = db.handleSqlBeforeExec(query)
	return CommonQueryWithTx(tx, query, args...)
}

// ExecWithTx is exec method within the transaction.
func (db *Mssql) ExecWithTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	query = db.handleSqlBeforeExec(query)
	return CommonExecWithTx(tx, query, args...)
}
