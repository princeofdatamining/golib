
package sqlutil

import (
    "fmt"
    "strings"
    "regexp"
    "github.com/princeofdatamining/golib/sqlutil/dialect"
)

var (
    sFrom    = "FROM"
    sAs      = "AS"
    sGroupBy = "GROUP BY"
    sHaving  = "HAVING"
    sOrderBy = "ORDER BY"
    sLimit   = "LIMIT"
)
const (
    nMaxNavPage = 10
    nRowsPerPage = 20
)
func NewSQLQuery(t TableMap, args ...string) (m *SQLQuery) {
    m = &SQLQuery{
        page_maxNav: nMaxNavPage,
        page_perRows: nRowsPerPage,
    }
    var as string
    if args != nil {
        as = args[0]
    }
    m.bind(t, as)
    return 
}
type SQLQuery struct {
    dialect dialect.Dialect
    db      *dbMap
    table   *tableMap
    fields  string
    as      string
    joins   string
    wheres  string
    groups  string
    having  string
    orders  string
    limits  string
    //
    page_grouping   bool
    page_maxNav     int
    page_perRows    int
    page_sql        string
    page_sql_count  string
    page_allRows    int
    page_count      int
    page_base       int
    page_last       int
    page_from       int
    page_upto       int
    page_curr       int
}
func (this *SQLQuery) bind(t TableMap, as string) () {
    this.table, _ = t.(*tableMap)
    this.db = this.table.dbmap
    this.dialect = this.db.dialect
    this.as = strings.TrimSpace(as)
}
func isAll(s string) (bool) {
    switch s {
    case "", "*", "all", "ALL":
        return true
    }
    return false
}
func (this *SQLQuery) SetFields(s string) () {
    if s = strings.TrimSpace(s); s == "" {
        if this.as == "" {
            s = this.as
        } else {
            s = this.table.quoteTable()
        }
        s += ".*"
    }
    this.fields = s
}
func (this *SQLQuery) SetJoin(s string) () {
    if s = strings.TrimSpace(s); false {
        //
    }
    this.joins = s
}
func (this *SQLQuery) SetWhere(s string) () {
    if s = strings.TrimSpace(s); isAll(s) {
        s = "1"
    }
    this.wheres = s
}
func (this *SQLQuery) SetGroup(s string) () {
    if s = strings.TrimSpace(s); s != "" {
        s = sGroupBy + " " + s
    }
    this.groups = s
}
func (this *SQLQuery) SetHaving(s string) () {
    if s = strings.TrimSpace(s); s != "" {
        s = sHaving + " " + s
    }
    this.having = s
}
func (this *SQLQuery) SetOrderBy(s string) () {
    if s = strings.TrimSpace(s); s != "" {
        s = sOrderBy + " " + s
    }
    this.orders = s
}
func (this *SQLQuery) SetLimit(s string) () {
    if s = strings.TrimSpace(s); isAll(s) {
        s = ""
    } else {
        s = sLimit + " " + s
    }
    this.limits = s
}
var (
    reTableAs = regexp.MustCompile(sFrom + "\\s+\\S+")
    reLimit = regexp.MustCompile("^\\s*" + sLimit)
    reGroup = regexp.MustCompile("^\\s*" + sGroupBy)
)
func (this *SQLQuery) MakeSQL(pageMode bool, suffixes ...string) (s string) {
    selSQL := this.dialect.SelectSQL(this.table.schemaName, this.table.tableName)
    if this.as != "" {
        selSQL = reTableAs.ReplaceAllString(selSQL, "${0} " + sAs + " " + this.as)
    }
    var suffix string
    if this.groups != "" {
        suffix += " " + this.groups
        this.page_grouping = true
    }
    if this.having != "" {
        suffix += " " + this.having
    }
    if this.orders != "" {
        suffix += " " + this.orders
    }
    if this.limits != "" && !pageMode {
        suffix += " " + this.limits
    }
    for _, sfx := range suffixes {
        if sfx = strings.TrimSpace(sfx); sfx != "" {
            if !pageMode || !reLimit.MatchString(sfx) {
                suffix += " " + sfx
            }
            this.page_grouping = reGroup.MatchString(sfx)
        }
    }
    this.page_sql_count = fmt.Sprintf(selSQL, "", "COUNT(*)", this.joins, this.wheres, "")
    return fmt.Sprintf(selSQL, "", this.fields, this.joins, this.wheres, suffix)
}
func (this *SQLQuery) Get (slices interface{}, exec SQLExecutor,                       suffixes ...string) (int64, error) {
    if exec == nil {
        exec = this.table
    }
    return exec.SelectAll(slices, this.MakeSQL(false, suffixes...))
}
func (this *SQLQuery) Get1(slices interface{}, exec SQLExecutor,         where string, suffixes ...string) (int64, error) {
    this.SetWhere(where)
    return this.Get(slices, exec, suffixes...)
}
func (this *SQLQuery) Get2(slices interface{}, exec SQLExecutor, fields, where string, suffixes ...string) (int64, error) {
    this.SetFields(fields)
    this.SetWhere(where)
    return this.Get(slices, exec, suffixes...)
}

