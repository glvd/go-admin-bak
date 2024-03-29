package models

import (
	"github.com/glvd/go-admin/modules/db"
	"github.com/glvd/go-admin/modules/db/dialect"
	"strconv"
	"time"
)

// MenuModel is menu model structure.
type MenuModel struct {
	Base

	Id        int64
	Title     string
	ParentId  int64
	Icon      string
	Uri       string
	Header    string
	CreatedAt string
	UpdatedAt string
}

// Menu return a default menu model.
func Menu() MenuModel {
	return MenuModel{Base: Base{TableName: "adm_menu"}}
}

// MenuWithId return a default menu model of given id.
func MenuWithId(id string) MenuModel {
	idInt, _ := strconv.Atoi(id)
	return MenuModel{Base: Base{TableName: "adm_menu"}, Id: int64(idInt)}
}

func (t MenuModel) SetConn(con db.Connection) MenuModel {
	t.Conn = con
	return t
}

// Find return a default menu model of given id.
func (t MenuModel) Find(id interface{}) MenuModel {
	item, _ := t.Table(t.TableName).Find(id)
	return t.MapToModel(item)
}

// New create a new menu model.
func (t MenuModel) New(title, icon, uri, header string, parentId, order int64) MenuModel {

	id, _ := t.Table(t.TableName).Insert(dialect.H{
		"title":     title,
		"parent_id": parentId,
		"icon":      icon,
		"uri":       uri,
		"order":     order,
		"header":    header,
	})

	t.Id = id
	t.Title = title
	t.ParentId = parentId
	t.Icon = icon
	t.Uri = uri
	t.Header = header

	return t
}

// Delete delete the menu model.
func (t MenuModel) Delete() {
	_ = t.Table(t.TableName).Where("id", "=", t.Id).Delete()
	_ = t.Table("adm_role_menu").Where("menu_id", "=", t.Id).Delete()
	items, _ := t.Table(t.TableName).Where("parent_id", "=", t.Id).All()

	if len(items) > 0 {
		ids := make([]interface{}, len(items))
		for i := 0; i < len(ids); i++ {
			ids[i] = items[i]["id"]
		}
		_ = t.Table("adm_role_menu").WhereIn("menu_id", ids).Delete()
	}

	_ = t.Table(t.TableName).Where("parent_id", "=", t.Id).Delete()
}

// Update update the menu model.
func (t MenuModel) Update(title, icon, uri, header string, parentId int64) MenuModel {

	_, _ = t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(dialect.H{
			"title":      title,
			"parent_id":  parentId,
			"icon":       icon,
			"uri":        uri,
			"header":     header,
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		})

	t.Title = title
	t.ParentId = parentId
	t.Icon = icon
	t.Uri = uri
	t.Header = header

	return t
}

// ResetOrder update the order of menu models.
func (t MenuModel) ResetOrder(data []map[string]interface{}) {
	count := 1
	for _, v := range data {
		if child, ok := v["children"]; ok {
			_, _ = t.Table(t.TableName).
				Where("id", "=", v["id"]).Update(dialect.H{
				"order":     count,
				"parent_id": 0,
			})

			for _, v2 := range child.([]interface{}) {
				_, _ = t.Table(t.TableName).
					Where("id", "=", v2.(map[string]interface{})["id"]).Update(dialect.H{
					"order":     count,
					"parent_id": v["id"],
				})
				count++
			}
		} else {
			_, _ = t.Table(t.TableName).
				Where("id", "=", v["id"]).Update(dialect.H{
				"order":     count,
				"parent_id": 0,
			})
			count++
		}
	}
}

// CheckRole check the role if has permission to get the menu.
func (t MenuModel) CheckRole(roleId string) bool {
	checkRole, _ := t.Table("adm_role_menu").
		Where("role_id", "=", roleId).
		Where("menu_id", "=", t.Id).
		First()
	return checkRole != nil
}

// AddRole add a role to the menu.
func (t MenuModel) AddRole(roleId string) {
	if roleId != "" {
		if !t.CheckRole(roleId) {
			_, _ = t.Table("adm_role_menu").
				Insert(dialect.H{
					"role_id": roleId,
					"menu_id": t.Id,
				})
		}
	}
}

// DeleteRoles delete roles with menu.
func (t MenuModel) DeleteRoles() {
	_ = t.Table("adm_role_menu").
		Where("menu_id", "=", t.Id).
		Delete()
}

// MapToModel get the menu model from given map.
func (t MenuModel) MapToModel(m map[string]interface{}) MenuModel {
	t.Id = m["id"].(int64)
	t.Title, _ = m["title"].(string)
	t.ParentId = m["parent_id"].(int64)
	t.Icon, _ = m["icon"].(string)
	t.Uri, _ = m["uri"].(string)
	t.Header, _ = m["header"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
