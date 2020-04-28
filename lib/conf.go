package lib

import (
	"bytes"
	"database/sql"
	"github.com/e421083458/gorm"
	_ "github.com/e421083458/gorm/dialects/mysql"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"github.com/xiaka53/DeployAndLog/log"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type BaseConf struct {
	DebugMode    string    `mapstructure:"debug_mode"`
	TimeLocation string    `mapstructure:"time_location"`
	Log          LogConfig `mapstructure:"log"`
	Base         struct {
		DebugMode    string `mapstructure:"debug_mode"`
		TimeLocation string `mapstructure:"time_location"`
		WebUrl       string `mapstructure:"web_url"`
		WebName      string `mapstructure:"web_name"`
	} `mapstructure:"base"`
}

type LogConfFileWriter struct {
	On              bool   `mapstructure:"on"`
	LogPath         string `mapstructure:"log_path"`
	RotateLogPath   string `mapstructure:"rotate_log_path"`
	WfLogPath       string `mapstructure:"wf_log_path"`
	RotateWfLogPath string `mapstructure:"rotate_wf_log_path"`
}

type LogConfConsoleWriter struct {
	On    bool `mapstructure:"on"`
	Color bool `mapstructure:"color"`
}

type LogConfig struct {
	Level string               `mapstructure:"log_level"`
	FW    LogConfFileWriter    `mapstructure:"file_writer"`
	CW    LogConfConsoleWriter `mapstructure:"console_writer"`
}

type MysqlMapConf struct {
	List map[string]*MySQLConf `mapstructure:"list"`
}

type MySQLConf struct {
	DriverName      string `mapstructure:"driver_name"`
	DataSourceName  string `mapstructure:"data_source_name"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"`
}

type RedisMapConf struct {
	List map[string]*RedisConf `mapstructure:"list"`
}

type RedisConf struct {
	ProxyList    []string `mapstructure:"proxy_list"`
	MaxIdle      int      `mapstructure:"max_idle"`
	MaxActive    int      `mapstructure:"max_active"`
	ConnTimeout  int      `mapstructure:"conn_timeout"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
}

var (
	ConfBase         *BaseConf
	DBMapPool        map[string]*sql.DB
	GORMMapPool      map[string]*gorm.DB
	DBDefaultPool    *sql.DB
	GORMDefaultPool  *gorm.DB
	ConfRedis        *RedisConf
	ConfRedisMap     *RedisMapConf
	RedisMapPool     map[string]*redis.Pool
	RedisDefaultPool *redis.Pool
	ViperConfMap     map[string]*viper.Viper
)

//初始化base配置
func InitBaseConf(path string) (err error) {
	var (
		logConf log.LogConfig
	)
	ConfBase = &BaseConf{}
	if err = ParseConfig(path, ConfBase); err != nil {
		return
	}

	if ConfBase.DebugMode == "" {
		if ConfBase.Base.DebugMode != "" {
			ConfBase.DebugMode = ConfBase.Base.DebugMode
		} else {
			ConfBase.DebugMode = "debug"
		}
	}
	if ConfBase.TimeLocation == "" {
		if ConfBase.Base.TimeLocation != "" {
			ConfBase.TimeLocation = ConfBase.Base.TimeLocation
		} else {
			ConfBase.TimeLocation = "Asia/Shanghai"
		}
		log.SetLoadLocation(ConfBase.TimeLocation)
	}
	if ConfBase.Log.Level == "" {
		ConfBase.Log.Level = "trace"
	}

	//配置日志
	logConf = log.LogConfig{
		Level: ConfBase.Log.Level,
		FW: log.ConfFileWriter{
			On:              ConfBase.Log.FW.On,
			LogPath:         ConfBase.Log.FW.LogPath,
			RotateLogPath:   ConfBase.Log.FW.RotateLogPath,
			WfLogPath:       ConfBase.Log.FW.WfLogPath,
			RotateWfLogPath: ConfBase.Log.FW.RotateWfLogPath,
		},
		CW: log.ConfConsoleWriter{
			On:    ConfBase.Log.CW.On,
			Color: ConfBase.Log.CW.Color,
		},
	}

	if err = log.SetupDefaultLogWithConf(logConf); err != nil {
		panic(err)
	}
	log.SetLayout("2006-01-02T15:04:05.000")
	return
}

//TODO 旧的获取redis
//func InitRedisConf(path string) error {
//	ConfRedis := &RedisMapConf{}
//	err := ParseConfig(path, ConfRedis)
//	if err != nil {
//		return err
//	}
//	ConfRedisMap = ConfRedis
//	return nil
//}

//初始化配置文件
func InitViperConf() (err error) {
	var (
		f        *os.File
		fileList []os.FileInfo
	)
	if f, err = os.Open(ConfEnvPath + "/"); err != nil {
		return
	}
	if fileList, err = f.Readdir(1024); err != nil {
		return
	}
	for _, v := range fileList {
		var (
			bts     []byte
			val     *viper.Viper
			pathArr []string
		)
		if !v.IsDir() {
			if bts, err = ioutil.ReadFile(ConfEnvPath + "/" + v.Name()); err != nil {
				return
			}
			val = viper.New()
			val.SetConfigType("toml")
			if err = val.ReadConfig(bytes.NewBuffer(bts)); err != nil {
				return
			}
			pathArr = strings.Split(v.Name(), ".")
			if ViperConfMap == nil {
				ViperConfMap = make(map[string]*viper.Viper)
			}
			ViperConfMap[pathArr[0]] = val
		}
	}
	return
}

//获取get配置信息
func GetStringConf(key string) string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return ""
	}
	v, ok := ViperConfMap[keys[0]]
	if !ok {
		return ""
	}
	confString := v.GetString(strings.Join(keys[1:len(keys)], "."))
	return confString
}

//获取get配置信息
func GetStringMapConf(key string) map[string]interface{} {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetStringMap(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetConf(key string) interface{} {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.Get(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetBoolConf(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetBool(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetFloat64Conf(key string) float64 {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetFloat64(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetIntConf(key string) int {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetInt(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetStringMapStringConf(key string) map[string]string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetStringMapString(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetStringSliceConf(key string) []string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetStringSlice(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取get配置信息
func GetTimeConf(key string) time.Time {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return time.Now()
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetTime(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//获取时间阶段长度
func GetDurationConf(key string) time.Duration {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetDuration(strings.Join(keys[1:len(keys)], "."))
	return conf
}

//是否设置了key
func IsSetConf(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	conf := v.IsSet(strings.Join(keys[1:len(keys)], "."))
	return conf
}
