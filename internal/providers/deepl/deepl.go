package deepl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Provider struct {
	apiKey string
	apiURL string
}

type TranslateRequest struct {
	Text               []string `json:"text"`
	SourceLang         string   `json:"source_lang,omitempty"`
	TargetLang         string   `json:"target_lang"`
	GlossaryID         string   `json:"glossary_id,omitempty"`
	Formality          string   `json:"formality,omitempty"`
	PreserveFormatting int      `json:"preserve_formatting,omitempty"`
	TagHandling        string   `json:"tag_handling,omitempty"`
}

type TranslateResponse struct {
	Translations []Translation `json:"translations"`
}

type Translation struct {
	Text               string `json:"text"`
	DetectedSourceLang string `json:"detected_source_lang,omitempty"`
}

type GlossaryEntry struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

func New() (*Provider, error) {
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPL_API_KEY not set")
	}

	apiURL := os.Getenv("DEEPL_API_URL")
	if apiURL == "" {
		apiURL = "https://api-free.deepl.com/v2/translate"
	}

	return &Provider{
		apiKey: apiKey,
		apiURL: apiURL,
	}, nil
}

func (p *Provider) Available() bool {
	return p.apiKey != ""
}

func (p *Provider) Translate(text string, sourceLang string, targetLang string) (string, error) {
	lang := deeplLangCode(targetLang)
	srcLang := deeplLangCode(sourceLang)

	reqBody := TranslateRequest{
		Text:       []string{text},
		SourceLang: srcLang,
		TargetLang: lang,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.apiURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deepl API error %d: %s", resp.StatusCode, string(body))
	}

	var result TranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Translations) == 0 {
		return "", fmt.Errorf("no translations returned")
	}

	return result.Translations[0].Text, nil
}

func (p *Provider) TranslateBatch(texts []string, sourceLang string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	lang := deeplLangCode(targetLang)
	srcLang := deeplLangCode(sourceLang)

	reqBody := TranslateRequest{
		Text:       texts,
		SourceLang: srcLang,
		TargetLang: lang,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deepl API error %d: %s", resp.StatusCode, string(body))
	}

	var result TranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	translated := make([]string, len(result.Translations))
	for i, t := range result.Translations {
		translated[i] = t.Text
	}

	return translated, nil
}

func CompileGlossary(entries []GlossaryEntry) string {
	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s,%s", e.Source, e.Target))
	}
	return strings.Join(lines, "\n")
}

func deeplLangCode(lang string) string {
	lang = strings.ToUpper(lang)
	switch lang {
	case "EN":
		return "EN-US"
	case "ZH-TW":
		return "ZH"
	case "ZH-CN":
		return "ZH"
	case "JA":
		return "JA"
	case "KO":
		return "KO"
	case "DE":
		return "DE"
	case "FR":
		return "FR"
	case "ES":
		return "ES"
	case "PT":
		return "PT-PT"
	case "PT-BR":
		return "PT-BR"
	case "RU":
		return "RU"
	case "IT":
		return "IT"
	case "NL":
		return "NL"
	case "PL":
		return "PL"
	case "BG":
		return "BG"
	case "CS":
		return "CS"
	case "DA":
		return "DA"
	case "EL":
		return "EL"
	case "ET":
		return "ET"
	case "FI":
		return "FI"
	case "HU":
		return "HU"
	case "LT":
		return "LT"
	case "LV":
		return "LV"
	case "RO":
		return "RO"
	case "SK":
		return "SK"
	case "SL":
		return "SL"
	case "SV":
		return "SV"
	default:
		return lang
	}
}
