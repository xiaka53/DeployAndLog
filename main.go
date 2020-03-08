package main

import (
	"github.com/xiaka53/DeployAndLog/lib"
	"log"
	"time"
)

func main() {
	if err := lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"}); err != nil {
		log.Fatal(err)
	}
	defer lib.Destroy()

	lib.Log.TagInfo(lib.NewTrace(), lib.DLTagUndefind, map[string]interface{}{"message": "todo sth"})
	time.Sleep(time.Second)
}
