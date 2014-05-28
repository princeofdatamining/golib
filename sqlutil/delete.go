
package sqlutil

import (
    "fmt"
    "strings"
    "reflect"
    "database/sql"
)

var (
    errfBindDelete = errFormatFactory("bindDelete: table %q has not any primary|unique keys")
)

func (this *tableMap) bindDelete() (bind *bindObj, err error) {
    if bind = this.delBind; bind != nil {
        return 
    }
    bind = &bindObj{}
    dialect := this.dbmap.dialect
    sql := dialect.DeleteSQL(this.schemaName, this.tableName)
    var (
        wheres []string
    )
    //
    keys := this.primaries
    if len(keys) <= 0 {
        keys = this.uniques
    }
    //
    for _, cols := range keys {
        L := len(cols)
        if L <= 0 {
            continue
        }
        bind.keyFields = make([]string, L)
        wheres = make([]string, L)
        for i, col := range cols {
            colName, fldName := col.GetColumnName(), col.GetFieldName()
            wheres[i] = fmt.Sprintf("%s = %s", dialect.QuoteField(colName), dialect.BindVar(i))
            bind.keyFields[i] = fldName
        }
        break
    }
    if wheres == nil {
        return nil, errfBindDelete(this.tableName)
    }
    bind.argFields = bind.keyFields
    bind.query = fmt.Sprintf(sql, strings.Join(wheres, " AND "))
    this.updBind = bind
    return 
}
func (this *tableMap) delete(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value) (rows int64, err error) {
    var (
        bind        *bindObj
        res         sql.Result
    )
    if err = triggerRun("PreDelete", vptr, execVal); err != nil {
        return 
    }
    if bind, err = this.bindDelete(); err != nil {
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
    if err = triggerRun("PostDelete", vptr, execVal); err != nil {
        return 
    }
    return 
}
func (this *dbMap) delete(exec SQLExecutor, table TableMap, objects []interface{}) (rows int64, err error) {
    var (
        triggerArgs = triggerArg(exec)
        affected int64
    )
    for _, obj := range objects {
        vptr := reflect.ValueOf(obj)
        if table == nil {
            if table, err = this.getTableByPType(vptr.Type(), "Delete"); err != nil {
                return 
            }
        } else {
            if err = table.checkPType(vptr.Type(), "Delete"); err != nil {
                return 
            }
        }
        if affected, err = table.delete(vptr, exec, triggerArgs); err != nil {
            return 
        }
        rows += affected
    }
    return 
}
func (this *dbMap) Delete(objects ...interface{}) (rows int64, err error) {
    return this.delete(this, nil, objects)
}
func (this *txMap) Delete(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.delete(this, nil, objects)
}
func (this *tableMap) Delete(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.delete(this, this, objects)
}
