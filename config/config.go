// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period   time.Duration `config:"period"`
	URLs     []string      `config:urls`
	Username string        `config:username`
	Password string        `config:password`
}

var DefaultConfig = Config{
	Period:   1 * time.Second,
	URLs:     []string{"http://www.apache.org/server-status"},
	Username: "",
	Password: "",
}
