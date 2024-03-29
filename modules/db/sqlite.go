// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"github.com/glvd/go-admin/modules/config"
)

// Sqlite is a Connection of sqlite.
type Sqlite struct {
	Base
}

// GetSqliteDB return the global sqlite connection.
func GetSqliteDB() *Sqlite {
	return &Sqlite{
		Base: Base{
			DbList: make(map[string]*sql.DB),
		},
	}
}

// Name implements the method Connection.Name.
func (db *Sqlite) Name() string {
	return "sqlite"
}

// GetDelimiter implements the method Connection.GetDelimiter.
func (db *Sqlite) GetDelimiter() string {
	return "`"
}

// QueryWithConnection implements the method Connection.QueryWithConnection.
func (db *Sqlite) QueryWithConnection(con string, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return CommonQuery(db.DbList[con], query, args...)
}

// ExecWithConnection implements the method Connection.ExecWithConnection.
func (db *Sqlite) ExecWithConnection(con string, query string, args ...interface{}) (sql.Result, error) {
	return CommonExec(db.DbList[con], query, args...)
}

// Query implements the method Connection.Query.
func (db *Sqlite) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return CommonQuery(db.DbList["default"], query, args...)
}

// Exec implements the method Connection.Exec.
func (db *Sqlite) Exec(query string, args ...interface{}) (sql.Result, error) {
	return CommonExec(db.DbList["default"], query, args...)
}

// InitDB implements the method Connection.InitDB.
func (db *Sqlite) InitDB(cfgList map[string]config.Database) Connection {
	db.Once.Do(func() {
		for conn, cfg := range cfgList {
			sqlDB, err := sql.Open("sqlite3", cfg.File)

			if err != nil {
				panic(err)
			} else {
				db.DbList[conn] = sqlDB
			}
		}
	})
	return db
}

// BeginTxWithReadUncommitted starts a transaction with level LevelReadUncommitted.
func (db *Sqlite) BeginTxWithReadUncommitted() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelReadUncommitted)
}

// BeginTxWithReadCommitted starts a transaction with level LevelReadCommitted.
func (db *Sqlite) BeginTxWithReadCommitted() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelReadCommitted)
}

// BeginTxWithRepeatableRead starts a transaction with level LevelRepeatableRead.
func (db *Sqlite) BeginTxWithRepeatableRead() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelRepeatableRead)
}

// BeginTx starts a transaction with level LevelDefault.
func (db *Sqlite) BeginTx() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelDefault)
}

// BeginTxWithLevel starts a transaction with given transaction isolation level.
func (db *Sqlite) BeginTxWithLevel(level sql.IsolationLevel) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], level)
}

// BeginTxWithReadUncommittedAndConnection starts a transaction with level LevelReadUncommitted and connection.
func (db *Sqlite) BeginTxWithReadUncommittedAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelReadUncommitted)
}

// BeginTxWithReadCommittedAndConnection starts a transaction with level LevelReadCommitted and connection.
func (db *Sqlite) BeginTxWithReadCommittedAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelReadCommitted)
}

// BeginTxWithRepeatableReadAndConnection starts a transaction with level LevelRepeatableRead and connection.
func (db *Sqlite) BeginTxWithRepeatableReadAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelRepeatableRead)
}

// BeginTxAndConnection starts a transaction with level LevelDefault and connection.
func (db *Sqlite) BeginTxAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelDefault)
}

// BeginTxWithLevelAndConnection starts a transaction with given transaction isolation level and connection.
func (db *Sqlite) BeginTxWithLevelAndConnection(conn string, level sql.IsolationLevel) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], level)
}

// QueryWithTx is query method within the transaction.
func (db *Sqlite) QueryWithTx(tx *sql.Tx, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return CommonQueryWithTx(tx, query, args...)
}

// ExecWithTx is exec method within the transaction.
func (db *Sqlite) ExecWithTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	return CommonExecWithTx(tx, query, args...)
}
