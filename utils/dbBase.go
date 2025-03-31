package utils

import (
	"time"
)

type configDb struct {
	Dsn            string
	TimeoutConnect time.Duration
}

var ConfigDb = configDb{
	TimeoutConnect: 5,
}
