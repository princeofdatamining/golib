
package sqlutil

import (
    "fmt"
    "strings"
    "reflect"
    "database/sql"
)

var (
    errfBindGetKeys = errFormatFactory("bindGet: table %q has not any key %q")
    errfBindGetCols = errFormatFactory("bindGet: table %q has not any registered columns")
)

func (this *tableMap) bindGet(key string) (bind *bindObj, err error) {
    if this.getBinds == nil {
        this.getBinds = map[string]*bindObj{}
    }
    if bind, ok := this.getBinds[key]; ok {
        return bind, nil
    }
    bind = &bindObj{}
    dialect := this.dbmap.dialect
    sql := dialect.SelectSQL(this.schemaName, this.tableName)
    var (
        getBinds []string
        wheres []string
        fieldIsKey = map[string]bool{}
        L int
    )
    //
    keyCols := this.getKeyColumns(key)
    L = len(keyCols)
    if L <= 0 {
        return nil, errfBindGetKeys(this.tableName, key)
    }
    wheres = make([]string, L)
    bind.keyFields = make([]string, L)
    for i, col := range keyCols {
        colName, fldName := col.GetColumnName(), col.GetFieldName()
        wheres[i] = fmt.Sprintf("%s = %s", dialect.QuoteField(colName), dialect.BindVar(i))
        fieldIsKey[fldName] = true
        bind.keyFields[i] = fldName
    }
    //
    var v int
    L = len(this.columns)
    getBinds = make([]string, L)
    bind.setFields = make([]string, L)
    for _, col := range this.columns {
        colName, fldName := col.GetColumnName(), col.GetFieldName()
        if fieldIsKey[fldName] {
            continue
        }
        getBinds[v] = dialect.QuoteField(colName)
        bind.setFields[v] = fldName
        v++
    }
    if v <= 0 {
        return nil, errfBindGetCols(this.tableName)
    }
    getBinds = getBinds[:v]
    bind.setFields = bind.setFields[:v]
    bind.argFields = bind.keyFields
    //
    bind.query = fmt.Sprintf(sql, strings.Join(getBinds[:v], ", "), strings.Join(wheres, " AND "), "")
    this.getBinds[key] = bind
    return 
}
func (this *tableMap) get(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value, key string) (rows int64, err error) {
    var (
        bind        *bindObj
    )
    if bind, err = this.bindGet(key); err != nil {
        return 
    }
    data := vptr.Elem()
    if err = bind.bindArgs(data); err != nil {
        return 
    }
    //
    dest := make([]interface{}, len(bind.setFields))
    for i, fldName := range bind.setFields {
        f := data.FieldByName(fldName)
        target := f.Addr().Interface()
        dest[i] = target
    }
    //*
    err = exec.QueryRow(bind.query, bind.argValues...).Scan(dest...)
    if err != nil {
        if err == sql.ErrNoRows {
            err = nil
        }
        return 
    }
    rows++
    //*/
    if err = triggerRun("PostGet", vptr, execVal); err != nil {
        return 
    }
    return 
}
func (this *dbMap) get(exec SQLExecutor, table TableMap, objects []interface{}) (rows int64, err error) {
    var (
        triggerArgs = triggerArg(exec)
        affected int64
        n int
        key string
        useKey bool
    )
    for _, obj := range objects {
        if key, useKey = obj.(string); useKey {
            break
        }
        n++
    }
    if n <= 0 {
        return 
    }
    for i, obj := range objects {
        if i >= n {
            return 
        }
        vptr := reflect.ValueOf(obj)
        if table == nil {
            if table, err = this.getTableByPType(vptr.Type(), "Get"); err != nil {
                return 
            }
        } else {
            if err = table.checkPType(vptr.Type(), "Get"); err != nil {
                return 
            }
        }
        if affected, err = table.get(vptr, exec, triggerArgs, key); err != nil {
            return 
        }
        rows += affected
    }
    return 
}
func (this *dbMap) Get(objects ...interface{}) (rows int64, err error) {
    return this.get(this, nil, objects)
}
func (this *txMap) Get(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.get(this, nil, objects)
}
func (this *tableMap) Get(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.get(this, this, objects)
}
