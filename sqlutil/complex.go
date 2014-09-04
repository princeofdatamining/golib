
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
    m.bind(t)
    m.SetAs(args...)
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
    suffixs []string
    //
    page_grouping   bool
    page_maxNav     int
    page_perRows    int
    page_sql        string
    page_sql_count  string
    page_allRows    int
    page_count      int
    page_first      int
    page_last       int
    page_left       int
    page_right      int
    page_select     int
}
func (this *SQLQuery) bind(t TableMap) () {
    this.table, _ = t.(*tableMap)
    this.db = this.table.dbmap
    this.dialect = this.db.dialect
    return 
}
func isAll(s string) (bool) {
    switch s {
    case "", "*", "all", "ALL":
        return true
    }
    return false
}
func (this *SQLQuery) SetAs(args ...string) (*SQLQuery) {
    if len(args) <= 0 {
        this.as = ""
    } else {
        this.as = strings.TrimSpace(args[0])
    }
    return this
}
func (this *SQLQuery) SetFields(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); s == "" {
        if this.as != "" {
            s = this.as
        } else {
            s = this.table.quoteTable()
        }
        s += ".*"
    }
    this.fields = s
    return this
}
func (this *SQLQuery) SetJoin(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); false {
        //
    }
    this.joins = s
    return this
}
func (this *SQLQuery) SetWhere(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); isAll(s) {
        s = "1"
    }
    this.wheres = s
    return this
}
func (this *SQLQuery) SetGroupBy(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); s != "" {
        s = sGroupBy + " " + s
    }
    this.groups = s
    return this
}
func (this *SQLQuery) SetHaving(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); s != "" {
        s = sHaving + " " + s
    }
    this.having = s
    return this
}
func (this *SQLQuery) SetOrderBy(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); s != "" {
        s = sOrderBy + " " + s
    }
    this.orders = s
    return this
}
func (this *SQLQuery) SetLimit(s string) (*SQLQuery) {
    if s = strings.TrimSpace(s); isAll(s) {
        s = ""
    } else {
        s = sLimit + " " + s
    }
    this.limits = s
    return this
}
func (this *SQLQuery) SetSuffixes(suffixes ...string) (*SQLQuery) {
    this.suffixs = suffixes
    return this
}
var (
    reTableAs = regexp.MustCompile(sFrom + "\\s+\\S+")
    reLimit = regexp.MustCompile("^\\s*" + sLimit)
    reGroup = regexp.MustCompile("^\\s*" + sGroupBy)
    reSemicolon = regexp.MustCompile(";\\s*$")
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
    var page_suffix string
    if pageMode {
        page_suffix = " %s"
    }
    return fmt.Sprintf(selSQL, "", this.fields, this.joins, this.wheres, suffix + page_suffix)
}
func (this *SQLQuery) get (all bool, holder interface{}, exec SQLExecutor,                       args ...interface{}) (rows int64, err error) {
    if exec == nil {
        exec = this.table
    }
    if all {
        rows, err = exec.SelectAll(holder, this.MakeSQL(false, this.suffixs...), args...)
    } else {
        err       = exec.SelectOne(holder, this.MakeSQL(false, this.suffixs...), args...)
    }
    return 
}
func (this *SQLQuery) GetAll(slices interface{}, exec SQLExecutor, args ...interface{}) (int64, error) {
    return this.get( true, slices, exec, args...)
}
func (this *SQLQuery) GetOne(holder interface{}, exec SQLExecutor, args ...interface{}) (err error) {
    _, err = this.get(false, holder, exec, args...)
    return 
}

