package bkmonitoring

import (
	"strconv"

	"github.com/elastic/beats/libbeat/monitoring"
)

var (
	bkbeat                    = monitoring.Default.NewRegistry("bkbeat")
	bkbeatTask                = monitoring.Default.NewRegistry("bkbeat_tasks")
	taskRegistry              = make(map[string]*monitoring.Registry)
	taskMetrics  *TaskMetrics = &TaskMetrics{
		Bools:   make(map[string]*monitoring.Bool),
		Strings: make(map[string]*monitoring.String),
		Ints:    make(map[string]*monitoring.Int),
		Floats:  make(map[string]*monitoring.Float),
	}
)

// NewBool 创建bkbeat指标-bool
func NewBool(name string, opts ...monitoring.Option) *monitoring.Bool {
	return monitoring.NewBool(bkbeat, name, opts...)
}

// NewString 创建bkbeat指标-string
func NewString(name string, opts ...monitoring.Option) *monitoring.String {
	return monitoring.NewString(bkbeat, name, opts...)
}

// NewInt 创建bkbeat指标-int
func NewInt(name string, opts ...monitoring.Option) *monitoring.Int {
	return monitoring.NewInt(bkbeat, name, opts...)
}

// NewFloat 创建bkbeat指标-float
func NewFloat(name string, opts ...monitoring.Option) *monitoring.Float {
	return monitoring.NewFloat(bkbeat, name, opts...)
}

// newRegistry creates and register a new registry by dataID
func newRegistryWithDataID(dataID int) *monitoring.Registry {
	registryKey := strconv.Itoa(dataID)
	if _, found := taskRegistry[registryKey]; !found {
		taskRegistry[registryKey] = bkbeatTask.NewRegistry(registryKey)
	}
	return taskRegistry[registryKey]
}

// TaskMetrics 用于动态获取任务指标
type TaskMetrics struct {
	Bools   map[string]*monitoring.Bool
	Strings map[string]*monitoring.String
	Ints    map[string]*monitoring.Int
	Floats  map[string]*monitoring.Float
}

// NewBoolWithDataID：获取任务指标-bool
func NewBoolWithDataID(dataID int, name string, opts ...monitoring.Option) *monitoring.Bool {
	reg := newRegistryWithDataID(dataID)
	metricKey := strconv.Itoa(dataID) + "_" + name
	if _, found := taskMetrics.Bools[metricKey]; !found {
		taskMetrics.Bools[metricKey] = monitoring.NewBool(reg, name, opts...)
	}
	return taskMetrics.Bools[metricKey]
}

// NewStringWithDataID：获取任务指标-string
func NewStringWithDataID(dataID int, name string, opts ...monitoring.Option) *monitoring.String {
	reg := newRegistryWithDataID(dataID)
	metricKey := strconv.Itoa(dataID) + "_" + name
	if _, found := taskMetrics.Strings[metricKey]; !found {
		taskMetrics.Strings[metricKey] = monitoring.NewString(reg, name, opts...)
	}
	return taskMetrics.Strings[metricKey]
}

// NewIntWithDataID：获取任务指标-int
func NewIntWithDataID(dataID int, name string, opts ...monitoring.Option) *monitoring.Int {
	reg := newRegistryWithDataID(dataID)
	metricKey := strconv.Itoa(dataID) + "_" + name
	if _, found := taskMetrics.Ints[metricKey]; !found {
		taskMetrics.Ints[metricKey] = monitoring.NewInt(reg, name, opts...)
	}
	return taskMetrics.Ints[metricKey]
}

// NewFloatWithDataID：获取任务指标-float
func NewFloatWithDataID(dataID int, name string, opts ...monitoring.Option) *monitoring.Float {
	reg := newRegistryWithDataID(dataID)
	metricKey := strconv.Itoa(dataID) + "_" + name
	if _, found := taskMetrics.Floats[metricKey]; !found {
		taskMetrics.Floats[metricKey] = monitoring.NewFloat(reg, name, opts...)
	}
	return taskMetrics.Floats[metricKey]
}
