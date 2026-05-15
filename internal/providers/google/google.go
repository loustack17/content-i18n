package google

import (
	"context"
	"fmt"
	"os"
	"strings"

	translate "cloud.google.com/go/translate/apiv3"
	translatepb "cloud.google.com/go/translate/apiv3/translatepb"
)

type Provider struct {
	client   *translate.TranslationClient
	project  string
	location string
}

type GlossaryEntry struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

func New() (*Provider, error) {
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if creds == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS not set")
	}

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT not set")
	}

	location := os.Getenv("GOOGLE_TRANSLATE_LOCATION")
	if location == "" {
		location = "global"
	}

	ctx := context.Background()
	client, err := translate.NewTranslationClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create translation client: %w", err)
	}

	return &Provider{
		client:   client,
		project:  project,
		location: location,
	}, nil
}

func (p *Provider) Available() bool {
	return p.client != nil
}

func (p *Provider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

func (p *Provider) Translate(text string, sourceLang string, targetLang string) (string, error) {
	results, err := p.TranslateBatch([]string{text}, sourceLang, targetLang)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no translation returned")
	}
	return results[0], nil
}

func (p *Provider) TranslateBatch(texts []string, sourceLang string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	ctx := context.Background()
	parent := fmt.Sprintf("projects/%s/locations/%s", p.project, p.location)

	req := &translatepb.TranslateTextRequest{
		Parent:             parent,
		TargetLanguageCode: googleLangCode(targetLang),
		MimeType:           "text/plain",
		Contents:           texts,
	}

	if sourceLang != "" {
		req.SourceLanguageCode = googleLangCode(sourceLang)
	}

	resp, err := p.client.TranslateText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("translate: %w", err)
	}

	results := make([]string, len(resp.GetTranslations()))
	for i, t := range resp.GetTranslations() {
		results[i] = t.GetTranslatedText()
	}

	return results, nil
}

func (p *Provider) CreateGlossary(ctx context.Context, glossaryID string, sourceLang string, targetLang string, entries []GlossaryEntry) error {
	return fmt.Errorf("glossary creation requires GCS upload; use TranslateWithGlossary with existing glossary")
}

func (p *Provider) TranslateWithGlossary(ctx context.Context, text string, sourceLang string, targetLang string, glossaryID string) (string, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", p.project, p.location)
	glossaryConfig := &translatepb.TranslateTextGlossaryConfig{
		Glossary: fmt.Sprintf("%s/glossaries/%s", parent, glossaryID),
	}

	req := &translatepb.TranslateTextRequest{
		Parent:             parent,
		TargetLanguageCode: googleLangCode(targetLang),
		SourceLanguageCode: googleLangCode(sourceLang),
		MimeType:           "text/plain",
		Contents:           []string{text},
		GlossaryConfig:     glossaryConfig,
	}

	resp, err := p.client.TranslateText(ctx, req)
	if err != nil {
		return "", fmt.Errorf("translate with glossary: %w", err)
	}

	translations := resp.GetTranslations()
	if len(translations) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	return translations[0].GetTranslatedText(), nil
}

func googleLangCode(lang string) string {
	lang = strings.ToLower(lang)
	switch lang {
	case "zh-tw":
		return "zh-TW"
	case "zh-cn":
		return "zh-CN"
	case "en":
		return "en"
	case "ja":
		return "ja"
	case "ko":
		return "ko"
	case "de":
		return "de"
	case "fr":
		return "fr"
	case "es":
		return "es"
	case "pt-br":
		return "pt-BR"
	case "pt":
		return "pt"
	case "ru":
		return "ru"
	case "it":
		return "it"
	default:
		return lang
	}
}
