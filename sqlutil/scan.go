
package sqlutil

import (
    "database/sql"
    "database/sql/driver"
    "reflect"
    "strings"
    "errors"
)

func buildValues(columns []string, values []interface{}, data []interface{}, dummy interface{}) (bool) {
    n := len(data)
    var v reflect.Value
    var instruct bool

    if n == 1 {
        v = reflect.Indirect(reflect.ValueOf(data[0]))
        instruct = v.Kind() == reflect.Struct
    }
    if instruct {
        for i, cname := range columns {
            f := v.FieldByNameFunc(func (s string) (bool) {
                return strings.ToLower(s) == strings.ToLower(cname)
            })
            if f.IsValid() && f.CanSet() {
                values[i] = f.Addr().Interface()
            } else {
                values[i] = dummy
            }
        }
        return true
    }

    if true {
        for i := range values {
            if i < n {
                values[i] = data[i]
            } else {
                values[i] = dummy
            }
        }
        return true
    }

    return false
}

//*/

type StmtBind interface {
    Close() (error)
    Exec (args ...interface{}) (sql.Result, error)
    Query(args ...interface{}) (error)
    Next() (error)
    One () (error)
}

func NewStmtBind(stmt *sql.Stmt, data ...interface{}) (StmtBind) {
    return &stmtBind{
        stmt:   stmt,
        data:   data,
    }
}
func NewStmt(db *sql.DB, query string, data ...interface{}) (StmtBind, error) {
    stmt, err := db.Prepare(query)
    if err != nil {
        return nil, err
    }
    bind := NewStmtBind(stmt, data...)
    return bind, nil
}
type stmtBind struct {
    stmt    *sql.Stmt
    data    []interface{}
    //
    err     error
    rows    *sql.Rows
    //
    columns []string
    values  []interface{}
    dummy   interface{}
}
func (this *stmtBind) Close() (error) {
    stmt := this.stmt
    this.stmt = nil
    return stmt.Close()
}
func (this *stmtBind) clear() () {
    this.err = nil
    if this.rows != nil {
        this.rows.Close()
        this.rows = nil
    }
}
func (this *stmtBind) Exec(args ...interface{}) (res sql.Result, err error) {
    this.clear()
    // Like sql.DB.Query() try more times
    for i := 0; i < 10; i++ {
        res, err = this.stmt.Exec(args...)
        if err != driver.ErrBadConn {
            break
        }
    }
    return 
}
func (this *stmtBind) Query(args ...interface{}) (err error) {
    this.clear()
    var done bool
    defer func() {
        if !done {
            this.clear()
        }
    }()
    // Like sql.DB.Query() try more times
    for i := 0; i < 10; i++ {
        this.rows, err = this.stmt.Query(args...)
        if err != driver.ErrBadConn {
            break 
        }
    }
    if err != nil {
        return 
    }
    //
    if this.columns != nil {
        done = true
        return
    }
    //
    var columns []string
    columns, err = this.rows.Columns()
    if err != nil {
        return 
    }
    this.columns = columns
    this.values = make([]interface{}, len(columns))
    if !buildValues(this.columns, this.values, this.data, &this.dummy) {
        return errors.New("sqlutil: can't build values for scan")
    }
    //
    done = true
    return 
}
func (this *stmtBind) scan() (error) {
    return this.rows.Scan(this.values...)
}
func (this *stmtBind) fetch(close bool) (error) {
    if this.err != nil {
        return this.err
    }
    if close {
        defer this.clear()
    }
    if this.rows == nil || !this.rows.Next() {
        return sql.ErrNoRows
    }
    return this.scan()
}
func (this *stmtBind) Next() (error) {
    return this.fetch(false)
}
func (this *stmtBind) One() (error) {
    return this.fetch( true)
}
