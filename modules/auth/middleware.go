// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/glvd/go-admin/context"
	"github.com/glvd/go-admin/modules/config"
	"github.com/glvd/go-admin/modules/db"
	"github.com/glvd/go-admin/modules/language"
	"github.com/glvd/go-admin/modules/logger"
	"github.com/glvd/go-admin/modules/page"
	"github.com/glvd/go-admin/plugins/admin/models"
	template2 "github.com/glvd/go-admin/template"
	"github.com/glvd/go-admin/template/types"
	"html/template"
	"regexp"
	"strings"
)

// Invoker contains the callback functions which are used
// in the route middleware.
type Invoker struct {
	prefix                 string
	authFailCallback       MiddlewareCallback
	permissionDenyCallback MiddlewareCallback
	conn                   db.Connection
}

// Middleware is the default auth middleware of plugins.
func Middleware(conn db.Connection) context.Handler {
	return DefaultInvoker(conn).Middleware()
}

// DefaultInvoker return a default Invoker.
func DefaultInvoker(conn db.Connection) *Invoker {
	return &Invoker{
		prefix: config.Get().Prefix(),
		authFailCallback: func(ctx *context.Context) {
			ctx.Write(302, map[string]string{
				"Location": config.Get().Url("/login"),
			}, ``)
		},
		permissionDenyCallback: func(ctx *context.Context) {
			page.SetPageContent(ctx, Auth(ctx), func(ctx interface{}) (types.Panel, error) {
				alert := template2.Get(config.Get().Theme).Alert().
					SetTitle(template.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
					SetTheme("warning").SetContent(template.HTML("permission denied")).GetContent()

				return types.Panel{
					Content:     alert,
					Description: "Error",
					Title:       "Error",
				}, nil
			}, conn)
		},
		conn: conn,
	}
}

// SetPrefix return the default Invoker with the given prefix.
func SetPrefix(prefix string, conn db.Connection) *Invoker {
	i := DefaultInvoker(conn)
	i.prefix = prefix
	return i
}

// SetAuthFailCallback set the authFailCallback of Invoker.
func (invoker *Invoker) SetAuthFailCallback(callback MiddlewareCallback) *Invoker {
	invoker.authFailCallback = callback
	return invoker
}

// SetPermissionDenyCallback set the permissionDenyCallback of Invoker.
func (invoker *Invoker) SetPermissionDenyCallback(callback MiddlewareCallback) *Invoker {
	invoker.permissionDenyCallback = callback
	return invoker
}

// MiddlewareCallback is type of callback function.
type MiddlewareCallback func(ctx *context.Context)

// Middleware get the auth middleware from Invoker.
func (invoker *Invoker) Middleware() context.Handler {
	return func(ctx *context.Context) {
		user, authOk, permissionOk := Filter(ctx, invoker.conn)

		if authOk && permissionOk {
			ctx.SetUserValue("user", user)
			ctx.Next()
			return
		}

		if !authOk {
			invoker.authFailCallback(ctx)
			ctx.Abort()
			return
		}

		if !permissionOk {
			ctx.SetUserValue("user", user)
			invoker.permissionDenyCallback(ctx)
			ctx.Abort()
			return
		}
	}
}

// Filter retrieve the user model from Context and check the permission
// at the same time.
func Filter(ctx *context.Context, conn db.Connection) (models.UserModel, bool, bool) {
	var (
		id   float64
		ok   bool
		user = models.User()
	)

	if id, ok = InitSession(ctx, conn).Get("user_id").(float64); !ok {
		return user, false, false
	}

	user, ok = GetCurUserByID(int64(id), conn)

	if !ok {
		return user, false, false
	}

	return user, true, CheckPermissions(user, ctx.Request.URL.String(), ctx.Method())
}

const defaultUserIDSesKey = "user_id"

// GetUserID return the user id from the session.
func GetUserID(sesKey string, conn db.Connection) int64 {
	id := GetSessionByKey(sesKey, defaultUserIDSesKey, conn)
	if idFloat64, ok := id.(float64); ok {
		return int64(idFloat64)
	}
	return -1
}

// GetCurUser return the user model.
func GetCurUser(sesKey string, conn db.Connection) (user models.UserModel, ok bool) {

	if sesKey == "" {
		ok = false
		return
	}

	id := GetUserID(sesKey, conn)
	if id == -1 {
		ok = false
		return
	}
	return GetCurUserByID(id, conn)
}

// GetCurUserByID return the user model of given user id.
func GetCurUserByID(id int64, conn db.Connection) (user models.UserModel, ok bool) {

	user = models.User().SetConn(conn).Find(id)

	if user.IsEmpty() {
		ok = false
		return
	}

	if user.Avatar == "" || config.Get().Store.Prefix == "" {
		user.Avatar = ""
	} else {
		user.Avatar = "/" + config.Get().Store.Prefix + "/" + user.Avatar
	}

	user = user.WithRoles().WithPermissions().WithMenus()

	ok = user.HasMenu()

	return
}

// CheckPermissions check the permission of the user.
func CheckPermissions(user models.UserModel, path string, method string) bool {

	logoutCheck, _ := regexp.Compile(config.Get().Url("/logout") + "(.*?)")

	if logoutCheck.MatchString(path) {
		return true
	}

	if path != "/" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	pathArr := strings.Split(path, "?")

	for _, v := range user.Permissions {

		if v.HttpMethod[0] == "" || inMethodArr(v.HttpMethod, method) {

			if v.HttpPath[0] == "*" {
				return true
			}

			for i := 0; i < len(v.HttpPath); i++ {

				matchPath := config.Get().Url(strings.TrimSpace(v.HttpPath[i]))

				if len(pathArr) > 1 {
					if pathArr[0] == matchPath && !strings.Contains(matchPath, "?") {
						matchPath += "(.*)"
					} else if strings.Contains(matchPath, "?id=") && !strings.Contains(matchPath, "(.*)") {
						matchPath = strings.Replace(matchPath, "?", "(.*)", -1) + "(.*)"
					}
				}

				if matchPath == path {
					return true
				}

				reg, err := regexp.Compile(matchPath)

				if err != nil {
					logger.Error("CheckPermissions error: ", err)
					continue
				}

				if reg.FindString(path) == path {
					return true
				}
			}
		}
	}

	return false
}

func inMethodArr(arr []string, str string) bool {
	for i := 0; i < len(arr); i++ {
		if strings.EqualFold(arr[i], str) {
			return true
		}
	}
	return false
}
