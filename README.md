# DeployAndLog

### 定位

配置 Golang 基础服务（mysql、redis、http.client、log）比较繁琐，如果想 快速接入 基础服务可以使用本类库。 没有多余复杂的功能，方便你拓展其他功能。 你可以 import 引入直接使用，也可以拷贝代码到自己项目中使用，也可以用于构建自己的基础类库。

功能
多套配置环境设置，比如：dev、prod。
mysql、redis 多套数据源配置。
支持默认和自定义日志实例，自动滚动日志。
支持 mysql(基于gorm的二次开发支持ctx功能，不影响gorm原功能使用)、redis(redigo)、http.client 请求链路日志输出。


### log消息打印代码举例：
```go
package main

import (
	"github.com/e421083458/golang_common/lib"
	"log"
	"time"
)

func main() {
	if err := lib.InitModule("./conf/dev/",[]string{"base","mysql","redis",}); err != nil {
		log.Fatal(err)
	}
	defer lib.Destroy()

	//todo sth
	lib.Log.TagInfo(lib.NewTrace(), lib.DLTagUndefind, map[string]interface{}{"message": "todo sth"})
	time.Sleep(time.Second)
}
```