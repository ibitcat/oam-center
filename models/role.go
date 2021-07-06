package models

import (
	"fmt"
	"log"
	"strings"
)

type Role struct {
	Id         int    `form:"id"`
	RoleName   string `db:"role_name" form:"role_name" binding:"required"`
	Detail     string `form:"detail" binding:"required"`
	CreateId   int    `db:"create_id"`
	UpdateId   int    `db:"update_id"`
	Status     int
	CreateTime int64 `db:"create_time"`
	UpdateTime int64 `db:"update_time"`
}

var RoleFields = []string{
	"role_name",
	"detail",
	"create_id",
	"update_id",
	"status",
	"create_time",
	"update_time",
}

func (r *Role) Insert() error {
	vals := make([]string, 0, len(RoleFields))
	for _, v := range RoleFields {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_role (%s) VALUES (%s)", strings.Join(RoleFields, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, r)
	if err == nil {
		lastId, _ := res.LastInsertId()
		r.Id = int(lastId)
	} else {
		log.Println("Role Insert Failed, err:", err)
	}
	return err
}

func (r *Role) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = RoleFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_role SET	%s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, r)
		if err != nil {
			log.Println("Role Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Role Update Need fields")
}

func RoleGetList(filters ...interface{}) ([]Role, int) {
	roles := []Role{}
	count := MysqlGetList("kgo_role", &roles, filters...)
	if count == 0 {
		count = len(roles)
	}
	return roles, count
}

// 分页查找
func RoleGetPageList(page, limit int, filters ...interface{}) ([]Role, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return RoleGetList(fs...)
}

func RoleGetByName(name string) *Role {
	role := &Role{}
	err := mysqlDB.Get(role, "SELECT * FROM kgo_role WHERE role_name=?", name)
	if err != nil {
		log.Println("RoleGetByName failed, err:", err, name)
		return nil
	}
	return role
}

func RoleGetById(id int) *Role {
	role := &Role{}
	err := mysqlDB.Get(role, "SELECT * FROM kgo_role WHERE id=?", id)
	if err != nil {
		log.Println("RoleGetById failed, err:", err, id)
		return nil
	}
	return role
}
