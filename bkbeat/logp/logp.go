package logp

import (
	"github.com/elastic/beats/libbeat/logp"
)

var L *logp.Logger

func SetLogger(logger *logp.Logger) {
	L = logger
}
