package models

import (
	"github.com/glvd/go-admin/modules/db"
	"github.com/glvd/go-admin/modules/db/dialect"
	"strconv"
	"time"
)

// UserModel is user model structure.
type UserModel struct {
	Base `json:"-"`

	Id            int64             `json:"id"`
	Name          string            `json:"name"`
	UserName      string            `json:"user_name"`
	Password      string            `json:"password"`
	Avatar        string            `json:"avatar"`
	RememberToken string            `json:"remember_token"`
	Permissions   []PermissionModel `json:"permissions"`
	MenuIds       []int64           `json:"menu_ids"`
	Roles         []RoleModel       `json:"role"`
	Level         string            `json:"level"`
	LevelName     string            `json:"level_name"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// User return a default user model.
func User() UserModel {
	return UserModel{Base: Base{TableName: "adm_users"}}
}

// UserWithId return a default user model of given id.
func UserWithId(id string) UserModel {
	idInt, _ := strconv.Atoi(id)
	return UserModel{Base: Base{TableName: "adm_users"}, Id: int64(idInt)}
}

func (t UserModel) SetConn(con db.Connection) UserModel {
	t.Conn = con
	return t
}

// Find return a default user model of given id.
func (t UserModel) Find(id interface{}) UserModel {
	item, _ := t.Table(t.TableName).Find(id)
	return t.MapToModel(item)
}

// FindByUserName return a default user model of given name.
func (t UserModel) FindByUserName(username interface{}) UserModel {
	item, _ := t.Table(t.TableName).Where("username", "=", username).First()
	return t.MapToModel(item)
}

// IsEmpty check the user model is empty or not.
func (t UserModel) IsEmpty() bool {
	return t.Id == int64(0)
}

// HasMenu check the user has visitable menu or not.
func (t UserModel) HasMenu() bool {
	return len(t.MenuIds) != 0 || t.IsSuperAdmin()
}

// IsSuperAdmin check the user model is super admin or not.
func (t UserModel) IsSuperAdmin() bool {
	for _, per := range t.Permissions {
		if len(per.HttpPath) > 0 && per.HttpPath[0] == "*" {
			return true
		}
	}
	return false
}

// UpdateAvatar update the avatar of user.
func (t UserModel) ReleaseConn() UserModel {
	t.Conn = nil
	return t
}

// UpdateAvatar update the avatar of user.
func (t UserModel) UpdateAvatar(avatar string) {
	t.Avatar = avatar
}

// WithRoles query the role info of the user.
func (t UserModel) WithRoles() UserModel {
	roleModel, _ := t.Table("adm_role_users").
		LeftJoin("adm_roles", "adm_roles.id", "=", "adm_role_users.role_id").
		Where("user_id", "=", t.Id).
		Select("adm_roles.id", "adm_roles.name", "adm_roles.slug",
			"adm_roles.created_at", "adm_roles.updated_at").
		All()

	for _, role := range roleModel {
		t.Roles = append(t.Roles, Role().MapToModel(role))
	}

	if len(t.Roles) > 0 {
		t.Level = t.Roles[0].Slug
		t.LevelName = t.Roles[0].Name
	}

	return t
}

func (t UserModel) GetAllRoleId() []interface{} {

	var ids = make([]interface{}, len(t.Roles))

	for key, role := range t.Roles {
		ids[key] = role.Id
	}

	return ids
}

// WithPermissions query the permission info of the user.
func (t UserModel) WithPermissions() UserModel {

	var permissions = make([]map[string]interface{}, 0)

	roleIds := t.GetAllRoleId()

	if len(roleIds) > 0 {
		permissions, _ = t.Table("adm_role_permissions").
			LeftJoin("adm_permissions", "adm_permissions.id", "=", "adm_role_permissions.permission_id").
			WhereIn("role_id", roleIds).
			Select("adm_permissions.http_method", "adm_permissions.http_path",
				"adm_permissions.id", "adm_permissions.name", "adm_permissions.slug",
				"adm_permissions.created_at", "adm_permissions.updated_at").
			All()
	}

	userPermissions, _ := t.Table("adm_user_permissions").
		LeftJoin("adm_permissions", "adm_permissions.id", "=", "adm_user_permissions.permission_id").
		Where("user_id", "=", t.Id).
		Select("adm_permissions.http_method", "adm_permissions.http_path",
			"adm_permissions.id", "adm_permissions.name", "adm_permissions.slug",
			"adm_permissions.created_at", "adm_permissions.updated_at").
		All()

	permissions = append(permissions, userPermissions...)

	for i := 0; i < len(permissions); i++ {
		exist := false
		for j := 0; j < len(t.Permissions); j++ {
			if t.Permissions[j].Id == permissions[i]["id"] {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		t.Permissions = append(t.Permissions, Permission().MapToModel(permissions[i]))
	}

	return t
}

// WithMenus query the menu info of the user.
func (t UserModel) WithMenus() UserModel {

	var menuIdsModel []map[string]interface{}

	if t.IsSuperAdmin() {
		menuIdsModel, _ = t.Table("adm_role_menu").
			LeftJoin("adm_menu", "adm_menu.id", "=", "adm_role_menu.menu_id").
			Select("menu_id", "parent_id").
			All()
	} else {
		rolesId := t.GetAllRoleId()
		if len(rolesId) > 0 {
			menuIdsModel, _ = t.Table("adm_role_menu").
				LeftJoin("adm_menu", "adm_menu.id", "=", "adm_role_menu.menu_id").
				WhereIn("adm_role_menu.role_id", rolesId).
				Select("menu_id", "parent_id").
				All()
		}
	}

	var menuIds []int64

	for _, mid := range menuIdsModel {
		if parentId, ok := mid["parent_id"].(int64); ok && parentId != 0 {
			for _, mid2 := range menuIdsModel {
				if mid2["menu_id"].(int64) == mid["parent_id"].(int64) {
					menuIds = append(menuIds, mid["menu_id"].(int64))
					break
				}
			}
		} else {
			menuIds = append(menuIds, mid["menu_id"].(int64))
		}
	}

	t.MenuIds = menuIds
	return t
}

// New create a user model.
func (t UserModel) New(username, password, name, avatar string) UserModel {

	id, _ := t.Table(t.TableName).Insert(dialect.H{
		"username": username,
		"password": password,
		"name":     name,
		"avatar":   avatar,
	})

	t.Id = id
	t.UserName = username
	t.Password = password
	t.Avatar = avatar
	t.Name = name

	return t
}

// Update update the user model.
func (t UserModel) Update(username, password, name, avatar string) UserModel {

	fieldValues := dialect.H{
		"username":   username,
		"name":       name,
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	if avatar != "" {
		fieldValues["avatar"] = avatar
		t.Avatar = avatar
	}

	if password != "" {
		fieldValues["password"] = password
		t.Avatar = avatar
	}

	_, _ = t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(fieldValues)

	t.UserName = username
	t.Password = password
	t.Name = name

	return t
}

// UpdatePwd update the password of the user model.
func (t UserModel) UpdatePwd(password string) UserModel {

	_, _ = t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(dialect.H{
			"password": password,
		})

	t.Password = password
	return t
}

// CheckRole check the role of the user model.
func (t UserModel) CheckRoleId(roleId string) bool {
	checkRole, _ := t.Table("adm_role_users").
		Where("role_id", "=", roleId).
		Where("user_id", "=", t.Id).
		First()
	return checkRole != nil
}

// DeleteRoles delete all the roles of the user model.
func (t UserModel) DeleteRoles() {
	_ = t.Table("adm_role_users").
		Where("user_id", "=", t.Id).
		Delete()
}

// AddRole add a role of the user model.
func (t UserModel) AddRole(roleId string) {
	if roleId != "" {
		if !t.CheckRoleId(roleId) {
			_, _ = t.Table("adm_role_users").
				Insert(dialect.H{
					"role_id": roleId,
					"user_id": t.Id,
				})
		}
	}
}

// CheckRole check the role of the user.
func (t UserModel) CheckRole(slug string) bool {
	for _, role := range t.Roles {
		if role.Slug == slug {
			return true
		}
	}

	return false
}

// CheckPermission check the permission of the user.
func (t UserModel) CheckPermissionById(permissionId string) bool {
	checkPermission, _ := t.Table("adm_user_permissions").
		Where("permission_id", "=", permissionId).
		Where("user_id", "=", t.Id).
		First()
	return checkPermission != nil
}

// CheckPermission check the permission of the user.
func (t UserModel) CheckPermission(permission string) bool {
	for _, per := range t.Permissions {
		if per.Slug == permission {
			return true
		}
	}

	return false
}

// DeletePermissions delete all the permissions of the user model.
func (t UserModel) DeletePermissions() {
	_ = t.Table("adm_user_permissions").
		Where("user_id", "=", t.Id).
		Delete()
}

// AddPermission add a permission of the user model.
func (t UserModel) AddPermission(permissionId string) {
	if permissionId != "" {
		if !t.CheckPermissionById(permissionId) {
			_, _ = t.Table("adm_user_permissions").
				Insert(dialect.H{
					"permission_id": permissionId,
					"user_id":       t.Id,
				})
		}
	}
}

// MapToModel get the user model from given map.
func (t UserModel) MapToModel(m map[string]interface{}) UserModel {
	t.Id, _ = m["id"].(int64)
	t.Name, _ = m["name"].(string)
	t.UserName, _ = m["username"].(string)
	t.Password, _ = m["password"].(string)
	t.Avatar, _ = m["avatar"].(string)
	t.RememberToken, _ = m["remember_token"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
