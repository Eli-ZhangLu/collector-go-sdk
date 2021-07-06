package reloader

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/paths"
)

// ReloadIf
type ReloadIf interface {
	Reload(*common.Config)
}

// Reloader is used to register and reload modules
type Reloader struct {
	hadnler ReloadIf
	name    string
	done    chan struct{}
	fd      interface{}
}

// PathConfig struct contains the basic path configuration of every beat
type PathConfig struct {
	Path paths.Path `config:"path"`
}
