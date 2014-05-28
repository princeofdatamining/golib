
package sqlutil

import (
    "reflect"
    "regexp"
    "database/sql"
)

var (
    keyRegexp = regexp.MustCompile(`:[[:word:]]+`)
)

type (
    nameMapFunc func (key string) (reflect.Value)
)
func (this *dbMap) expandNamedQuery(query string, getter nameMapFunc) (formated string, args []interface{}) {
    var n int
    formated = keyRegexp.ReplaceAllStringFunc(query, func (holder string) (string) {
        val := getter(holder[1:])
        if !val.IsValid() {
            return holder
        }
        args = append(args, val.Interface())
        n++
        return this.dialect.BindVar(n-1)
    })
    return 
}
func (this *dbMap) maybeExpandNamedQuery(query string, args []interface{}) (string, []interface{}) {
    arg := reflect.Indirect(reflect.ValueOf(args[0]))
    aK, aT := arg.Kind(), arg.Type()
    switch {
    case aK == reflect.Map && aT.Key().Kind() == reflect.String:
        return this.expandNamedQuery(query, func (key string) (reflect.Value) {
            return arg.MapIndex(reflect.ValueOf(key))
        })
    case aK == reflect.Struct && !(aT.PkgPath() == "time" && aT.Name() == "Time"):
        return this.expandNamedQuery(query, arg.FieldByName)
    }
    return query, args
}
func (this *dbMap) selectVal(exec SQLExecutor, holder interface{}, query string, args ...interface{}) (err error) {
    if len(args) == 1 {
        query, args = this.maybeExpandNamedQuery(query, args)
    }
    if err = exec.QueryRow(query, args...).Scan(holder); err == sql.ErrNoRows {
        err = nil
    }
    return 
}
func (this *dbMap   ) SelectVal(holder interface{}, query string, args ...interface{}) (error) { return this      .selectVal(this, holder, query, args...) }
func (this *txMap   ) SelectVal(holder interface{}, query string, args ...interface{}) (error) { return this.dbmap.selectVal(this, holder, query, args...) }
func (this *tableMap) SelectVal(holder interface{}, query string, args ...interface{}) (error) { return this.dbmap.selectVal(this, holder, query, args...) }

func (this *dbMap) selectBool(exec SQLExecutor, query string, args ...interface{}) (val bool, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectBool(query string, args ...interface{}) (bool, error) { return this.selectBool(this, query, args...) }
func (this *txMap   ) SelectBool(query string, args ...interface{}) (bool, error) { return this.dbmap.selectBool(this, query, args...) }
func (this *tableMap) SelectBool(query string, args ...interface{}) (bool, error) { return this.dbmap.selectBool(this, query, args...) }

func (this *dbMap) selectNullBool(exec SQLExecutor, query string, args ...interface{}) (val sql.NullBool, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectNullBool(query string, args ...interface{}) (sql.NullBool, error) { return this.selectNullBool(this, query, args...) }
func (this *txMap   ) SelectNullBool(query string, args ...interface{}) (sql.NullBool, error) { return this.dbmap.selectNullBool(this, query, args...) }
func (this *tableMap) SelectNullBool(query string, args ...interface{}) (sql.NullBool, error) { return this.dbmap.selectNullBool(this, query, args...) }

func (this *dbMap) selectInt(exec SQLExecutor, query string, args ...interface{}) (val int64, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectInt(query string, args ...interface{}) (int64, error) { return this.selectInt(this, query, args...) }
func (this *txMap   ) SelectInt(query string, args ...interface{}) (int64, error) { return this.dbmap.selectInt(this, query, args...) }
func (this *tableMap) SelectInt(query string, args ...interface{}) (int64, error) { return this.dbmap.selectInt(this, query, args...) }

func (this *dbMap) selectNullInt(exec SQLExecutor, query string, args ...interface{}) (val sql.NullInt64, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error) { return this.selectNullInt(this, query, args...) }
func (this *txMap   ) SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error) { return this.dbmap.selectNullInt(this, query, args...) }
func (this *tableMap) SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error) { return this.dbmap.selectNullInt(this, query, args...) }

func (this *dbMap) selectFloat(exec SQLExecutor, query string, args ...interface{}) (val float64, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectFloat(query string, args ...interface{}) (float64, error) { return this.selectFloat(this, query, args...) }
func (this *txMap   ) SelectFloat(query string, args ...interface{}) (float64, error) { return this.dbmap.selectFloat(this, query, args...) }
func (this *tableMap) SelectFloat(query string, args ...interface{}) (float64, error) { return this.dbmap.selectFloat(this, query, args...) }

func (this *dbMap) selectNullFloat(exec SQLExecutor, query string, args ...interface{}) (val sql.NullFloat64, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error) { return this.selectNullFloat(this, query, args...) }
func (this *txMap   ) SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error) { return this.dbmap.selectNullFloat(this, query, args...) }
func (this *tableMap) SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error) { return this.dbmap.selectNullFloat(this, query, args...) }

func (this *dbMap) selectStr(exec SQLExecutor, query string, args ...interface{}) (val string, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectStr(query string, args ...interface{}) (string, error) { return this.selectStr(this, query, args...) }
func (this *txMap   ) SelectStr(query string, args ...interface{}) (string, error) { return this.dbmap.selectStr(this, query, args...) }
func (this *tableMap) SelectStr(query string, args ...interface{}) (string, error) { return this.dbmap.selectStr(this, query, args...) }

func (this *dbMap) selectNullStr(exec SQLExecutor, query string, args ...interface{}) (val sql.NullString, err error) {
    err = this.selectVal(exec, &val, query, args...)
    return 
}
func (this *dbMap   ) SelectNullStr(query string, args ...interface{}) (sql.NullString, error) { return this.selectNullStr(this, query, args...) }
func (this *txMap   ) SelectNullStr(query string, args ...interface{}) (sql.NullString, error) { return this.dbmap.selectNullStr(this, query, args...) }
func (this *tableMap) SelectNullStr(query string, args ...interface{}) (sql.NullString, error) { return this.dbmap.selectNullStr(this, query, args...) }
