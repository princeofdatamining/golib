
package httputil

import (
    "net/http"
    "strings"
    "fmt"
    "bytes"
    "strconv"
    "sort"
)

// Get the content type.
// e.g. From "multipart/form-data; boundary=--" to "multipart/form-data"
// If none is specified, returns "text/html" by default.
func ResolveContentType(req *http.Request) string {
    contentType := req.Header.Get("Content-Type")
    if contentType == "" {
        return "text/html"
    }
    return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

// ResolveFormat maps the request's Accept MIME type declaration to
// a Request.Format attribute, specifically "html", "xml", "json", or "txt",
// returning a default of "html" when Accept header cannot be mapped to a
// value above.
func ResolveFormat(req *http.Request) string {
    accept := req.Header.Get("accept")

    switch {
    case accept == "",
        strings.HasPrefix(accept, "*/*"), // */
        strings.Contains(accept, "application/xhtml"),
        strings.Contains(accept, "text/html"):
        return "html"
    case strings.Contains(accept, "application/xml"),
        strings.Contains(accept, "text/xml"):
        return "xml"
    case strings.Contains(accept, "text/plain"):
        return "txt"
    case strings.Contains(accept, "application/json"),
        strings.Contains(accept, "text/javascript"):
        return "json"
    }

    return "html"
}

// AcceptLanguage is a single language from the Accept-Language HTTP header.
type AcceptLanguage struct {
    Language string
    Quality  float32
}
// AcceptLanguages is collection of sortable AcceptLanguage instances.
type AcceptLanguages []AcceptLanguage
func (this AcceptLanguages) Len() int           { return len(this) }
func (this AcceptLanguages) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }
func (this AcceptLanguages) Less(i, j int) bool { return this[i].Quality > this[j].Quality }
func (this AcceptLanguages) String() string {
    output := bytes.NewBufferString("")
    for i, language := range this {
        output.WriteString(fmt.Sprintf("%s (%1.1f)", language.Language, language.Quality))
        if i != len(this)-1 {
            output.WriteString(", ")
        }
    }
    return output.String()
}

// ResolveAcceptLanguage returns a sorted list of Accept-Language
// header values.
//
// The results are sorted using the quality defined in the header for each
// language range with the most qualified language range as the first
// element in the slice.
//
// See the HTTP header fields specification
// (http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.4) for more details.
func ResolveAcceptLanguage(req *http.Request) AcceptLanguages {
    header := req.Header.Get("Accept-Language")
    if header == "" {
        return nil
    }

    acceptLanguageHeaderValues := strings.Split(header, ",")
    acceptLanguages := make(AcceptLanguages, len(acceptLanguageHeaderValues))

    for i, languageRange := range acceptLanguageHeaderValues {
        if qualifiedRange := strings.Split(languageRange, ";q="); len(qualifiedRange) == 2 {
            quality, error := strconv.ParseFloat(qualifiedRange[1], 32)
            if error != nil {
                //WARN.Printf("Detected malformed Accept-Language header quality in '%s', assuming quality is 1\n", languageRange)
                acceptLanguages[i] = AcceptLanguage{qualifiedRange[0], 1}
            } else {
                acceptLanguages[i] = AcceptLanguage{qualifiedRange[0], float32(quality)}
            }
        } else {
            acceptLanguages[i] = AcceptLanguage{languageRange, 1}
        }
    }

    sort.Sort(acceptLanguages)
    return acceptLanguages
}
