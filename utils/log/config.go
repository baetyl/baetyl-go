package log

import (
	"encoding/base64"
	"fmt"
)

// Config for logging
type Config struct {
	Path   string `yaml:"path" json:"path"`
	Level  string `yaml:"level" json:"level" default:"info" validate:"regexp=^(fatal|panic|error|warn|info|debug)$"`
	Format string `yaml:"format" json:"format" default:"text" validate:"regexp=^(text|json)$"`
	Age    struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
	} `yaml:"age" json:"age"` // days
	Size struct {
		Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
	} `yaml:"size" json:"size"` // in MB
	Backup struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
	} `yaml:"backup" json:"backup"`
}

func (c *Config) String() string {
	return fmt.Sprintf("path=%s&level=%s&format=%s&age_max=%d&size_max=%d&backup_max=%d",
		base64.URLEncoding.EncodeToString([]byte(c.Path)),
		c.Level,
		c.Format,
		c.Age.Max,
		c.Size.Max,
		c.Backup.Max)
}
