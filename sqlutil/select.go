
package sqlutil

import (
    "fmt"
    "strings"
    "regexp"
    "errors"
    "reflect"
    "database/sql"
)

var (
    ErrSelectOneGetMore = errors.New("SelectOne: got multiple rows")
    errfSelectIntoNonStructType = errFormatFactory("SELECT into non-struct type: %v")
    errfSelectIntoPointerSlices = errFormatFactory("SELECT into non-pointer slices: %v")
    errfSelectIntoNonStructMoreCols = errFormatFactory("SELECT into non-struct, only 1 column requires: %d")
    errfTableMetaNoField = errFormatFactory("SELECT into struct, column %q not in meta of table %q")
)

func (this *dbMap) columnsToFieldIndexList(meta reflect.Type, cols []string) ([][]int, error) {
    fieldIndexList := make([][]int, len(cols))

    table, ok := this.getTableByMeta(meta)
    if !ok {
        table = this.getOrAddPseudoTable(meta)
    }
    for i, colName := range cols {
        if col, ok := table.colDict[colName]; ok {
            if f, ok := meta.FieldByName(col.fieldName); ok {
                fieldIndexList[i] = f.Index
            }
        }
        if fieldIndexList[i] == nil {
            return nil, errfTableMetaNoField(colName, table.tableName)
        }
    }
    return fieldIndexList, nil
}
func checkSlices(holder interface{}, appendToSlice bool) (reflect.Type, error) {
    t := reflect.TypeOf(holder)
    raw := t
    // if not append to holder slices, holder MUST be valid meta for new slices
    if !appendToSlice {
        for t.Kind() == reflect.Ptr {
            t = t.Elem()
        }
        if t.Kind() != reflect.Struct {
            return nil, errfSelectIntoNonStructType(raw)
        }
        return t, nil
    }
    // else; check holder as slices
    if true {
        if t.Kind() != reflect.Ptr {
            return nil, errfSelectIntoPointerSlices(raw)
        }
        if t = t.Elem(); t.Kind() != reflect.Slice {
            return nil, errfSelectIntoPointerSlices(raw)
        }
        return t.Elem(), nil
    }
    return nil, nil
}
func (this *dbMap) selectIntoOrNew(exec SQLExecutor, holder interface{}, appendToSlice bool, query string, args ...interface{}) (list []interface{}, err error) {
    var (
        meta reflect.Type
        elemIsPointer = true
        elemIsStruct = true
        rows    *sql.Rows
        cols    []string
        colN    int
        colToFieldIndex     [][]int
        holderSlice = reflect.Indirect(reflect.ValueOf(holder))
        newSlice []interface{}
    )
    if meta, err = checkSlices(holder, appendToSlice); err != nil {
        return 
    }
    if appendToSlice {
        if elemIsPointer = meta.Kind() == reflect.Ptr; elemIsPointer {
            meta = meta.Elem()
        }
        elemIsStruct = meta.Kind() == reflect.Struct
    }

    if len(args) == 1 {
        query, args = this.maybeExpandNamedQuery(query, args)
    }

    if rows, err = exec.Query(query, args...); err != nil {
        return 
    }
    defer rows.Close()

    if cols, err = rows.Columns(); err != nil {
        return 
    }
    colN = len(cols)
    if !elemIsStruct && colN > 1 {
        return nil, errfSelectIntoNonStructMoreCols(colN)
    }
    if elemIsStruct {
        if colToFieldIndex, err = this.columnsToFieldIndexList(meta, cols); err != nil {
            return 
        }
    }

    dest := make([]interface{}, colN)
    for {
        if !rows.Next() {
            if err = rows.Err(); err != nil {
                return 
            }
            break
        }

        pvElem := reflect.New(meta)
        vElem := pvElem.Elem()
        for i := range cols {
            target := vElem
            if elemIsStruct {
                target = target.FieldByIndex(colToFieldIndex[i])
            }
            dest[i] = target.Addr().Interface()
        }

        if err = rows.Scan(dest...); err != nil {
            return 
        }

        if appendToSlice {
            if elemIsPointer {
                vElem = pvElem
            }
            holderSlice.Set(reflect.Append(holderSlice, vElem))
        } else {
            newSlice = append(newSlice, pvElem.Interface())
        }
    }

    if appendToSlice {
        //
    } else {
        list = newSlice
    }
    return 
}
func (this *dbMap) selectAll(exec SQLExecutor, slices interface{}, query string, args ...interface{}) (rows int64, err error) {
    if _, err = this.selectIntoOrNew(exec, slices, true, query, args...); err != nil {
        return 
    }
    triggerArgs := triggerArg(exec)
    vslices := reflect.Indirect(reflect.ValueOf(slices))
    rows = int64(vslices.Len())
    for i := 0; i < int(rows); i++ {
        if err = triggerRun("PostGet", vslices.Index(i), triggerArgs); err != nil {
            return 
        }
    }
    return 
}
func (this *dbMap   ) SelectAll(slices interface{}, query string, args ...interface{}) (int64, error) { return this      .selectAll(this, slices, query, args...) }
func (this *txMap   ) SelectAll(slices interface{}, query string, args ...interface{}) (int64, error) { return this.dbmap.selectAll(this, slices, query, args...) }
func (this *tableMap) SelectAll(slices interface{}, query string, args ...interface{}) (int64, error) { return this.dbmap.selectAll(this, slices, query, args...) }

