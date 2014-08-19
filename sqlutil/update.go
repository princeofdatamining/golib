
package sqlutil

import (
    "fmt"
    "strings"
    "reflect"
    "database/sql"
)

var (
    errfBindUpdateKeys = errFormatFactory("bindUpdate: table %q has not any primary|unique keys")
    errfBindUpdateCols = errFormatFactory("bindUpdate: table %q has not any registered columns")
)

func (this *tableMap) bindUpdate() (bind *bindObj, err error) {
    if bind = this.updBind; bind != nil {
        return 
    }
    bind = &bindObj{}
    dialect := this.dbmap.dialect
    sql := dialect.UpdateSQL(this.schemaName, this.tableName)
    var (
        updateBinds []string
        whereKeys []string
        fieldIsKey = map[string]bool{}
    )
    //
    keys := this.primaries
    if len(keys) <= 0 {
        keys = this.uniques
    }
    for _, cols := range keys {
        L := len(cols)
        if L <= 0 {
            continue
        }
        bind.keyFields = make([]string, L)
        whereKeys = make([]string, L)
        for i, col := range cols {
            colName, fldName := col.GetColumnName(), col.GetFieldName()
            whereKeys[i] = dialect.QuoteField(colName)
            fieldIsKey[fldName] = true
            bind.keyFields[i] = fldName
        }
        break
    }
    if whereKeys == nil {
        return nil, errfBindUpdateKeys(this.tableName)
    }
    //
    var v int
    L := len(this.columns)
    updateBinds = make([]string, L)
    bind.argFields = make([]string, L)
    for _, col := range this.columns {
        colName, fldName := col.GetColumnName(), col.GetFieldName()
        if col == this.autoincrCol || fieldIsKey[fldName] {
            continue
        }
        updateBinds[v] = fmt.Sprintf("%s = %s", dialect.QuoteField(colName), dialect.BindVar(v))
        bind.argFields[v] = fldName
        v++
    }
    if v <= 0 {
        return nil, errfBindUpdateCols(this.tableName)
    }
    //
    n := copy(bind.argFields[v:], bind.keyFields)
    bind.argFields = bind.argFields[:v+n]
    for i, whereKey := range whereKeys {
        whereKeys[i] = fmt.Sprintf("%s = %s", whereKey, dialect.BindVar(v+i))
    }
    bind.query = fmt.Sprintf(sql, strings.Join(updateBinds[:v], ", "), strings.Join(whereKeys, " AND "))
    this.updBind = bind
    return 
}
func (this *tableMap) update(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value) (rows int64, err error) {
    var (
        bind        *bindObj
        res         sql.Result
    )
    if err = triggerRun("PreUpdate", vptr, execVal); err != nil {
        return 
    }
    if bind, err = this.bindUpdate(); err != nil {
        return 
    }
    if err = bind.bindArgs(vptr.Elem()); err != nil {
        return 
    }
    if res, err = exec.Exec(bind.query, bind.argValues...); err != nil {
        return 
    }
    if rows, err = res.RowsAffected(); err != nil {
        return 
    }
    if err = triggerRun("PostUpdate", vptr, execVal); err != nil {
        return 
    }
    return 
}
func (this *dbMap) update(exec SQLExecutor, table TableMap, objects []interface{}) (rows int64, err error) {
    var (
        triggerArgs = triggerArg(exec)
        affected int64
    )
    for _, obj := range objects {
        vptr := reflect.ValueOf(obj)
        if table == nil {
            if table, err = this.getTableByPType(vptr.Type(), "Update"); err != nil {
                return 
            }
        } else {
            if err = table.checkPType(vptr.Type(), "Update"); err != nil {
                return 
            }
        }
        if affected, err = table.update(vptr, exec, triggerArgs); err != nil {
            return 
        }
        rows += affected
    }
    return 
}
func (this *dbMap) Update(objects ...interface{}) (rows int64, err error) {
    return this.update(this, nil, objects)
}
func (this *txMap) Update(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.update(this, nil, objects)
}
func (this *tableMap) Update(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.update(this, this, objects)
}

func setWhere(where string) (ret string) {
    ret = strings.TrimSpace(where)
    switch ret {
    case "", "all", "ALL", "*":
        return "1"
    }
    return ret
}
func (this *tableMap) Update2(exec SQLExecutor, where string, data map[string]string, except []string) (rows int64, err error) {
    dialect := this.dbmap.dialect
    updSQL := dialect.UpdateSQL(this.schemaName, this.tableName)
    //
    L := len(data)
    setFields := make([]string, L)
    quote := ""
    noQuotes := array2dict(except)
    var i int
    for key, val := range data {
        if _, ok := noQuotes[key]; ok {
            quote = ""
        } else {
            quote = "'"
        }
        setFields[i] = fmt.Sprintf("%s = %s%s%s", dialect.QuoteField(key), quote, val, quote)
        i++
    }
    //
    query := fmt.Sprintf(updSQL, strings.Join(setFields, ", "), setWhere(where))
    if exec == nil {
        exec = this
    }
    var res sql.Result
    if res, err = exec.Exec(query); err == nil {
        rows, err = res.RowsAffected()
    }
    return 
}
