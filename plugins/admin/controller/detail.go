package controller

import (
	"fmt"
	"github.com/glvd/go-admin/context"
	"github.com/glvd/go-admin/modules/auth"
	config2 "github.com/glvd/go-admin/modules/config"
	"github.com/glvd/go-admin/modules/language"
	"github.com/glvd/go-admin/modules/menu"
	"github.com/glvd/go-admin/plugins/admin/modules"
	"github.com/glvd/go-admin/plugins/admin/modules/parameter"
	"github.com/glvd/go-admin/plugins/admin/modules/table"
	"github.com/glvd/go-admin/template"
	"github.com/glvd/go-admin/template/types"
	"github.com/glvd/go-admin/template/types/form"
	template2 "html/template"
	"net/http"
)

func ShowDetail(ctx *context.Context) {
	prefix := ctx.Query("__prefix")
	id := ctx.Query("__goadmin_detail_pk")
	panel := table.Get(prefix)
	user := auth.Auth(ctx)

	newPanel := panel.Copy()

	formModel := newPanel.GetForm()

	formModel.FieldList = make([]types.FormField, len(panel.GetInfo().FieldList))

	for i, field := range panel.GetInfo().FieldList {
		formModel.FieldList[i] = types.FormField{
			Field:        field.Field,
			TypeName:     field.TypeName,
			Head:         field.Head,
			FormType:     form.Default,
			FieldDisplay: field.FieldDisplay,
		}
	}

	formData, _, _, _, _, err := newPanel.GetDataFromDatabaseWithId(id)

	var alert template2.HTML

	if err != nil && alert == "" {
		alert = aAlert().SetTitle(template2.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
			SetTheme("warning").
			SetContent(template2.HTML(err.Error())).
			GetContent()
	}

	params := parameter.GetParam(ctx.Request.URL.Query(), panel.GetInfo().DefaultPageSize, panel.GetPrimaryKey().Name,
		panel.GetInfo().GetSort())

	editUrl := modules.AorB(panel.GetEditable(), config.Url("/info/"+prefix+"/edit"+params.GetRouteParamStr()), "")
	deleteUrl := modules.AorB(panel.GetDeletable(), config.Url("/delete/"+prefix), "")
	infoUrl := config2.Get().Url("/info/" + prefix + params.GetRouteParamStr())

	deleteJs := ""

	if deleteUrl != "" {
		deleteJs = fmt.Sprintf(`<script>
function DeletePost(id) {
	swal({
			title: '%s',
			type: "warning",
			showCancelButton: true,
			confirmButtonColor: "#DD6B55",
			confirmButtonText: '%s',
			closeOnConfirm: false,
			cancelButtonText: '%s',
		},
		function () {
			$.ajax({
				method: 'post',
				url: '%s',
				data: {
					id: id
				},
				success: function (data) {
					if (typeof (data) === "string") {
						data = JSON.parse(data);
					}
					if (data.code === 200) {
						location.href = '%s'
					} else {
						swal(data.msg, '', 'error');
					}
				}
			});
		});
}

$('.delete-btn').on('click', function (event) {
	DeletePost(%s)
});

</script>`, language.Get("are you sure to delete"), language.Get("yes"), language.Get("cancel"), deleteUrl, infoUrl, id)
	}

	title := language.Get("Detail")

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content: alert + detailContent(aForm().
			SetTitle(template.HTML(title)).
			SetContent(formData).
			SetFooter(template.HTML(deleteJs)).
			SetInfoUrl(infoUrl).
			SetPrefix(config.PrefixFixSlash()), editUrl, deleteUrl),
		Description: title,
		Title:       title,
	}, config, menu.GetGlobalMenu(user, conn).SetActiveClass(config.URLRemovePrefix(ctx.Path())))

	ctx.HTML(http.StatusOK, buf.String())
}
