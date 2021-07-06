package conf

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// https://www.jianshu.com/p/84499381a7da
type Mysql struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
}

type AppConf struct {
	Debug  bool   `yaml:"debug"`
	DDpush string `yaml:"ddpush"`
	Port   string `yaml:"port"`
}

type OamConf struct {
	TestSid  int `yaml:"test_sid"`
	AuditSid int `yaml:"audit_sid"`
}

type CenterConf struct {
	Version string  `yaml:"version"`
	SqlCnf  Mysql   `yaml:"mysql"`
	App     AppConf `yaml:"app"`
	Oam     OamConf `yaml:"oam"`
}

var YamlConf CenterConf

func init() {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &YamlConf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if YamlConf.Oam.TestSid >= YamlConf.Oam.AuditSid {
		log.Fatalf("配置错误：测试服id不能>=提审服id")
	}
	log.Println("conf", YamlConf)
}
