package models

import (
	"fmt"
	"log"
	"strings"
)

type Menu struct {
	Id         int    `form:"id"`
	Pid        int    `form:"pid"`
	AuthName   string `db:"auth_name" form:"auth_name" binding:"required"`
	AuthUrl    string `db:"auth_url" form:"auth_url"`
	AuthBit    uint   `db:"auth_bit"`
	Sort       int    `form:"sort" binding:"required"`
	Icon       string `form:"icon" binding:"required"`
	IsShow     int    `db:"is_show" form:"is_show"`
	UserId     int    `db:"user_id"`
	CreateId   int    `db:"create_id"`
	UpdateId   int    `db:"update_id"`
	Status     int
	CreateTime int64 `db:"create_time"`
	UpdateTime int64 `db:"update_time"`
}

var MenuFields = []string{
	"pid",
	"auth_name",
	"auth_url",
	"auth_bit",
	"sort",
	"icon",
	"is_show",
	"user_id",
	"create_id",
	"update_id",
	"status",
	"create_time",
	"update_time",
}

func (m *Menu) Insert() error {
	vals := make([]string, 0, len(MenuFields))
	for _, v := range MenuFields {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_menu (%s) VALUES (%s)", strings.Join(MenuFields, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, m)
	if err == nil {
		lastId, _ := res.LastInsertId()
		m.Id = int(lastId)
	} else {
		log.Println("Menu Insert Failed, err:", err)
	}
	return err
}

func (m *Menu) Update() error {
	if len(MenuFields) > 0 {
		flds := make([]string, 0, len(MenuFields))
		for _, v := range MenuFields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_menu SET	%s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, m)
		if err != nil {
			log.Println("Menu Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Menu Update Need fields")
}

func MenuGetList(filters ...interface{}) ([]Menu, int) {
	menus := []Menu{}
	count := MysqlGetList("kgo_menu", &menus, filters...)
	if count == 0 {
		count = len(menus)
	}
	return menus, count
}

func MenuGetById(id int) *Menu {
	menu := &Menu{}
	err := mysqlDB.Get(menu, "SELECT * FROM kgo_menu WHERE id=?", id)
	if err != nil {
		log.Println("MenuGetById failed, err:", err, id)
		return nil
	}
	return menu
}

func MenuGetAuth(roleId int) []Menu {
	if roleId <= 30 && roleId > 0 {
		menus := []Menu{}
		authBit := uint(1) << uint(roleId-1)
		sqlStr := fmt.Sprintf("SELECT * FROM kgo_menu WHERE status=1 AND auth_bit&%d=%d ORDER BY pid,sort", authBit, authBit)
		err := mysqlDB.Select(&menus, sqlStr)
		if err != nil {
			log.Printf("MenuGetAuth failed, err:%v\n", err, roleId)
			return nil
		}
		return menus
	} else {
		return nil
	}
}
