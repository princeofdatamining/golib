
package sqlutil

import (
    "reflect"
    "strings"
    "strconv"
    "github.com/princeofdatamining/golib/sqlutil/dialect"
)

func (this *tableMap) buildColumns(t reflect.Type) () {
    for i,n :=0,t.NumField(); i<n; i++ {
        f := t.Field(i)
        if f.Anonymous && f.Type.Kind() == reflect.Struct {
            this.buildColumns(f.Type)
        } else {
            this.newColumn(f)
        }
    }
}
func (this *tableMap) newColumn(f reflect.StructField) () {
    col := &columnMap{
        columnName: strings.ToLower(f.Name),
        gotype: f.Type,
        fieldName: f.Name,
        precision: -1,
    }
    fields := strings.Split(f.Tag.Get("db"), ",")
    var (
        sprimary, sunique, sindex string
        bprimary, bunique, bindex bool
        named bool
        err error
    )
    for _, field := range fields {
        parts := strings.SplitN(field, "=", 2)
        if len(parts) < 2 {
            switch field {
            case "":
                //
            case "autoincr":
                col.autoincr = true
            case "primary":
                col.primary = true
            case "unique":
                col.unique = true
            case "index":
                col.index = true
            case "notnull":
                col.notnull =  true
            default:
                if named {
                    col.comment = field
                } else {
                    named = true
                    col.setname(field)
                }
            }
            continue
        }
        switch parts[0] {
        case "name":
            col.setname(parts[1])
        case "comment":
            col.comment = parts[1]
        case "primary":
            sprimary, bprimary = parts[1], true
        case "unique":
            sunique , bunique  = parts[1], true
        case "index":
            sindex  , bindex   = parts[1], true
        case "type":
            col.newtype = parts[1]
        case "size":
            if col.maxsize, err = strconv.Atoi(parts[1]); err != nil {
                col.maxsize = -1
            }
        case "precision":
            if col.precision, err = strconv.Atoi(parts[1]); err != nil {
                col.precision = -1
            }
        case "default":
            col.hasDefault = true
            col.defaults = parts[1]
        }
    }
    _, ok := this.colDict[col.columnName]
    if col.transient || ok {
        return 
    }
    this.columns = append(this.columns, col)
    this.colDict[col.columnName] = col
    if this.pseudo {
        return 
    }
    if col.autoincr {
        this.autoincrCol = col
    }
    // NOT sure autoincr represent primary for all database
    if !bprimary && (col.primary || col.autoincr) {
        bprimary, sprimary = true, col.columnName
    }
    if !bunique  && col.unique  {
        bunique , sunique  = true, col.columnName
    }
    if !bindex   && col.index   {
        bindex  , sindex   = true, col.columnName
    }
    this.addCombinedKey(this.primaries, bprimary, sprimary, col)
    this.addCombinedKey(this.uniques  , bunique , sunique , col)
    this.addCombinedKey(this.indexes  , bindex  , sindex  , col)
}
func (this *tableMap) addCombinedKey(list map[string][]dialect.ColumnMeta, work bool, key string, col *columnMap) {
    if !work {
        return 
    }
    list[key] = append(list[key], col)
}

func (this *tableMap) getKeyColumns(key string) (cols []dialect.ColumnMeta) {
    var ok bool
    if cols, ok = this.primaries[key]; ok {
        return 
    }
    if cols, ok = this.uniques[key]; ok {
        return 
    }
    if cols, ok = this.indexes[key]; ok {
        return 
    }
    if col, ok := this.colDict[key]; ok {
        return []dialect.ColumnMeta{col}
    }
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
        return cols
    }
    return nil
}

type ColumnMap interface {
}
type columnMap struct {
    fieldName   string
    gotype      reflect.Type

    columnName  string
    transient   bool
    notnull     bool
    autoincr    bool
    primary     bool
    unique      bool
    index       bool

    newtype     string
    maxsize     int
    precision   int
    hasDefault  bool
    defaults    string
    comment     string
}
func (this *columnMap) GetColumnName() (string) { return this.columnName }
func (this *columnMap) GetFieldName() (string) { return this.fieldName }
func (this *columnMap) GetNotNull() (bool) { return this.notnull }
func (this *columnMap) GetAutoIncr() (bool) { return this.autoincr }
func (this *columnMap) GetGoType() (reflect.Type) { return this.gotype }
func (this *columnMap) GetForceType() (string) { return this.newtype }
func (this *columnMap) GetSize() (int, int) { return this.maxsize, this.precision }
func (this *columnMap) GetDefault() (bool, string) { return this.hasDefault, this.defaults }
func (this *columnMap) GetComment() (string) { return this.comment }

func (this *columnMap) setname(colname string) (ColumnMap) {
    this.columnName = colname
    this.transient = colname == "-" || colname == ""
    return this
}
