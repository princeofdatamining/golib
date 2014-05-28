
package sqlutil

import (
    "fmt"
    "database/sql"
)

type Transaction interface {
    SQLExecutor
    //
    Commit() (error)
    Rollback() (error)
}

func (this *dbMap) Begin() (Transaction, error) {
    tx, err := this.db.Begin()
    if err != nil {
        return nil, err
    }
    return &txMap{
        dbmap: this,
        tx: tx,
    }, nil
}

type txMap struct {
    dbmap   *dbMap
    tx      *sql.Tx
    closed  bool
}
func (this *txMap) Commit() (error) {
    if !this.closed {
        this.closed = true
        return this.tx.Commit()
    }
    return sql.ErrTxDone
}
func (this *txMap) Rollback() (error) {
    if !this.closed {
        this.closed = true
        return this.tx.Rollback()
    }
    return sql.ErrTxDone
}

func (this *txMap) Exec(query string, args ...interface{}) (sql.Result, error) {
    fmt.Println("Tx.Exec:", query)
    return this.tx.Exec(query, args...)
}
func (this *txMap) exec(query string, args ...interface{}) (err error) { _, err = this.Exec(query, args...); return }
func (this *txMap) Query(query string, args ...interface{}) (*sql.Rows, error) {
    fmt.Println("Tx.Query:", query)
    return this.tx.Query(query, args...)
}
func (this *txMap) QueryRow(query string, args ...interface{}) (*sql.Row) {
    fmt.Println("Tx.QueryRow:", query)
    return this.tx.QueryRow(query, args...)
}
