package models

import (
	"fmt"
	"log"
	"strings"
)

type User struct {
	Id         int    `form:"id"`
	RoleId     int    `db:"role_id" form:"role_id"`
	LoginName  string `db:"login_name" form:"login_name" binding:"required"`
	RealName   string `db:"real_name" form:"real_name" binding:"required"`
	Password   string
	Salt       string
	Phone      string `form:"phone" binding:"required"`
	Email      string `form:"email" binding:"required"`
	LastLogin  int64  `db:"last_login"`
	LastIp     string `db:"last_ip"`
	Status     int
	CreateId   int   `db:"create_id"`
	UpdateId   int   `db:"update_id"`
	CreateTime int64 `db:"create_time"`
	UpdateTime int64 `db:"update_time"`
}

var UserFields = []string{
	"role_id",
	"login_name",
	"real_name",
	"password",
	"salt",
	"phone",
	"email",
	"last_login",
	"last_ip",
	"status",
	"create_id",
	"update_id",
	"create_time",
	"update_time",
}

func (u *User) Insert() error {
	flds := UserFields

	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_user (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, u)
	if err == nil {
		lastId, _ := res.LastInsertId()
		u.Id = int(lastId)
	} else {
		log.Println("User Insert Failed, err:", err)
	}
	return err
}

func (u *User) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = UserFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_user SET	%s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, u)
		if err != nil {
			log.Println("User Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("User Update Need fields")
}

func UserGetList(filters ...interface{}) ([]User, int) {
	users := []User{}
	count := MysqlGetList("kgo_user", &users, filters...)
	if count == 0 {
		count = len(users)
	}
	return users, count
}

// 分页查找
func UserGetPageList(page, limit int, filters ...interface{}) ([]User, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return UserGetList(fs...)
}

func UserGetByName(name string) *User {
	user := &User{}
	err := mysqlDB.Get(user, "SELECT * FROM kgo_user WHERE login_name=?", name)
	if err != nil {
		log.Println("UserGetByName failed, err:", err, name)
		return nil
	}
	return user
}

func UserGetById(id int) *User {
	user := &User{}
	err := mysqlDB.Get(user, "SELECT * FROM kgo_user WHERE id=?", id)
	if err != nil {
		log.Println("UserGetById failed, err:", err, id)
		return nil
	}
	return user
}
