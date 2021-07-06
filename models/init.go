package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oam-center/conf"
	"oam-center/libs"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var mysqlDB = &sqlx.DB{}

type HttpJsonRsp struct {
	Status  int             `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

var StartTime int64

func Init() {
	StartTime = time.Now().Unix()
	cnf := conf.YamlConf.SqlCnf
	dbPort := cnf.Port
	if dbPort == "" {
		dbPort = "3306"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", cnf.User, cnf.Password, cnf.Host, dbPort, cnf.Name)
	log.Println(dsn)

	var err error
	mysqlDB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal("mysql 连接失败", err.Error())
	}

	go CheckVpsTimer()
}

func CheckVpsTimer() {
	tick := time.Tick(3 * time.Hour)
	for {
		select {
		case <-tick:
			checkVpsOnline()
		}
	}
}

func checkVpsOnline() {
	vpss, _ := VpsGetList()
	for _, vps := range vpss {
		go PingVps(&vps)
	}
}

func PingVps(v *Vps) error {
	rsp, err := HttpPostJsonByVps(v.Id, "ping", nil)
	if err == nil {
		v.Status = 1
		if len(rsp.Message) > 0 {
			v.VpsTime = libs.ParseTime(rsp.Message)
		}
	} else {
		v.Status = 0
	}
	v.Update("status", "vps_time")
	return err
}

func HttpPostForm(urlStr string, data url.Values) (*HttpJsonRsp, error) {
	resp, err := http.PostForm(urlStr, data)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rsp := &HttpJsonRsp{}
	err = json.Unmarshal(body, rsp)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func HttpPostJson(urlStr string, data interface{}) (*HttpJsonRsp, error) {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	//log.Println(string(bytesData))
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", urlStr, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	//log.Println(string(respBytes))
	if err != nil {
		return nil, err
	}

	rsp := &HttpJsonRsp{}
	err = json.Unmarshal(respBytes, rsp)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func HttpPostFormByVps(vpsId int, api string, data url.Values) (*HttpJsonRsp, error) {
	vpsDomain := VpsGetDomain(vpsId)
	if len(vpsDomain) == 0 {
		return nil, fmt.Errorf("vps 域名错误")
	}

	apiUrl := fmt.Sprintf("%s/%s", vpsDomain, api)
	return HttpPostForm(apiUrl, data)
}

func HttpPostJsonByVps(vpsId int, api string, data interface{}) (*HttpJsonRsp, error) {
	vpsDomain := VpsGetDomain(vpsId)
	if len(vpsDomain) == 0 {
		return nil, fmt.Errorf("vps 域名错误")
	}

	apiUrl := fmt.Sprintf("%s/%s", vpsDomain, api)
	return HttpPostJson(apiUrl, data)
}

func HttpGetByVps(vpsId int, api string, params *url.Values) (*HttpJsonRsp, error) {
	vpsDomain := VpsGetDomain(vpsId)
	if len(vpsDomain) == 0 {
		return nil, fmt.Errorf("vps 域名错误")
	}

	apiUrl := fmt.Sprintf("%s/%s", vpsDomain, api)
	paramstr := ""
	if params != nil {
		paramstr = params.Encode()
	}
	urlObj, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	urlObj.RawQuery = paramstr

	res, err := http.Get(urlObj.String())
	if err != nil {
		return nil, err
	}

	rsp, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	rspJson := &HttpJsonRsp{}
	err = json.Unmarshal(rsp, rspJson)
	return rspJson, err
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// 分页查找范围条目总数，全部查询返回0
func MysqlGetList(table string, des interface{}, filters ...interface{}) int {
	page, pageSize := 0, 0

	where := ""
	orderBy := ""
	num := len(filters) / 2
	vals := make([]interface{}, 0, num)
	if len(filters) > 0 {
		exWhere := ""
		flds := make([]string, 0, num)
		for i := 0; i < len(filters); i += 2 {
			field := filters[i].(string)
			if field == "_page" {
				page = filters[i+1].(int)
			} else if field == "_limit" {
				pageSize = filters[i+1].(int)
			} else if field == "_order" {
				orderBy = "ORDER BY " + filters[i+1].(string)
			} else if field == "_where" {
				exWhere = filters[i+1].(string)
			} else {
				flds = append(flds, fmt.Sprintf("%s=?", field))
				vals = append(vals, filters[i+1])
			}
		}

		if len(flds) > 0 {
			where += ("WHERE " + strings.Join(flds, " AND "))
		}

		if len(exWhere) > 0 {
			if len(where) > 0 {
				where += (" AND " + exWhere)
			} else {
				where += "WHERE " + exWhere
			}
		}
	}

	var count int = 0
	if page > 0 && pageSize > 0 {
		// 分页查找
		offset := (page - 1) * pageSize

		sqlStr := fmt.Sprintf("SELECT count(*) FROM %s %s %s", table, where, orderBy)
		err := mysqlDB.Get(&count, sqlStr, vals...)
		if err != nil {
			log.Printf("MysqlGetList [Table=%s] Count failed, err:%s\n", table, err.Error())
		}
		if count <= 0 {
			return 0
		}

		sqlStr = fmt.Sprintf("SELECT * FROM %s %s LIMIT %d,%d", table, where, offset, pageSize)
		err = mysqlDB.Select(des, sqlStr, vals...)
		if err != nil {
			log.Printf("MysqlGetList [Table=%s] Page failed, err:%s\n", table, err.Error())
			return 0
		}
	} else {
		sqlStr := fmt.Sprintf("SELECT * FROM %s %s %s", table, where, orderBy)
		err := mysqlDB.Select(des, sqlStr, vals...)
		if err != nil {
			log.Printf("MysqlGetList [Table=%s] failed, err:%s\n", table, err.Error())
			return 0
		}
	}

	return count
}