func (this *SQLQuery) SetPageMode(maxNavPage, rowsPerPage int) () {
    if this.page_maxNav = maxNavPage; maxNavPage < 1 {
        this.page_maxNav = nMaxNavPage
    }
    if this.page_perRows = rowsPerPage; rowsPerPage < 1 {
        this.page_perRows = nRowsPerPage
    }
}
func (this *SQLQuery) InitPage(args ...interface{}) () {
    this.page_sql = this.MakeSQL(true, this.suffixs...)
    //
    if this.page_grouping {
        s := reSemicolon.ReplaceAllString(this.page_sql, "")
        this.page_sql_count = fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS IPaging;", fmt.Sprintf(s, ""))
    }
    rows, _ := this.table.SelectInt(this.page_sql_count, args...)
    allRows := int(rows)
    //
    this.page_allRows = allRows
    this.page_count = (allRows + this.page_perRows - 1) / this.page_perRows
    this.page_first = 1
    this.page_last = this.page_first + this.page_count - 1
    //
    this.page_left = this.page_first
    if this.page_right = this.page_left + this.page_maxNav - 1; this.page_right > this.page_last {
        this.page_right = this.page_last
    }
}
func (this *SQLQuery) InitPage1(        where string, args ...interface{}) () {
    this.SetWhere(where)
    this.InitPage(args...)
}
func (this *SQLQuery) InitPage2(fields, where string, args ...interface{}) () {
    this.SetFields(fields)
    this.SetWhere(where)
    this.InitPage(args...)
}
func (this *SQLQuery) GetPage(slices interface{}, pageNo int, args ...interface{}) (rows int64, err error) {
    if (this.page_allRows <= 0) {
        return 
    }
    switch {
    case pageNo < this.page_first:
        this.page_select = this.page_first
    case pageNo > this.page_last:
        this.page_select = this.page_last
    default:
        this.page_select = pageNo
    }
    if this.page_left      = this.page_select - (this.page_maxNav/2); this.page_left  < this.page_first {
        this.page_left     = this.page_first
    }
    if this.page_right     = this.page_left  + (this.page_maxNav-1); this.page_right > this.page_last   {
        this.page_right    = this.page_last
        if this.page_left  = this.page_right - (this.page_maxNav-1); this.page_left  < this.page_first  {
            this.page_left = this.page_first
        }
    }
    start := this.page_perRows * (this.page_select-this.page_first)
    s := fmt.Sprintf(this.page_sql, fmt.Sprintf("%s %d,%d", sLimit, start, this.page_perRows))
    return this.table.SelectAll(slices, s, args...)
}
var HTML_Paging = `
<![CDATA[FIRST]]>
<![CDATA[PREVN]]>
<![CDATA[PREV]]>
<![CDATA[BODY]]>
<![CDATA[NEXT]]>
<![CDATA[NEXTN]]>
<![CDATA[LAST]]>
<![CDATA[INFO]]>
`
var HTML_Paging_Elements = map[string]string{
    "FIRST": "[page] : href='[href]'",
    "PREVN": "[page] : href='[href]'",
    "PREV" : "[page] : href='[href]'",
    "NEXT" : "[page] : href='[href]'",
    "NEXTN": "[page] : href='[href]'",
    "LAST" : "[page] : href='[href]'",
    "INFO" : "[page]/[count]",
    //
    "SELECT": "\t[page]# href='[href]'\n",
    "AROUND": "\thref='[href]', [page]\n",
}
func (this *SQLQuery) GetPageBar(pattern string, elements map[string]string, f func (page int) (string)) (h string) {
    var (
        elem, body string
        page int
    )
    h = pattern
    p := func (page int) (string) { return fmt.Sprintf("%d", page) }
    //
    elem, page = "FIRST", this.page_first
    if replace, ok := elements[elem]; ok && this.page_select != page {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    elem, page = "LAST" , this.page_last
    if replace, ok := elements[elem]; ok && this.page_select != page {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    //
    elem, page = "PREVN", this.page_select - this.page_maxNav
    if replace, ok := elements[elem]; ok && page >= this.page_first {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    elem, page = "NEXTN", this.page_select + this.page_maxNav
    if replace, ok := elements[elem]; ok && page <= this.page_last  {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    //
    elem, page = "PREV" , this.page_select - 1
    if replace, ok := elements[elem]; ok && page >= this.page_first {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    elem, page = "NEXT" , this.page_select + 1
    if replace, ok := elements[elem]; ok && page <= this.page_last  {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    //
    elem, page = "INFO" , this.page_select
    if replace, ok := elements[elem]; ok {
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[count]",p(this.page_count), -1)
        h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), replace, 1)
    }
    elem, body = "BODY" , ""
    repSelect, repAround := elements["SELECT"], elements["AROUND"]
    for page = this.page_left; page <= this.page_right; page++ {
        replace := repAround
        if page == this.page_select { replace = repSelect }
        replace = strings.Replace(replace, "[page]", p(page), -1)
        replace = strings.Replace(replace, "[href]", f(page), -1)
        body += replace
    }
    h = strings.Replace(h, fmt.Sprintf("<![CDATA[%s]]>", elem), body, 1)
    //
    return 
}
