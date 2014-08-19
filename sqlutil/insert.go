
package sqlutil

import (
    "fmt"
    "strings"
    "reflect"
)

var (
    errfBindInsert = errFormatFactory("bindInsert: table %q has not any registered columns")
)

func (this *tableMap) bindInsert() (bind *bindObj, err error) {
    if bind = this.insBind; bind != nil {
        return 
    }
    bind = &bindObj{}
    L := len(this.columns)
    if L <= 0 {
        return nil, errfBindInsert(this.tableName)
    }
    dialect := this.dbmap.dialect
    sql := dialect.InsertSQL(this.schemaName, this.tableName, this.autoincrCol)
    //
    colNames := make([]string, L)
    bindVars := make([]string, L)
    bind.argFields = make([]string, L)
    var v int
    for i, col := range this.columns {
        colName, fldName := col.GetColumnName(), col.GetFieldName()
        colNames[i] = dialect.QuoteField(colName)
        if col == this.autoincrCol {
            bindVars[i] = dialect.BindAutoIncrVar()
        } else {
            bindVars[i] = dialect.BindVar(v)
            bind.argFields[v] = fldName
            v++
        }
    }
    bind.argFields = bind.argFields[:v]
    bind.query = fmt.Sprintf(sql, strings.Join(colNames, ", "), strings.Join(bindVars, ", "))
    this.updBind = bind
    return 
}
func (this *tableMap) insert(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value) (err error) {
    var (
        bind        *bindObj
        id          int64
    )
    if err = triggerRun("PreInsert", vptr, execVal); err != nil {
        return 
    }
    if bind, err = this.bindInsert(); err != nil {
        return 
    }
    if err = bind.bindArgs(vptr.Elem()); err != nil {
        return 
    }
    if this.autoincrCol == nil {
        if _, err = exec.Exec(bind.query, bind.argValues...); err != nil {
            return 
        }
    } else {
        if id, err = this.dbmap.dialect.InsertAndReturnId(exec, bind.query, bind.argValues...); err != nil {
            return 
        }
        f := vptr.Elem().FieldByName(this.autoincrCol.GetFieldName())
        f.SetInt(id)
    }
    if err = triggerRun("PostInsert", vptr, execVal); err != nil {
        return 
    }
    return 
}
func (this *dbMap) insert(exec SQLExecutor, table TableMap, objects []interface{}) (rows int64, err error) {
    var (
        triggerArgs = triggerArg(exec)
    )
    for _, obj := range objects {
        vptr := reflect.ValueOf(obj)
        if table == nil {
            if table, err = this.getTableByPType(vptr.Type(), "Insert"); err != nil {
                return 
            }
        } else {
            if err = table.checkPType(vptr.Type(), "Insert"); err != nil {
                return 
            }
        }
        if err = table.insert(vptr, exec, triggerArgs); err != nil {
            return 
        }
        rows++
    }
    return 
}
func (this *dbMap) Insert(objects ...interface{}) (rows int64, err error) {
    return this.insert(this, nil, objects)
}
func (this *txMap) Insert(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.insert(this, nil, objects)
}
func (this *tableMap) Insert(objects ...interface{}) (rows int64, err error) {
    return this.dbmap.insert(this, this, objects)
}

func array2dict(arr []string) (dict map[string]bool) {
    dict = make(map[string]bool)
    for _, s := range arr {
        dict[s] = true
    }
    return 
}
func (this *tableMap) Insert2(exec SQLExecutor, data map[string]string, except []string) (id int64, err error) {
    dialect := this.dbmap.dialect
    sql := dialect.InsertSQL(this.schemaName, this.tableName, this.autoincrCol)
    //
    L := len(data)
    colNames := make([]string, L)
    colBinds := make([]string, L)
    quote := ""
    noQuotes := array2dict(except)
    var i int
    for key, val := range data {
        if _, ok := noQuotes[key]; ok {
            quote = ""
        } else {
            quote = "'"
        }
        colNames[i] = dialect.QuoteField(key)
        colBinds[i] = fmt.Sprintf("%s%s%s", quote, val, quote)
        i++
    }
    query := fmt.Sprintf(sql, strings.Join(colNames, ", "), strings.Join(colBinds, ", "))
    if exec == nil {
        exec = this
    }
    return dialect.InsertAndReturnId(exec, query)
}