func (this *dbMap) selectOne(exec SQLExecutor, holder interface{}, query string, args ...interface{}) (err error) {
    v := reflect.Indirect(reflect.ValueOf(holder))
    if v.Kind() != reflect.Struct {
        return this.selectVal(exec, holder, query, args...)
    }

    list, err := this.selectIntoOrNew(exec, holder, false, query, args...)
    if err != nil {
        return err
    }
    if list == nil || len(list) <= 0 {
        return sql.ErrNoRows
    }
    if len(list) > 1 {
        err = ErrSelectOneGetMore
    }
    w := reflect.Indirect(reflect.ValueOf(list[0]))
    v.Set(w)
    return 
}
func (this *dbMap   ) SelectOne(holder interface{}, query string, args ...interface{}) (error) { return this      .selectOne(this, holder, query, args...) }
func (this *txMap   ) SelectOne(holder interface{}, query string, args ...interface{}) (error) { return this.dbmap.selectOne(this, holder, query, args...) }
func (this *tableMap) SelectOne(holder interface{}, query string, args ...interface{}) (error) { return this.dbmap.selectOne(this, holder, query, args...) }

//

var (
    reField = regexp.MustCompile("^\\s*(\\w+)(?:\\s+(\\w+)\\s+(.+)\\s*)?\\s*$")
)

func (this *tableMap) makeSelectSQL(fields, where, suffix string) (string) {
    sql := this.dbmap.dialect.SelectSQL(this.schemaName, this.tableName)

    fs := strings.TrimSpace(fields)
    /*
    fs := ""
    for _, field := range strings.Split(fields, ",") {
        if fs != "" {
            fs += ", "
        }
        subs := reField.FindStringSubmatch(field)
        if subs == nil {
            fs += field
        } else {
            fs += this.dbmap.dialect.QuoteField(subs[1])
            if subs[2] != "" {
                fs += fmt.Sprintf(" %s %s", strings.ToUpper(subs[2]), subs[3])
            }
        }
    }
    //*/
    if fs == "" {
        fs = "*"
    }

    if where == "" {
        where = "1"
    }
    return fmt.Sprintf(sql, "", fs, "", where, suffix)
}

func (this *tableMap) SelectOne2 (holder interface{},         where         string, args ...interface{}) (error) {
    return this.SelectOne(holder, this.makeSelectSQL(""    , where, ""    ), args...)
}
func (this *tableMap) SelectOne2x(holder interface{},         where, suffix string, args ...interface{}) (error) {
    return this.SelectOne(holder, this.makeSelectSQL(""    , where, suffix), args...)
}
func (this *tableMap) SelectOne3 (holder interface{}, fields, where         string, args ...interface{}) (error) {
    return this.SelectOne(holder, this.makeSelectSQL(fields, where, ""    ), args...)
}
func (this *tableMap) SelectOne3x(holder interface{}, fields, where, suffix string, args ...interface{}) (error) {
    return this.SelectOne(holder, this.makeSelectSQL(fields, where, suffix), args...)
}


func (this *tableMap) SelectAll2 (slices interface{},         where         string, args ...interface{}) (int64, error) {
    return this.SelectAll(slices, this.makeSelectSQL(""    , where, ""    ), args...)
}
func (this *tableMap) SelectAll2x(slices interface{},         where, suffix string, args ...interface{}) (int64, error) {
    return this.SelectAll(slices, this.makeSelectSQL(""    , where, suffix), args...)
}
func (this *tableMap) SelectAll3 (slices interface{}, fields, where         string, args ...interface{}) (int64, error) {
    return this.SelectAll(slices, this.makeSelectSQL(fields, where, ""    ), args...)
}
func (this *tableMap) SelectAll3x(slices interface{}, fields, where, suffix string, args ...interface{}) (int64, error) {
    return this.SelectAll(slices, this.makeSelectSQL(fields, where, suffix), args...)
}
