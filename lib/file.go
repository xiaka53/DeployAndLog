package lib

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var (
	ConfEnvPath string //配置文件夹
	ConfEnv     string //配置环境 如：dev，test
)

//解析配置文件目录
//配置文件必须放到同一个文件夹里
//如：config=conf/dev/base.json  ConfEnvPath=conf/dev  ConfEnv=dev
//如：config=conf/base.json  ConfEnvPath=conf  ConfEnv=config
func ParseConfPath(config string) (err error) {
	var (
		path   []string
		prefix string
	)
	path = strings.Split(config, "/")
	prefix = strings.Join(path[:len(path)-1], "/")
	ConfEnvPath = prefix
	ConfEnv = path[len(path)-2]
	return
}

//获取配置环境名
func GetConfEnv() string {
	return ConfEnv
}

func GetConfPath(fileName string) string {
	return ConfEnvPath + "/" + fileName + ".toml"
}

func GetConfFilePath(fileName string) string {
	return ConfEnvPath + "/" + fileName
}

//本地解析文件
func ParseLocalConfig(fileName string, st interface{}) error {
	path := GetConfFilePath(fileName)
	err := ParseConfig(path, st)
	if err != nil {
		return err
	}
	return nil
}

//读取配置文件并获取配置信息
func ParseConfig(path string, conf interface{}) (err error) {
	var (
		file *os.File
		data []byte
		v    *viper.Viper
	)
	if file, err = os.Open(path); err != nil {
		return fmt.Errorf("Open config %v fail,%v", path, err)
	}
	if data, err = ioutil.ReadAll(file); err != nil {
		return fmt.Errorf("Read config fail,%v", err)
	}

	v = viper.New()
	v.SetConfigType("toml")
	if err = v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return
	}
	if err = v.Unmarshal(conf); err != nil {
		return fmt.Errorf("Parse config fail,config:%v,err:%v", string(data), err)
	}
	return
}
