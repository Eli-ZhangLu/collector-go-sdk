package bkpipe

import (
	"time"
)

//监控数据上报配置：如果DataID为空，则直接打日志
type config struct {
	BkBizID    int32         `config:"bk_biz_id"`
	DataID     int32         `config:"dataid"`
	TaskDataID int32         `config:"task_dataid"`
	Period     time.Duration `config:"period"`
}

var defaultConfig = config{
	BkBizID: 2,
	DataID:  0,
	Period:  60 * time.Second,
}