func (this *SQLQuery) SetPageMode(maxNavPage, rowsPerPage int) () {
    if this.page_maxNav = maxNavPage; maxNavPage < 1 {
        this.page_maxNav = nMaxNavPage
    }
    if this.page_perRows = rowsPerPage; rowsPerPage < 1 {
        this.page_perRows = nRowsPerPage
    }
}
func (this *SQLQuery) InitPage(suffixes ...string) () {
    this.page_sql = this.MakeSQL(true, suffixes...)
    //
    if this.page_grouping {
        this.page_sql_count = fmt.Sprintf("SELECT COUNT(*) FROM (%s)", this.page_sql)
    }
    rows, _ := this.table.SelectInt(this.page_sql_count)
    allRows := int(rows)
    //
    this.page_allRows = allRows
    this.page_count = (allRows + this.page_perRows - 1) / this.page_perRows
    this.page_base = 1
    this.page_last = this.page_base + this.page_count - 1
    //
    this.page_from = this.page_base
    if this.page_upto = this.page_from + this.page_maxNav - 1; this.page_upto > this.page_last {
        this.page_upto = this.page_last
    }
}
func (this *SQLQuery) InitPage1(        where string, suffixes ...string) () {
    this.SetWhere(where)
    this.InitPage(suffixes...)
}
func (this *SQLQuery) InitPage2(fields, where string, suffixes ...string) () {
    this.SetFields(fields)
    this.SetWhere(where)
    this.InitPage(suffixes...)
}
func (this *SQLQuery) GetPage(slices interface{}, pageNo int) (rows int64, err error) {
    if (this.page_allRows <= 0) {
        return 
    }
    switch {
    case pageNo < this.page_base:
        this.page_curr = this.page_base
    case pageNo > this.page_last:
        this.page_curr = this.page_last
    default:
        this.page_curr = pageNo
    }
    if this.page_from = this.page_curr - (this.page_maxNav/2); this.page_from < this.page_base {
        this.page_from = this.page_base
    }
    if this.page_upto = this.page_from + (this.page_maxNav-1); this.page_upto > this.page_last {
        this.page_upto = this.page_last
        if this.page_from = this.page_upto+1-this.page_maxNav; this.page_from < this.page_base {
            this.page_from = this.page_base
        }
    }
    start := this.page_perRows * (this.page_curr-this.page_base)
    s := fmt.Sprintf("%s %s %d,%d", this.page_sql, sLimit, start, this.page_perRows)
    return this.table.SelectAll(slices, s)
}
func (this *SQLQuery) GetPageBar(pattern string) (h string) {
    return 
}
