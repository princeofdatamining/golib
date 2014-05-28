
package dialect

import (
    "fmt"
    "reflect"
)

func init() () {
    Register("mysql", func (params map[string]string) (Dialect) {
        dialect := new(mysqlDialect)
        var ok bool
        if dialect.engine, ok = params["engine"]; !ok {
            dialect.engine = "InnoDB"
        }
        if dialect.charset, ok = params["charset"]; !ok {
            dialect.charset = "utf8"
        }
        return dialect
    })
}

type mysqlDialect struct {
    engine      string
    charset     string
    comment     string
}

func (this *mysqlDialect) QuoteField(f string) (string) { return QuoteField(f) }
func (this *mysqlDialect) QuotedTableForQuery(schemaName, tableName string) (string) { return this.QuoteField(tableName) }
func (this *mysqlDialect) BindVar(i int) (string) { return "?" }
func (this *mysqlDialect) BindAutoIncrVar() (string) { return "NULL" }

func (this *mysqlDialect) PrimaryKeyStr() (string) { return "PRIMARY KEY" }
func (this *mysqlDialect) UniqueKeyStr () (string) { return "UNIQUE KEY" }
func (this *mysqlDialect) NormalKeyStr () (string) { return "KEY" }

func (this *mysqlDialect) CreateSchema(schemaName string, ifNotExists bool) (string) { return "" }
func (this *mysqlDialect) createColumnType(col ColumnMeta) (string) {
    var tpname string
    if tpname = col.GetForceType(); tpname == "" {
        tpname = this.toSqlType( col.GetGoType() )
    }

    size, precision := col.GetSize()
    switch tpname {
    case "text", "datetime", "timestamp":
        return tpname
    case "varchar", "char":
        if size <= 0 {
            size = 255
        }
    }
    if size <= 0 {
        return tpname
    }

    switch tpname {
    case "decimal", "double", "float":
        //
    default:
        precision = -1
    }
    if precision >= 0 {
        return fmt.Sprintf("%s(%d,%d)", tpname, size, precision)
    }

    return fmt.Sprintf("%s(%d)", tpname, size)
}
func (this *mysqlDialect) toSqlType(t reflect.Type) (string) {
    switch t.Kind() {
        case reflect.Ptr: return this.toSqlType(t.Elem())
        case reflect.Bool: return "boolean"
        case reflect.Int8 : return "tinyint"
        case reflect.Int16: return "smallint"
        case reflect.Int32, reflect.Int: return "int"
        case reflect.Int64: return "bigint"
        case reflect.Uint8 : return "tinyint unsigned"
        case reflect.Uint16: return "smallint unsigned"
        case reflect.Uint32, reflect.Uint: return "int unsigned"
        case reflect.Uint64: return "bigint unsigned"
        case reflect.Float32: return "float"
        case reflect.Float64: return "double"
        case reflect.Slice: return "mediumblob"
    }
    switch t.Name() {
        case "NullBool": return "tinyint"
        case "NullInt64": return "bigint"
        case "NullFloat64": return "double"
        case "Time": return "datetime"
    }
    return "varchar"
}
func (this *mysqlDialect) CreateColumnStr(col ColumnMeta) (string) {
    var s string
    if col.GetAutoIncr() || col.GetNotNull() {
        s += " NOT NULL"
    }
    if col.GetAutoIncr() {
        s += " AUTO_INCREMENT"
    }
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
    return fmt.Sprintf("  %s %s%s", this.QuoteField(col.GetColumnName()), this.createColumnType(col), s)
}
func (this *mysqlDialect) CreatePrimaryKey(key string, cols ...ColumnMeta) (s string) { return CreatePrimaryKey(this, key, cols) }
func (this *mysqlDialect) CreateUniqueKey (key string, cols ...ColumnMeta) (s string) { return CreateUniqueKey (this, key, cols) }
func (this *mysqlDialect) CreateIndexKey  (key string, cols ...ColumnMeta) (s string) { return CreateIndexKey  (this, key, cols) }
func (this *mysqlDialect) createTableSuffix(params map[string]string) (string) {
    var (
        extend string
    )
    if s, ok := params["comment"]; ok && s != "" {
        extend += fmt.Sprintf(" COMMENT='%s'", s)
    }
    return fmt.Sprintf(" ENGINE=%s CHARSET=%s%s", this.engine, this.charset, extend)
}
func (this *mysqlDialect) CreateTableSQL(schemaName, tableName string, ifNotExists bool, params map[string]string) (string) {
    return CreateTableSQL(this, schemaName, tableName, ifNotExists, this.createTableSuffix(params))
}
func (this *mysqlDialect) DropTableSQL(schemaName, tableName string, ifExists bool) (string) {
    return DropTableSQL(this, schemaName, tableName, ifExists)
}
func (this *mysqlDialect) TruncateTableSQL(schemaName, tableName string) (string) {
    return fmt.Sprintf("TRUNCATE %s;", this.QuotedTableForQuery(schemaName, tableName))
}
func (this *mysqlDialect) InsertAndReturnId(exec exec_queryer, query string, args ...interface{}) (int64, error) {
    return InsertAndReturnId(this, exec, query, args...)
}
func (this *mysqlDialect) InsertSQL(schemaName, tableName string, autoincr ColumnMeta) (string) {
    return InsertSQL(this, schemaName, tableName, "")
}
func (this *mysqlDialect) UpdateSQL(schemaName, tableName string) (string) {
    return UpdateSQL(this, schemaName, tableName)
}
func (this *mysqlDialect) SelectSQL(schemaName, tableName string) (string) {
    return SelectSQL(this, schemaName, tableName)
}
func (this *mysqlDialect) DeleteSQL(schemaName, tableName string) (string) {
    return DeleteSQL(this, schemaName, tableName)
}
