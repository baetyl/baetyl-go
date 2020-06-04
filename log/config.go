package log

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"github.com/baetyl/baetyl-go/errors"
)

// Config for logging
type Config struct {
	Level      string `yaml:"level" json:"level" default:"info" validate:"regexp=^(fatal|panic|error|warn|info|debug)$"`
	Encoding   string `yaml:"encoding" json:"encoding" default:"json" validate:"regexp=^(json|console)$"`
	Filename   string `yaml:"filename" json:"filename"`
	Compress   bool   `yaml:"compress" json:"compress"`
	MaxAge     int    `yaml:"maxAge" json:"maxAge" default:"15" validate:"min=1"`   // days
	MaxSize    int    `yaml:"maxSize" json:"maxSize" default:"50" validate:"min=1"` // MB
	MaxBackups int    `yaml:"maxBackups" json:"maxBackups" default:"15" validate:"min=1"`
}

func (c *Config) String() string {
	return fmt.Sprintf("level=%s&encoding=%s&filename=%s&compress=%t&maxAge=%d&maxSize=%d&maxBackups=%d",
		c.Level,
		c.Encoding,
		base64.URLEncoding.EncodeToString([]byte(c.Filename)),
		c.Compress,
		c.MaxAge,
		c.MaxSize,
		c.MaxBackups)
}

// FromURL creates config from url
func FromURL(u *url.URL) (*Config, error) {
	args := u.Query()
	c := new(Config)
	c.Level = args.Get("level")
	c.Encoding = args.Get("encoding")
	filename, err := base64.URLEncoding.DecodeString(args.Get("filename"))
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.Filename = string(filename)
	c.Compress, err = strconv.ParseBool(args.Get("compress"))
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.MaxAge, err = strconv.Atoi(args.Get("maxAge"))
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.MaxSize, err = strconv.Atoi(args.Get("maxSize"))
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.MaxBackups, err = strconv.Atoi(args.Get("maxBackups"))
	if err != nil {
		return nil, errors.Trace(err)
	}
	return c, nil
}
