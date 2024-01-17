package gasp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tonysoft.com/gasp/resources"
)

func GenerateView(html string, variables map[string]string, handledEvents []string, useTls bool) string {
	html = insertStyling(html)
	html = insertScript(html, handledEvents, useTls)
	html = replaceNowVariable(html)
	html = replaceVariables(html, variables)
	return html
}

func insertStyling(html string) string {
	return strings.ReplaceAll(html, "<!--gasp_css-->", "<style>"+resources.GaspStyle+"</style>")
}

func insertScript(html string, handledEvents []string, useTls bool) string {
	if len(handledEvents) == 0 {
		return strings.ReplaceAll(html, "<!--gasp_js-->", "<script>"+resources.GaspScript+"</script>")
	}

	handlers := ""
	for _, handler := range handledEvents {
		view := strings.Split(handler, "#")[0]
		eventType := strings.Split(handler, "!")[1]
		id := strings.ReplaceAll(strings.ReplaceAll(handler, view+"#", ""), "!"+eventType, "")
		handlers += fmt.Sprintf("this.addControlEventHandler('%s', '%s');\n\t\t", id, eventType)
	}

	script := strings.ReplaceAll(resources.GaspScript, "/*event_handlers*/", handlers)
	script = strings.ReplaceAll(script, "/*tls_override*/", "useTls = "+strconv.FormatBool(useTls)+";")
	return strings.ReplaceAll(html, "<!--gasp_js-->", "<script>"+script+"</script>")
}

func replaceVariables(html string, variables map[string]string) string {
	if variables == nil {
		return html
	}
	for k, v := range variables {
		html = strings.ReplaceAll(html, "<!--"+k+"-->", v)
	}
	return html
}

func replaceNowVariable(html string) string {
	now := time.Now().UTC()
	nowRegex := regexp.MustCompile(`<!--now:([\w\d\-/:,_ ])+-->`)
	for {
		nowPlaceholder := nowRegex.FindString(html)
		if nowPlaceholder != "" {
			format := strings.ReplaceAll(nowPlaceholder, "<!--now:", "")
			format = strings.ReplaceAll(format, "-->", "")
			html = strings.ReplaceAll(html, nowPlaceholder, now.Format(format))
		} else {
			html = strings.ReplaceAll(html, "<!--now-->", now.String())
			return html
		}
	}
}
