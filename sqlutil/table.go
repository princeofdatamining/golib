
package sqlutil

import (
    "reflect"
    "errors"
    "fmt"
    "strings"
    "database/sql"
    "github.com/princeofdatamining/golib/sqlutil/dialect"
)

var (
    errfTableNotFound = errFormatFactory("table %q was not registered.")
    errfTableHasExists = errFormatFactory("table %q has been registered.")
    errMetaNotFound = errors.New("no table registered with this struct")
    errfOpMustWithPointer = errFormatFactory("%s(object...): object must be pointer")
    errfOpInvalidMeta = errFormatFactory("%s(object...): object must be %q type, but got %q")
)

func (this *dbMap) getOrAddPseudoTable(meta reflect.Type) (tm *tableMap) {
    for _, tm = range this.pseudos {
        if tm.gotype == meta {
            return tm
        }
    }
    tm = &tableMap{
        dbmap: this,
        pseudo: true,
        gotype: meta,
        colDict: make(map[string]*columnMap),
    }
    tm.buildColumns(meta)
    this.pseudos = append(this.pseudos, tm)
    return tm
}
func (this *dbMap) AddTable3(meta interface{}, schema, name, comment string) (TableMap, error) {
    t := reflect.TypeOf(meta)
    if name == "" {
        name = t.Name()
    }
    //*
    if _, ok := this.tableD[name]; ok {
        return nil, errfTableHasExists(name)
    }
    for _, table := range this.tables {
        if table.gotype == t {
            table.tableName = name
            return table, nil
        }
    }
    //*/
    tmap := &tableMap{
        dbmap: this,
        schemaName: schema,
        tableName: name,
        gotype: t,
        comment: comment,
        colDict: make(map[string]*columnMap),
        primaries:   make(map[string][]dialect.ColumnMeta),
        uniques  :   make(map[string][]dialect.ColumnMeta),
        indexes  :   make(map[string][]dialect.ColumnMeta),
    }
    tmap.buildColumns(t)

    this.tables = append(this.tables, tmap)
    this.tableD[name] = tmap
    return tmap, nil
}
func (this *dbMap) AddTable2(meta interface{}, name, comment string) (TableMap, error) { return this.AddTable3(meta, "", name, comment) }
func (this *dbMap) AddTable (meta interface{}, name          string) (TableMap, error) { return this.AddTable2(meta, "", name) }
func (this *dbMap) AddTable0(meta interface{}                      ) (TableMap, error) { return this.AddTable (meta, "") }
func (this *dbMap) enumTables(f func (TableMap) (string), args []interface{}) (sql string, err error) {
    onlysql := len(args) > 0
    lines := make([]string, len(this.tables))
    for i, table := range this.tables {
        query := f(table)
        lines[i] = query
        if onlysql || err != nil {
            continue
        }
        _, err = table.Exec(query)
    }
    sql = strings.Join(lines, "\n\n")
    return  
}
func (this *dbMap) CreateTables(ifNotExists bool, args ...interface{}) (sql string, err error) {
    return this.enumTables( func (table TableMap) (string) {
        return table.CreateSQL(ifNotExists)
    }, args )
}
func (this *dbMap) TruncateTables(args ...interface{}) (sql string, err error) {
    return this.enumTables( func (table TableMap) (string) {
        return table.TruncateSQL()
    }, args )
}
func (this *dbMap) DropTables  (ifExists    bool, args ...interface{}) (sql string, err error) {
    return this.enumTables( func (table TableMap) (string) {
        return table.DropSQL(ifExists)
    }, args )
}
func (this *dbMap) DropTableByName(t string, ifExists bool) (error) {
    if table, find := this.tableD[t]; !find {
        return errfTableNotFound(t)
    } else {
        return table.Drop(ifExists)
    }
}
func (this *dbMap) DropTableByMeta(meta interface{}, ifExists bool) (error) {
    if table, find := this.GetTableByMeta(meta); !find {
        return errMetaNotFound
    } else {
        return table.Drop(ifExists)
    }
}
func (this *dbMap) GetTableByName(t string) (table TableMap, find bool) {
    table, find = this.tableD[t]
    return 
}
func (this *dbMap) getTableByMeta(t reflect.Type) (table *tableMap, find bool) {
    for _, tbl := range this.tables {
        if tbl.gotype == t {
            return tbl, true
        }
    }
    return nil, false
}
func (this *dbMap) GetTableByMeta(meta interface{}) (table TableMap, find bool) {
    return this.getTableByMeta( reflect.Indirect(reflect.ValueOf(meta)).Type() )
}
func (this *dbMap) getTableByPType(pt reflect.Type, hint string) (table TableMap, err error) {
    if pt.Kind() != reflect.Ptr {
        return nil, errfOpMustWithPointer(hint)
    }
    if table, ok := this.getTableByMeta(pt.Elem()); ok {
        return table, nil
    }
    return nil, errMetaNotFound
}

