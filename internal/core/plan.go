package core

import (
	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
)

func Plan(cfg *config.Config, sourceFile string, targetLang string) ([]content.FileInfo, error) {
	if sourceFile != "" && targetLang != "" {
		_, err := GenerateWorkPacket(cfg, sourceFile, targetLang)
		if err != nil {
			return nil, err
		}
		targetPath, err := content.TargetPath(cfg, sourceFile, targetLang)
		if err != nil {
			return nil, err
		}
		return []content.FileInfo{{
			SourcePath: sourceFile,
			TargetPath: targetPath,
			Language:   targetLang,
			Status:     "planned",
		}}, nil
	}

	files, err := content.Discover(cfg)
	if err != nil {
		return nil, err
	}

	return content.MissingTranslations(files), nil
}
