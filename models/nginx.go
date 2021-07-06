package models

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type Nginx struct {
	Id         int    `db:"id"`
	Aid        int    `db:"aid" form:"aid" binding:"required"`
	Sid        int    `db:"sid" form:"sid" binding:"required"`
	Vpsid      int    `db:"vpsid" form:"vpsid" binding:"required"`
	Domain     string `db:"domain" form:"domain" binding:"required"`
	Ws         string `db:"ws"`
	Single     string `db:"single"`
	CreateTime int64  `db:"create_time"`
	Status     int
}

var NginxFields = []string{
	"aid",
	"sid",
	"vpsid",
	"domain",
	"ws",
	"single",
	"create_time",
	"status",
}

func (n *Nginx) Insert() error {
	flds := NginxFields

	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_nginx (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, n)
	if err == nil {
		lastId, _ := res.LastInsertId()
		n.Id = int(lastId)
	} else {
		log.Println("Nginx Insert Failed, err:", err)
	}
	return err
}

func (n *Nginx) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = NginxFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_nginx SET %s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, n)
		if err != nil {
			log.Println("Nginx Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Nginx Update Need fields")
}

func (n *Nginx) Install() error {
	data := gin.H{"domain": n.Domain}

	var err error
	var rsp *HttpJsonRsp
	rsp, err = HttpPostJsonByVps(n.Vpsid, "nginx/install", data)
	if err == nil {
		if rsp.Status == 0 {
			n.Status = 1
		} else {
			err = fmt.Errorf(rsp.Message)
		}
	} else {
		n.Status = 0
	}
	n.Update("status")
	return err
}

func NginxGetList(filters ...interface{}) ([]Nginx, int) {
	nginxs := []Nginx{}
	count := MysqlGetList("kgo_nginx", &nginxs, filters...)
	if count == 0 {
		count = len(nginxs)
	}
	return nginxs, count
}

// 分页查找
func NginxGetPageList(page, limit int, filters ...interface{}) ([]Nginx, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return NginxGetList(fs...)
}

func NginxGetListByAid(aid int) []Nginx {
	nginxs := []Nginx{}

	sqlStr := fmt.Sprintf("SELECT * FROM kgo_nginx WHERE aid=%d ORDER BY sid", aid)
	err := mysqlDB.Select(&nginxs, sqlStr)
	if err != nil {
		log.Printf("NginxGetListByAid failed, err:%v\n", err, aid)
		return nil
	}

	return nginxs
}

func NginxGetById(id int) *Nginx {
	nginx := &Nginx{}
	err := mysqlDB.Get(nginx, "SELECT * FROM kgo_nginx WHERE id=?", id)
	if err != nil {
		log.Println("NginxGetById failed, err:", err, id)
		return nil
	}
	return nginx
}
