package components

import (
	"github.com/glvd/go-admin/plugins/admin/modules"
	"github.com/glvd/go-admin/template/types"
	"html/template"
)

type ImgAttribute struct {
	Name     string
	Width    string
	Height   string
	Uuid     string
	HasModal bool
	Src      template.HTML
	types.Attribute
}

func (compo *ImgAttribute) SetWidth(value string) types.ImgAttribute {
	compo.Width = value
	return compo
}

func (compo *ImgAttribute) SetHeight(value string) types.ImgAttribute {
	compo.Height = value
	return compo
}

func (compo *ImgAttribute) WithModal() types.ImgAttribute {
	compo.HasModal = true
	compo.Uuid = modules.Uuid()
	return compo
}

func (compo *ImgAttribute) SetSrc(value template.HTML) types.ImgAttribute {
	compo.Src = value
	return compo
}

func (compo *ImgAttribute) GetContent() template.HTML {
	return ComposeHtml(compo.TemplateList, *compo, "image")
}
