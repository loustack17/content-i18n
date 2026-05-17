package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type toolSpec struct {
	def     mcp.Tool
	handler func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

func allToolSpecs(s *Server) []toolSpec {
	return []toolSpec{
		{
			def: mcp.NewTool("content_i18n_status",
				mcp.WithDescription("Return project status: source/target counts, missing translations"),
			),
			handler: s.handleStatus,
		},
		{
			def: mcp.NewTool("content_i18n_prepare_translation",
				mcp.WithDescription("Prepare everything needed to translate a source file. Returns source content, prompt, glossary, style rules, structure fingerprint, and target path. Replaces read_source + create_work_packet."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Absolute path to source file"),
				),
				mcp.WithString("language",
					mcp.Required(),
					mcp.Description("Target language code"),
				),
			),
			handler: s.handlePrepareTranslation,
		},
		{
			def: mcp.NewTool("content_i18n_review_translation",
				mcp.WithDescription("Review a translation against its source. Returns passed, ready_to_sync, word ratio, and severity-tagged issues. ready_to_sync=true means the translation passes all structural and content checks and can be synced without further repair."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Absolute path to source file"),
				),
				mcp.WithString("target",
					mcp.Required(),
					mcp.Description("Absolute path to translated target file"),
				),
			),
			handler: s.handleReviewTranslation,
		},
		{
			def: mcp.NewTool("content_i18n_sync_status",
				mcp.WithDescription("Sync completion status for a directly-edited translation target. Validates target exists, source exists, path pair matches config, and content passes validation. Updates official status store on success."),
				mcp.WithString("target",
					mcp.Required(),
					mcp.Description("Target translation file path"),
				),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source file path"),
				),
			),
			handler: s.handleSyncStatus,
		},
		{
			def: mcp.NewTool("content_i18n_translation_queue",
				mcp.WithDescription("Get the full translation queue status: total, completed, stale, missing, and next candidate. Supports group filtering. Replaces next_translation."),
				mcp.WithString("group",
					mcp.Description("Optional group filter (e.g. DevOps)"),
				),
			),
			handler: s.handleTranslationQueue,
		},
		{
			def: mcp.NewTool("content_i18n_translate_batch",
				mcp.WithDescription("Batch translate all queued files. Drains queue automatically: prepare → translate → review → repair → validate → CJK check → sync-status. Returns completed/failed/remaining report. Never marks complete unless validation passes."),
				mcp.WithString("group",
					mcp.Description("Optional group filter (e.g. DevOps)"),
				),
				mcp.WithString("provider",
					mcp.Description("Translation provider: ai-harness, deepl, google, auto"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Max files to process (0 = all)"),
				),
				mcp.WithBoolean("stop_on_fail",
					mcp.Description("Stop on first failure"),
				),
				mcp.WithBoolean("continue_on_error",
					mcp.Description("Continue processing after failures"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Show plan without executing"),
				),
			),
			handler: s.handleTranslateBatch,
		},
		{
			def: mcp.NewTool("content_i18n_validate_site",
				mcp.WithDescription("Validate Hugo site URL policy and canonical routes"),
			),
			handler: s.handleValidateSite,
		},
	}
}
