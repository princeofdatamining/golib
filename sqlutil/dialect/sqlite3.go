
package dialect

import (
    "fmt"
    "reflect"
)

func init() () {
    Register("sqlite", func (params map[string]string) (Dialect) {
        dialect := new(sqliteDialect)
        var ok bool
        if dialect.suffix, ok = params["suffix"]; !ok {
            //
        }
        return dialect
    })
}

type sqliteDialect struct {
    suffix string
}

func (this *sqliteDialect) QuoteField(f string) (string) { return QuoteField(f) }
func (this *sqliteDialect) QuotedTableForQuery(schemaName, tableName string) (string) { return this.QuoteField(tableName) }
func (this *sqliteDialect) BindVar(i int) (string) { return "?" }
func (this *sqliteDialect) BindAutoIncrVar() (string) { return "NULL" }

func (this *sqliteDialect) PrimaryKeyStr() (string) { return "PRIMARY KEY" }
func (this *sqliteDialect) UniqueKeyStr () (string) { return "UNIQUE KEY" }
func (this *sqliteDialect) NormalKeyStr () (string) { return "KEY" }

func (this *sqliteDialect) CreateSchema(schemaName string, ifNotExists bool) (string) { return "" }
func (this *sqliteDialect) createColumnType(col ColumnMeta) (string) {
    var tpname string
    if tpname = col.GetForceType(); tpname == "" {
        tpname = this.toSqlType( col.GetGoType() )
    }

    size, precision := col.GetSize()
    switch tpname {
    case "datetime":
        return tpname
    case "varchar":
        if size <= 0 {
            size = 255
        }
    }
    if size <= 0 {
        return tpname
    }

    switch tpname {
    case "real":
        //
    default:
        precision = -1
    }
    if precision >= 0 {
        return fmt.Sprintf("%s(%d,%d)", tpname, size, precision)
    }

    return fmt.Sprintf("%s(%d)", tpname, size)
}
func (this *sqliteDialect) toSqlType(t reflect.Type) (string) {
    switch t.Kind() {
        case reflect.Ptr: return this.toSqlType(t.Elem())
        case reflect.Bool: return "integer"
        case reflect.Int8 : return "integer"
        case reflect.Int16: return "integer"
        case reflect.Int32, reflect.Int: return "integer"
        case reflect.Int64: return "integer"
        case reflect.Uint8 : return "integer"
        case reflect.Uint16: return "integer"
        case reflect.Uint32, reflect.Uint: return "integer"
        case reflect.Uint64: return "integer"
        case reflect.Float32: return "real"
        case reflect.Float64: return "real"
        case reflect.Slice: return "blob"
    }
    switch t.Name() {
        case "NullBool": return "integer"
        case "NullInt64": return "integer"
        case "NullFloat64": return "real"
        case "Time": return "datetime"
    }
    return "varchar"
}
func (this *sqliteDialect) CreateColumnStr(col ColumnMeta) (string) {
    var s string
    if col.GetAutoIncr() || col.GetNotNull() {
        s += " NOT NULL"
    }
    if col.GetAutoIncr() {
        s += " AUTOINCREMENT"
    }
    /*
    if has, str := col.GetDefault(); has {
        switch str {
        case "NULL", "CURRENT_TIMESTAMP":
            s += fmt.Sprintf(" DEFAULT %s", str)
        default:
            s += fmt.Sprintf(" DEFAULT '%s'", str)
        }
    }
    if str := col.GetComment(); str != "" {
        s += fmt.Sprintf(" COMMENT '%s'", str)
    }
    //*/
    return fmt.Sprintf("  %s %s%s", this.QuoteField(col.GetColumnName()), this.createColumnType(col), s)
}
func (this *sqliteDialect) CreatePrimaryKey(key string, cols ...ColumnMeta) (s string) { return CreatePrimaryKey(this, key, cols) }
func (this *sqliteDialect) CreateUniqueKey (key string, cols ...ColumnMeta) (s string) { return CreateUniqueKey (this, key, cols) }
func (this *sqliteDialect) CreateIndexKey  (key string, cols ...ColumnMeta) (s string) { return CreateIndexKey  (this, key, cols) }
func (this *sqliteDialect) createTableSuffix(params map[string]string) (string) { return this.suffix }
func (this *sqliteDialect) CreateTableSQL(schemaName, tableName string, ifNotExists bool, params map[string]string) (string) {
    return CreateTableSQL(this, schemaName, tableName, ifNotExists, this.createTableSuffix(params))
}
func (this *sqliteDialect) DropTableSQL(schemaName, tableName string, ifExists bool) (string) {
    return DropTableSQL(this, schemaName, tableName, ifExists)
}
func (this *sqliteDialect) TruncateTableSQL(schemaName, tableName string) (string) {
    return fmt.Sprintf("DELETE FROM %s;", this.QuotedTableForQuery(schemaName, tableName))
}
func (this *sqliteDialect) InsertAndReturnId(exec exec_queryer, query string, args ...interface{}) (int64, error) {
    return InsertAndReturnId(this, exec, query, args...)
}
func (this *sqliteDialect) InsertSQL(schemaName, tableName string, autoincr ColumnMeta) (string) {
    return InsertSQL(this, schemaName, tableName, "")
}
func (this *sqliteDialect) UpdateSQL(schemaName, tableName string) (string) {
    return UpdateSQL(this, schemaName, tableName)
}
func (this *sqliteDialect) SelectSQL(schemaName, tableName string) (string) {
    return SelectSQL(this, schemaName, tableName)
}
func (this *sqliteDialect) DeleteSQL(schemaName, tableName string) (string) {
    return DeleteSQL(this, schemaName, tableName)
}
