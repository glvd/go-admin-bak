package components

import (
	"bytes"
	"fmt"
	"github.com/glvd/go-admin/modules/language"
	"html/template"
	"strings"
)

func ComposeHtml(temList map[string]string, compo interface{}, templateName ...string) template.HTML {
	var text = ""
	for _, v := range templateName {
		text += temList["components/"+v]
	}

	tmpl, err := template.New("comp").Funcs(template.FuncMap{
		"lang":     language.Get,
		"langHtml": language.GetFromHtml,
		"link": func(cdnUrl, prefixUrl, assetsUrl string) string {
			if cdnUrl == "" {
				return prefixUrl + assetsUrl
			}
			return cdnUrl + assetsUrl
		},
		"isLinkUrl": func(s string) bool {
			return (len(s) > 7 && s[:7] == "http://") || (len(s) > 8 && s[:8] == "https://")
		},
	}).Parse(text)
	if err != nil {
		panic("ComposeHtml Error:" + err.Error())
	}
	buffer := new(bytes.Buffer)

	defineName := strings.Replace(templateName[0], "table/", "", -1)
	defineName = strings.Replace(defineName, "form/", "", -1)

	err = tmpl.ExecuteTemplate(buffer, defineName, compo)
	if err != nil {
		fmt.Println("ComposeHtml Error:", err)
	}
	return template.HTML(buffer.String())
}
