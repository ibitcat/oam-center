package router

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"path"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

const UpdateZipDir = "/data/update_zip"

var versionRegexp = regexp.MustCompile(`\w+?_(\w+?)_(\d+)$`)

func VersionIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "版本列表",
	}
	c.HTML(200, "versionIndex", data)
}

func VersionTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	versionFlag := c.DefaultQuery("version_flag", "")

	startIdx := (page - 1) * limit
	stopIdx := page * limit

	amout := 0
	list := make([]gin.H, 0)
	if len(versionFlag) > 0 {
		f, _ := os.Stat(UpdateZipDir + "/" + versionFlag)
		if f != nil && f.IsDir() {
			s := versionRegexp.FindString(f.Name())
			if len(s) > 0 {
				row := make(gin.H)
				zips, sizes := GetVersionZips(f.Name())
				row["id"] = 1
				row["flag"] = f.Name()
				row["zips"] = template.HTML(strings.Join(zips, "<br>"))
				row["sizes"] = template.HTML(strings.Join(sizes, "<br>"))
				row["create_time"] = f.ModTime().Format("2006-01-02 15:04:05")
				list = append(list, row)
				amout++
			}
		}
	} else {
		files, err := ioutil.ReadDir(UpdateZipDir)
		if err == nil {
			for _, f := range files {
				if f.IsDir() {
					s := versionRegexp.FindString(f.Name())
					if len(s) > 0 {
						if amout >= startIdx && amout < stopIdx {
							row := make(gin.H)
							zips, sizes := GetVersionZips(f.Name())
							row["id"] = amout + 1
							row["flag"] = f.Name()
							row["zips"] = template.HTML(strings.Join(zips, "<br>"))
							row["sizes"] = template.HTML(strings.Join(sizes, "<br>"))
							row["create_time"] = f.ModTime().Format("2006-01-02 15:04:05")
							list = append(list, row)
						}
						amout++
					}
				}
			}
		}
	}
	c.JSON(200, gin.H{"code": 0, "count": amout, "data": list})
}

// 获取版本标签下的所有zip文件
func GetVersionZips(tag string) ([]string, []string) {
	dir := fmt.Sprintf("%s/%s", UpdateZipDir, tag)
	files, err := ioutil.ReadDir(dir)
	if err == nil {
		zips := make([]string, 0)
		sizes := make([]string, 0)
		for _, f := range files {
			if !f.IsDir() {
				zipName := f.Name()
				if strings.HasSuffix(zipName, ".zip") {
					zips = append(zips, zipName)
					sizes = append(sizes, fmt.Sprintf("%d KB", f.Size()/1024))
				}
			}
		}
		return zips, sizes
	}
	return nil, nil
}

func VersionHotPatch(c *gin.Context) {
	tag := c.PostForm("tag")

	s := strings.Split(tag, "_")
	if len(s) == 3 {
		agent := models.AgentGetByFlag(s[1])
		if agent == nil {
			c.JSON(200, gin.H{"status": 0, "message": "平台不存在"})
			return
		}

		games, _ := models.GameGetList("aid", agent.Aid, "version", s[2])
		for _, game := range games {
			log.Println(game)
			go game.HotPatch()
		}
	} else {
		c.JSON(200, gin.H{"status": 0, "message": "版本标签错误"})
	}
}

// 获取某个平台，比指定version更大的所有版本
func GetVersions(agentFlag string, version int) []string {
	files, err := ioutil.ReadDir(UpdateZipDir)
	if err == nil {
		vers := make([]string, 0, len(files))
		for _, f := range files {
			if f.IsDir() {
				s := versionRegexp.FindStringSubmatch(f.Name())
				if len(s) == 3 && s[1] == agentFlag {
					ver, _ := strconv.Atoi(s[2])
					if ver > version {
						vers = append(vers, f.Name())
					}
				}
			}
		}
		return vers
	}
	return nil
}

// 获取所有的版本
func GetAllVersions() []os.FileInfo {
	files, err := ioutil.ReadDir(UpdateZipDir)
	if err == nil {
		vers := make([]os.FileInfo, 0, len(files))
		for _, f := range files {
			if f.IsDir() {
				s := versionRegexp.FindStringSubmatch(f.Name())
				if len(s) == 3 {
					vers = append(vers, f)
				}
			}
		}
		sort.Slice(vers, func(i, j int) bool {
			return vers[i].ModTime().Unix() > vers[j].ModTime().Unix()
		})
		return vers
	}
	return nil
}

// 获取某个版本tag下的所有补丁文件
func GetHotZips(tag string) []string {
	dir := fmt.Sprintf("%s/%s", UpdateZipDir, tag)
	files, err := ioutil.ReadDir(dir)
	if err == nil {
		zips := make([]string, 0)
		for _, f := range files {
			if !f.IsDir() {
				zipName := f.Name()
				if strings.HasPrefix(zipName, "hot_") && strings.HasSuffix(zipName, ".zip") {
					zips = append(zips, zipName)
				}
			}
		}
		return zips
	}
	return nil
}

func isVersionExist(version string) bool {
	verDir := fmt.Sprintf("%s/%s", UpdateZipDir, version)
	files, err := ioutil.ReadDir(verDir)
	if err == nil && len(files) > 0 {
		return true
	}
	return false
}

func VersionUpload(c *gin.Context) {
	userName := getUserName(c)
	dir := c.Query("dir")
	// log.Println(dir)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "文件上传",
		"dir":           dir,
	}
	c.HTML(200, "versionUpload", data)
}

func VersionUploadfile(c *gin.Context){
	file, _ := c.FormFile("file")
	dir := c.Query("dir")
	// log.Println(file.Filename)
	UpdateDir := path.Join(UpdateZipDir,dir,file.Filename)
	// Upload the file to specific dst.
	c.SaveUploadedFile(file, UpdateDir)
	c.JSON(200, gin.H{"code": 0, "msg": "", "message": "","data":file.Filename,})
}

func VersionDelfile(c *gin.Context){
	flag := c.PostForm("flag")
	DelDir := path.Join(UpdateZipDir,flag)
	os.RemoveAll(DelDir) 
	_, err := os.Stat(DelDir)
	if err == nil {
		c.JSON(200, gin.H{"status": -1, "message": "删除失败"})
	}
	if os.IsNotExist(err) {
		c.JSON(200, gin.H{"status": 0, "message": "删除成功"})
	}

}

func VersionMakedir(c *gin.Context){
	dir := c.PostForm("dir")
	// log.Println(file.Filename)
	NewDir := path.Join(UpdateZipDir,dir)
	err := os.MkdirAll(NewDir,os.ModePerm)
	if err == nil{
		c.JSON(200, gin.H{"status": 0, "message": "创建成功"})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "创建失败"})
	}
}