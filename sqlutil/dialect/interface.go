
package dialect

import (
    "fmt"
    "reflect"
    "database/sql"
)

type ColumnMeta interface {
    GetColumnName() (string)
    GetFieldName() (string)
    GetAutoIncr() (bool)
    GetNotNull() (bool)
    GetGoType() (reflect.Type)
    GetForceType() (string)
    GetSize() (int, int)
    GetDefault() (bool, string)
    GetComment() (string)
}
type Execer interface {
    Exec(query string, args ...interface{}) (sql.Result, error)
}
type Queryer interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) (*sql.Row)
}
type exec_queryer interface {
    Execer
    Queryer
}
type Dialect interface {
    QuoteField(f string) (string)
    QuotedTableForQuery(schemaName, tableName string) (string)
    BindVar(i int) (string)
    BindAutoIncrVar() (string)
    InsertAndReturnId(exec exec_queryer, query string, args ...interface{}) (int64, error)

    PrimaryKeyStr() (string)
    UniqueKeyStr() (string)
    NormalKeyStr() (string)

    CreateSchema(schemaName string, ifNotExists bool) (string)
    CreateColumnStr(col ColumnMeta) (string)
    CreatePrimaryKey(key string, cols ...ColumnMeta) (string)
    CreateUniqueKey (key string, cols ...ColumnMeta) (string)
    CreateIndexKey  (key string, cols ...ColumnMeta) (string)
    CreateTableSQL(schemaName, tableName string, ifNotExists bool, params map[string]string) (string)
    DropTableSQL(schemaName, tableName string, ifExists bool) (string)
    TruncateTableSQL(schemaName, tableName string) (string)
    InsertSQL(schemaName, tableName string, autoincr ColumnMeta) (string)
    UpdateSQL(schemaName, tableName string) (string)
    SelectSQL(schemaName, tableName string) (string)
    DeleteSQL(schemaName, tableName string) (string)
}

type newDialect func (params map[string]string) (Dialect)

var dialects = map[string]newDialect{}

func Register(name string, newf newDialect) {
    if _, exists := dialects[name]; exists {
        panic(fmt.Sprintf("Dialect %q has exists\n", name))
    }
    dialects[name] = newf
}

func Open(name string, params map[string]string) (dialect Dialect, err error) {
    if newf, exists := dialects[name]; !exists {
        return nil, fmt.Errorf("Dialect %q not found", name)
    } else {
        return newf(params), nil
    }
}

//

func QuoteField(f string) (string) { return fmt.Sprintf("`%s`", f) }
func CreatePrimaryKey(this Dialect, key string, cols []ColumnMeta) (s string) {
    s = fmt.Sprintf("  %s (", this.PrimaryKeyStr())
    for i, col := range cols {
        if i > 0 { s += ", " }
        s += this.QuoteField(col.GetColumnName())
    }
    return s + ")"
}
func CreateUniqueKey(this Dialect, key string, cols []ColumnMeta) (s string) {
    s = fmt.Sprintf("  %s %s (", this.UniqueKeyStr(), this.QuoteField(key))
    for i, col := range cols {
        if i > 0 { s += ", " }
        s += this.QuoteField(col.GetColumnName())
    }
    return s + ")"
}
func CreateIndexKey(this Dialect, key string, cols []ColumnMeta) (s string) {
    s = fmt.Sprintf("  %s %s (", this.NormalKeyStr(), this.QuoteField(key))
    for i, col := range cols {
        if i > 0 { s += ", " }
        s += this.QuoteField(col.GetColumnName())
    }
    return s + ")"
}
func CreateTableSQL(this Dialect, schemaName, tableName string, ifNotExists bool, suffix string) (string) {
    s0 := this.CreateSchema(schemaName, ifNotExists)
    addIfNotExists := ""
    if ifNotExists {
        addIfNotExists = " IF NOT EXISTS"
    }
    s := fmt.Sprintf("CREATE TABLE%s %s (\n", addIfNotExists, this.QuotedTableForQuery(schemaName, tableName))
    return s0 + s + "%s\n)" + suffix + ";"
}
func DropTableSQL(this Dialect, schemaName, tableName string, ifExists bool) (string) {
    addIfExists := ""
    if ifExists {
        addIfExists = " IF EXISTS"
    }
    quoted := this.QuotedTableForQuery(schemaName, tableName)
    return fmt.Sprintf("DROP TABLE%s %s;", addIfExists, quoted)
}
func InsertAndReturnId(this Dialect, exec exec_queryer, query string, args ...interface{}) (id int64, err error) {
    res, err := exec.Exec(query, args...)
    if err != nil {
        return 
    }
    return res.LastInsertId()
}
func InsertSQL(this Dialect, schemaName, tableName string, suffix string) (string) {
    return fmt.Sprintf("INSERT INTO %s (%%s) VALUES (%%s)%s;", this.QuotedTableForQuery(schemaName, tableName), suffix)
}
func UpdateSQL(this Dialect, schemaName, tableName string) (string) {
    return fmt.Sprintf("UPDATE %s SET %%s WHERE %%s;", this.QuotedTableForQuery(schemaName, tableName))
}
func SelectSQL(this Dialect, schemaName, tableName string) (string) {
    return fmt.Sprintf("SELECT %%s FROM %s WHERE %%s %%s;", this.QuotedTableForQuery(schemaName, tableName))
}
func DeleteSQL(this Dialect, schemaName, tableName string) (string) {
    return fmt.Sprintf("DELETE FROM %s WHERE %%s;", this.QuotedTableForQuery(schemaName, tableName))
}
