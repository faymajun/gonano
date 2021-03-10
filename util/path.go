package util

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func GetServiceIdByPath() (cluster string, index int, err error) {
	var dir string
	var strIndex string
	if dir, err = os.Getwd(); err != nil {
		err = fmt.Errorf("get cwd failed: %s", err.Error())
		return
	} else {
		id := strings.Split(path.Base(dir), "-")
		if len(id) != 2 {
			err = fmt.Errorf("last dir(%s) is not a valid serviceId", path.Base(dir))
			return
		}
		cluster, strIndex = id[0], id[1]
		if index, err = strconv.Atoi(strIndex); err != nil {
			err = fmt.Errorf("last dir(%s) is not a valid serviceId", path.Base(dir))
			return
		}
		return
	}
}
