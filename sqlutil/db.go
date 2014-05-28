
package sqlutil

import (
    "fmt"
    "database/sql"
    "github.com/princeofdatamining/golib/sqlutil/dialect"
)

type errFormatFunc func (args ...interface{}) (error)

func errFormatFactory(f string) (errFormatFunc) {
    return func (args ...interface{}) (error) {
        return fmt.Errorf(f, args...)
    }
}

type SQLExecutor interface {
    dialect.Execer
    dialect.Queryer
    //
    Insert(objects ...interface{}) (int64, error)
    Update(objects ...interface{}) (int64, error)
    Delete(objects ...interface{}) (int64, error)
    Get(objects ...interface{}) ( int64, error)
    //
    SelectBool(query string, args ...interface{}) (bool, error)
    SelectNullBool(query string, args ...interface{}) (sql.NullBool, error)
    SelectInt(query string, args ...interface{}) (int64, error)
    SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error)
    SelectFloat(query string, args ...interface{}) (float64, error)
    SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error)
    SelectStr(query string, args ...interface{}) (string, error)
    SelectNullStr(query string, args ...interface{}) (sql.NullString, error)
    //
    SelectVal(holder interface{}, query string, args ...interface{}) (error)
    SelectOne(holder interface{}, query string, args ...interface{}) (error)
    SelectAll(slices interface{}, query string, args ...interface{}) (int64, error)
}

type DbMap interface {
    SQLExecutor
    Begin() (Transaction, error)
    //
    GetTableByName(t string) (TableMap, bool)
    GetTableByMeta(meta interface{}) (TableMap, bool)
    AddTable(meta interface{}, name string) (TableMap, error)
    AddTable2(meta interface{}, name, comment string) (TableMap, error)
    AddTable3(meta interface{}, schema, name, comment string) (TableMap, error)
    CreateTables(ifNotExists bool, args ...interface{}) (sql string, err error)
    TruncateTables(args ...interface{}) (sql string, err error)
    DropTables  (ifExists    bool, args ...interface{}) (sql string, err error)
    DropTableByName(t string, ifExists bool) (error)
    DropTableByMeta(meta interface{}, ifExists bool) (error)
}
func NewDbMap(db *sql.DB, dialect dialect.Dialect) (DbMap) {
    return &dbMap{
        db: db,
        dialect: dialect,

        tableD:  make(map[string]*tableMap),
    }
}
type dbMap struct {
    db      *sql.DB
    dialect dialect.Dialect

    tables  []*tableMap
    tableD  map[string]*tableMap

    pseudos []*tableMap
}
func (this *dbMap) Exec(query string, args ...interface{}) (sql.Result, error) {
    fmt.Println("Db.Exec:", query)
    return this.db.Exec(query, args...)
}
func (this *dbMap) Query(query string, args ...interface{}) (*sql.Rows, error) {
    fmt.Println("Db.Query:", query)
    return this.db.Query(query, args...)
}
func (this *dbMap) QueryRow(query string, args ...interface{}) (*sql.Row) {
    fmt.Println("Db.QueryRow:", query)
    return this.db.QueryRow(query, args...)
}

var _, _, _ SQLExecutor = NewDbMap(nil, nil), &tableMap{}, &txMap{}
