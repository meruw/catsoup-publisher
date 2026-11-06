package uploads

import (
	_ "embed"
	"encoding/base64"
	"html/template"
	"net/http"
)

//go:embed browser_return.html
var browserReturnHTML string

var browserReturnPage = template.Must(template.New("return").Parse(browserReturnHTML))

func (c *Client) writeBrowserReturn(w http.ResponseWriter, platform string, status int, title, detail string) {
	names := map[string]string{"instagram": "Instagram", "tiktok": "TikTok", "youtube": "YouTube", "x": "X"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = browserReturnPage.Execute(w, struct{ Provider, Title, Detail, Icon string }{names[platform], title, detail, base64.StdEncoding.EncodeToString(c.BrowserIcon)})
}
