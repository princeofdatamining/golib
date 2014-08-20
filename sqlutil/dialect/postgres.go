
package dialect

import (
    "fmt"
    "reflect"
    "strings"
)

func init() () {
    Register("postgres", func (params map[string]string) (Dialect) {
        dialect := new(postgresDialect)
        var ok bool
        if dialect.suffix, ok = params["suffix"]; !ok {
            //
        }
        return dialect
    })
}

type postgresDialect struct {
    suffix string
}

func (this *postgresDialect) QuoteField(f string) (string) { return QuoteField(strings.ToLower(f)) }
func (this *postgresDialect) QuoteTable(schemaName, tableName string) (q string) {
    q = this.QuoteField(tableName)
    if schemaName != "" {
        q = this.QuoteField(schemaName) + "." + q
    }
    return 
}
func (this *postgresDialect) BindVar(i int) (string) { return fmt.Sprintf("$%d", i) }
func (this *postgresDialect) BindAutoIncrVar() (string) { return "DEFAULT" }

func (this *postgresDialect) PrimaryKeyStr() (string) { return "PRIMARY KEY" }
func (this *postgresDialect) UniqueKeyStr () (string) { return "UNIQUE KEY" }
func (this *postgresDialect) NormalKeyStr () (string) { return "KEY" }

func (this *postgresDialect) CreateSchema(schemaName string, ifNotExists bool) (s string) {
    if schemaName == "" {
        return 
    }
    addIfNotExists := ""
    if ifNotExists {
        addIfNotExists = " IF NOT EXISTS"
    }
    return fmt.Sprintf("CREATE SCHEMA%s %s;\n", addIfNotExists, schemaName)
}
func (this *postgresDialect) createColumnType(col ColumnMeta) (string) {
    var tpname string
    if tpname = col.GetForceType(); tpname == "" {
        tpname = this.toSqlType( col.GetGoType(), col.GetAutoIncr() )
    }

    size, precision := col.GetSize()
    switch tpname {
    case "text", "timestamp with time zone":
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
    case "real", "double precision":
        //
    default:
        precision = -1
    }
    if precision >= 0 {
        return fmt.Sprintf("%s(%d,%d)", tpname, size, precision)
    }

    return fmt.Sprintf("%s(%d)", tpname, size)
}
func inttype(big, autoincr bool) (string) {
    if big {
        if autoincr {
            return "bigserial"
        }
        return "bigint"
    } else {
        if autoincr {
            return "serial"
        }
        return "integer"
    }
}
func (this *postgresDialect) toSqlType(t reflect.Type, autoincr bool) (string) {
    switch t.Kind() {
        case reflect.Ptr: return this.toSqlType(t.Elem(), autoincr)
        case reflect.Bool: return "boolean"
        case reflect.Int8 : return inttype(false, autoincr)
        case reflect.Int16: return inttype(false, autoincr)
        case reflect.Int32, reflect.Int: return inttype(false, autoincr)
        case reflect.Int64: return inttype(true, autoincr)
        case reflect.Uint8 : return inttype(false, autoincr)
        case reflect.Uint16: return inttype(false, autoincr)
        case reflect.Uint32, reflect.Uint: return inttype(false, autoincr)
        case reflect.Uint64: return inttype(true, autoincr)
        case reflect.Float32: return "real"
        case reflect.Float64: return "double precision"
        case reflect.Slice: return "bytea"
    }
    switch t.Name() {
        case "NullBool": return "boolean"
        case "NullInt64": return "bigint"
        case "NullFloat64": return "double precision"
        case "Time": return "datetime"
    }
    return "varchar"
}
func (this *postgresDialect) CreateColumnStr(col ColumnMeta) (string) {
    var s string
    if col.GetAutoIncr() || col.GetNotNull() {
        s += " NOT NULL"
    }
    /*
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
    //*/
    return fmt.Sprintf("  %s %s%s", this.QuoteField(col.GetColumnName()), this.createColumnType(col), s)
}
func (this *postgresDialect) CreatePrimaryKey(key string, cols ...ColumnMeta) (s string) { return CreatePrimaryKey(this, key, cols) }
func (this *postgresDialect) CreateUniqueKey (key string, cols ...ColumnMeta) (s string) { return CreateUniqueKey (this, key, cols) }
func (this *postgresDialect) CreateIndexKey  (key string, cols ...ColumnMeta) (s string) { return CreateIndexKey  (this, key, cols) }
func (this *postgresDialect) createTableSuffix(params map[string]string) (string) { return this.suffix }
func (this *postgresDialect) CreateTableSQL(schemaName, tableName string, ifNotExists bool, params map[string]string) (string) {
    return CreateTableSQL(this, schemaName, tableName, ifNotExists, this.createTableSuffix(params))
}
func (this *postgresDialect) DropTableSQL(schemaName, tableName string, ifExists bool) (string) {
    return DropTableSQL(this, schemaName, tableName, ifExists)
}
func (this *postgresDialect) TruncateTableSQL(schemaName, tableName string) (string) {
    return fmt.Sprintf("TRUNCATE %s;", this.QuoteTable(schemaName, tableName))
}
func (this *postgresDialect) InsertAndReturnId(exec exec_queryer, query string, args ...interface{}) (id int64, err error) {
    rows, err := exec.Query(query, args...)
    if err != nil {
        return 
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&id)
    }
    return 
}
func (this *postgresDialect) InsertSQL(schemaName, tableName string, autoincr ColumnMeta) (string) {
    var suffix string
    if autoincr == nil {
    } else if colName := autoincr.GetColumnName(); colName != "" {
        suffix = fmt.Sprintf(" RETURNING %s", colName)
    }
    return InsertSQL(this, schemaName, tableName, suffix)
}
func (this *postgresDialect) UpdateSQL(schemaName, tableName string) (string) {
    return UpdateSQL(this, schemaName, tableName)
}
func (this *postgresDialect) SelectSQL(schemaName, tableName string) (string) {
    return SelectSQL(this, schemaName, tableName)
}
func (this *postgresDialect) DeleteSQL(schemaName, tableName string) (string) {
    return DeleteSQL(this, schemaName, tableName)
}
