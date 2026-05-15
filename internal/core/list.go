package core

import (
	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
)

func List(cfg *config.Config) ([]FileInfo, error) {
	return content.Discover(cfg)
}