type TableMap interface {
    SQLExecutor
    //
    CreateSQL(ifNotExists bool) (string)
    Create(ifNotExists bool) (error)
    DropSQL(ifExists bool) (string)
    Drop(ifExists bool) (error)
    TruncateSQL() (string)
    Truncate() (error)

    checkPType(pt reflect.Type, hint string) (err error)
    insert(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value) (err error)
    update(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value) (rows int64, err error)
    delete(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value) (rows int64, err error)
    get(vptr reflect.Value, exec SQLExecutor, execVal []reflect.Value, key string) (rows int64, err error)
}
type tableMap struct {
    dbmap       *dbMap
    schemaName  string
    tableName   string
    gotype      reflect.Type
    comment     string
    pseudo      bool

    columns     []*columnMap
    colDict     map[string]*columnMap
    autoincrCol dialect.ColumnMeta
    primaries   map[string][]dialect.ColumnMeta
    uniques     map[string][]dialect.ColumnMeta
    indexes     map[string][]dialect.ColumnMeta

    delBind     *bindObj
    insBind     *bindObj
    updBind     *bindObj
    getBinds    map[string]*bindObj
}
func (this *tableMap) checkPType(pt reflect.Type, hint string) (err error) {
    if pt.Kind() != reflect.Ptr {
        err = errfOpMustWithPointer(hint)
    } else if et := pt.Elem(); et != this.gotype {
        err = errfOpInvalidMeta(hint, this.gotype.Name(), et.Name())
    }
    return 
}

func (this *tableMap) CreateSQL(ifNotExists bool) (string) {
    f := this.dbmap.dialect.CreateTableSQL(this.schemaName, this.tableName, ifNotExists, map[string]string{
        "comment": this.comment,
    })
    var lines, primaries, uniques, indexes []string
    for _, col := range this.columns {
        if col.transient {
            continue
        }
        lines = append(lines, this.dbmap.dialect.CreateColumnStr(col))
    }
    //*
    for key, list := range this.primaries {
        if text := this.dbmap.dialect.CreatePrimaryKey(key, list...); text != "" { primaries = append(primaries, text) }
    }
    for key, list := range this.uniques   {
        if text := this.dbmap.dialect.CreateUniqueKey (key, list...); text != "" { uniques   = append(uniques  , text) }
    }
    for key, list := range this.indexes   {
        if text := this.dbmap.dialect.CreateIndexKey  (key, list...); text != "" { indexes   = append(indexes  , text) }
    }
    //*/
    if len(primaries) > 0 { lines = append(lines, primaries...) }
    if len(uniques  ) > 0 { lines = append(lines, uniques...  ) }
    if len(indexes  ) > 0 { lines = append(lines, indexes...  ) }
    columns := strings.Join(lines, ",\n")
    return fmt.Sprintf(f, columns)
}
func (this *tableMap) Create(ifNotExists bool) (error) { return this.exec( this.CreateSQL(ifNotExists) ) }
func (this *tableMap) DropSQL(ifExists bool) (string) {
    return this.dbmap.dialect.DropTableSQL(this.schemaName, this.tableName, ifExists)
}
func (this *tableMap) Drop(ifExists bool) (error) { return this.exec( this.DropSQL(ifExists) ) }
func (this *tableMap) TruncateSQL() (string) {
    return this.dbmap.dialect.TruncateTableSQL(this.schemaName, this.tableName)
}
func (this *tableMap) Truncate() (error) { return this.exec( this.TruncateSQL() ) }

func (this *tableMap) Exec(query string, args ...interface{}) (sql.Result, error) { return this.dbmap.Exec(query, args...) }
func (this *tableMap) exec(query string, args ...interface{}) (err error) { _, err = this.Exec(query, args...); return }
func (this *tableMap) Query(query string, args ...interface{}) (*sql.Rows, error) { return this.dbmap.Query(query, args...) }
func (this *tableMap) QueryRow(query string, args ...interface{}) (*sql.Row) { return this.dbmap.QueryRow(query, args...) }
