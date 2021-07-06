package models

import (
	"fmt"
	"log"
	"strings"
)

const (
	VpsGame uint = 1
	VpsCDN       = 1 << (iota)
	VpsNginx
	VpsCenter
	VpsCross
)

type Vps struct {
	Id         int    `form:"id"`
	Ip         string `db:"ip" form:"ip"`
	Type       uint   `db:"type" form:"type"`
	Domain     string `db:"domain" form:"domain"`
	Detail     string `db:"detail" form:"detail"`
	CreateTime int64  `db:"create_time"`
	Status     int    `db:"status"`
	CtlVersion string `db:"version"`
	VpsTime    int64  `db:"vps_time"`
}

var VpsFields = []string{
	"ip",
	"type",
	"domain",
	"detail",
	"create_time",
	"status",
	"version",
	"vps_time",
}

func (v *Vps) Insert() error {
	flds := VpsFields

	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_vps (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, v)
	if err == nil {
		lastId, _ := res.LastInsertId()
		v.Id = int(lastId)
	} else {
		log.Println("Vps Insert Failed, err:", err)
	}
	return err
}

func (v *Vps) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = VpsFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_vps SET	%s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, v)
		if err != nil {
			log.Println("Vps Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Vps Update Need fields")
}

func (v *Vps) CheckType(ty uint) bool {
	if v.Type&ty == ty {
		return true
	}
	return false
}

func VpsGetList(filters ...interface{}) ([]Vps, int) {
	vpss := []Vps{}
	count := MysqlGetList("kgo_vps", &vpss, filters...)
	if count == 0 {
		count = len(vpss)
	}
	return vpss, count
}

// 分页查找
func VpsGetPageList(page, limit int, filters ...interface{}) ([]Vps, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return VpsGetList(fs...)
}

func VpsGetById(id int) *Vps {
	vps := &Vps{}
	err := mysqlDB.Get(vps, "SELECT * FROM kgo_vps WHERE id=?", id)
	if err != nil {
		log.Println("VpsGetById failed, err:", err, id)
		return nil
	}
	return vps
}

func VpsGetByIp(ip string) *Vps {
	vps := &Vps{}
	err := mysqlDB.Get(vps, "SELECT * FROM kgo_vps WHERE ip=?", ip)
	if err != nil {
		log.Println("VpsGetByIp failed, err:", err, ip)
		return nil
	}
	return vps
}

func VpsGetDomain(id int) string {
	vps := VpsGetById(id)
	if vps == nil {
		return ""
	} else {
		domain := vps.Domain
		if len(domain) == 0 {
			// 被控制端默认端口 7700
			domain = fmt.Sprintf("%s:%d", vps.Ip, 7700)
		}
		return domain
	}
}
