package web

import (
	"fmt"
	"html"
	"strings"
)

// PageMeta holds HTML head metadata for the operator console.
type PageMeta struct {
	Title       string
	Description string
	ThemeColor  string
}

// DefaultPageMeta returns UI metadata.
func DefaultPageMeta() PageMeta {
	return PageMeta{
		Title:       "dockload — 吊具称重操作台",
		Description: "集装箱码头吊具称重流监控与标定",
		ThemeColor:  "#0b3d5c",
	}
}

// RenderHead returns HTML head fragment.
func RenderHead(meta PageMeta) string {
	var b strings.Builder
	b.WriteString("<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString(fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", html.EscapeString(meta.Description)))
	b.WriteString(fmt.Sprintf("<meta name=\"theme-color\" content=\"%s\">\n", html.EscapeString(meta.ThemeColor)))
	b.WriteString(fmt.Sprintf("<title>%s</title>\n", html.EscapeString(meta.Title)))
	b.WriteString("<link rel=\"stylesheet\" href=\"/style.css\">\n")
	b.WriteString("</head>\n")
	return b.String()
}

// FormatKg formats weight for HTML display.
func FormatKg(kg float64) string {
	return fmt.Sprintf("%.2f kg", kg)
}

// FormatCounts formats raw counts for display.
func FormatCounts(raw int64) string {
	return fmt.Sprintf("%d counts", raw)
}

// NavItems returns primary navigation labels.
func NavItems() []string {
	return []string{"标定", "最近称重", "吊具状态", "模拟上报"}
}

// ScriptTags returns script includes for page bottom.
func ScriptTags() string {
	return "<script src=\"/app.js\"></script>\n"
}

// PanelID returns DOM id for nav panel index.
func PanelID(index int) string {
	return fmt.Sprintf("panel-%d", index)
}

// EscapeJS escapes string for embedding in JavaScript literals.
func EscapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// BuildHookOption renders select option HTML.
func BuildHookOption(hookID string, selected bool) string {
	sel := ""
	if selected {
		sel = " selected"
	}
	return fmt.Sprintf("<option value=\"%s\"%s>%s</option>", html.EscapeString(hookID), sel, html.EscapeString(hookID))
}

// TableHeader renders thead row from columns.
func TableHeader(cols ...string) string {
	var b strings.Builder
	b.WriteString("<thead><tr>")
	for _, c := range cols {
		b.WriteString("<th>")
		b.WriteString(html.EscapeString(c))
		b.WriteString("</th>")
	}
	b.WriteString("</tr></thead>")
	return b.String()
}

// EmptyRow renders colspan empty table row.
func EmptyRow(cols int, message string) string {
	return fmt.Sprintf("<tr><td colspan=\"%d\" class=\"empty\">%s</td></tr>", cols, html.EscapeString(message))
}

// BadgeClass returns CSS class for FSM state badge.
func BadgeClass(state string) string {
	switch state {
	case "Stable", "Published":
		return "badge badge-ok"
	case "Collecting":
		return "badge badge-warn"
	default:
		return "badge"
	}
}

// RejectLabel maps API reject codes to Chinese labels.
func RejectLabel(code string) string {
	switch code {
	case "NOT_CALIBRATED":
		return "未标定"
	case "OVER_RANGE":
		return "超量程"
	case "NOT_STABLE":
		return "未稳定"
	default:
		return code
	}
}

// PollIntervalMs default UI refresh interval.
const PollIntervalMs = 3000

// APIPaths documents frontend fetch targets.
var APIPaths = struct {
	CalibList  string
	Recent     string
	HookStatus string
	Raw        string
	CalibPut   string
}{
	CalibList:  "/v1/calib",
	Recent:     "/v1/weighs/recent",
	HookStatus: "/v1/hooks/status",
	Raw:        "/v1/sensors/raw",
	CalibPut:   "/v1/calib/",
}
