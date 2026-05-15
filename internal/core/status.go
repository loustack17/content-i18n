package core

import (
	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
)

func Status(cfg *config.Config) (*content.StatusReport, error) {
	files, err := content.Discover(cfg)
	if err != nil {
		return nil, err
	}

	return content.BuildStatusReport(cfg, files), nil
}
