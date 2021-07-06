package models

import (
	"fmt"
	"log"
	"strings"
)

type CDN struct {
	Id         int    `form:"id"`
	Aid        int    `db:"aid" form:"aid" binding:"required"`
	Version    string `db:"version"`
	CreateTime int64  `db:"create_time"`
	InstallLog string `db:"install_log"`
	Status     int
}

var CDNFields = []string{
	"aid",
	"version",
	"create_time",
	"install_log",
	"status",
}

func (c *CDN) Insert() error {
	flds := CDNFields

	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_cdn (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, c)
	if err == nil {
		lastId, _ := res.LastInsertId()
		c.Id = int(lastId)
	} else {
		log.Println("CDN Insert Failed, err:", err)
	}
	return err
}

func (c *CDN) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = CDNFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_cdn SET %s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, c)
		if err != nil {
			log.Println("CDN Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("CDN Update Need fields")
}

func CDNGetList(filters ...interface{}) ([]CDN, int) {
	cdns := []CDN{}
	count := MysqlGetList("kgo_cdn", &cdns, filters...)
	if count == 0 {
		count = len(cdns)
	}
	return cdns, count
}

// 分页查找
func CDNGetPageList(page, limit int, filters ...interface{}) ([]CDN, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return CDNGetList(fs...)
}

func CDNGetById(id int) *CDN {
	cdn := &CDN{}
	err := mysqlDB.Get(cdn, "SELECT * FROM kgo_cdn WHERE id=?", id)
	if err != nil {
		log.Println("CDNGetById failed, err:", err, id)
		return nil
	}
	return cdn
}

func CDNGetByVersion(version string) *CDN {
	cdn := &CDN{}
	err := mysqlDB.Get(cdn, "SELECT * FROM kgo_cdn WHERE version=?", version)
	if err != nil {
		log.Println("CDNGetByVersion failed, err:", err, version)
		return nil
	}
	return cdn
}

func CDNGetByTag(tag string) *CDN {
	cdn := &CDN{}
	err := mysqlDB.Get(cdn, "SELECT * FROM kgo_cdn WHERE version=?", tag)
	if err != nil {
		log.Println("CDNGetByTag failed, err:", err, tag)
		return nil
	}
	return cdn
}

func CDNGetCount() int {
	var count int
	err := mysqlDB.Get(&count, "SELECT count(*) FROM kgo_cdn;")
	if err != nil {
		log.Println("CDNGetCount failed, err:", err.Error())
	}
	return count
}

func CDNDelete() {
	_, err := mysqlDB.Exec("DELETE FROM kgo_cdn WHERE status=3")
	if err != nil {
		log.Println("CDNDelete failed, err:", err.Error())
	}
}
