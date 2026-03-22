package main

import (
	"strings"
)

type legacyRenderBlock struct {
	Text string
	Rows int
}

func (cli *TelegramCLI) renderLegacyChatViewWithInlineImages(label string, target string, entries []legacyTranscriptEntry, width int, height int, cfg inlineImageConfig) (string, bool) {
	if width <= 0 {
		width = 80
	}
	if height <= 1 {
		height = 2
	}

	headerRows := []string{
		dim(truncateVisibleWidth("Active chat: "+strings.TrimSpace(label)+" ("+strings.TrimSpace(target)+")", width)),
		dim(strings.Repeat("─", maxInt(width, 1))),
	}

	maxRows := height - 1
	if maxRows < len(headerRows) {
		maxRows = len(headerRows)
	}
	availableBodyRows := maxRows - len(headerRows)
	if availableBodyRows < 0 {
		availableBodyRows = 0
	}

	if len(entries) == 0 {
		rows := append(append([]string{}, headerRows...), dim("No messages yet."))
		for len(rows) < maxRows {
			rows = append(rows, "")
		}
		return strings.Join(rows[:maxRows], "\n"), true
	}

	blocks := make([]legacyRenderBlock, 0, len(entries))
	usedRows := 0
	for i := len(entries) - 1; i >= 0; i-- {
		block, ok := cli.renderLegacyEntryBlock(target, entries[i], width, cfg)
		if !ok {
			return "", false
		}
		if block.Rows <= 0 {
			continue
		}

		extraRows := block.Rows
		if len(blocks) > 0 {
			extraRows++
		}

		if usedRows+extraRows > availableBodyRows {
			if len(blocks) == 0 {
				blocks = append(blocks, block)
			}
			break
		}

		blocks = append(blocks, block)
		usedRows += extraRows
	}

	for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}

	bodyRows := make([]string, 0, usedRows)
	for idx, block := range blocks {
		bodyRows = append(bodyRows, strings.Split(block.Text, "\n")...)
		if idx < len(blocks)-1 {
			bodyRows = append(bodyRows, "")
		}
	}

	rows := append(append([]string{}, headerRows...), bodyRows...)
	for len(rows) < maxRows {
		rows = append(rows, "")
	}
	if len(rows) > maxRows {
		rows = rows[len(rows)-maxRows:]
	}
	return strings.Join(rows, "\n"), true
}

func (cli *TelegramCLI) renderLegacyEntryBlock(target string, entry legacyTranscriptEntry, width int, cfg inlineImageConfig) (legacyRenderBlock, bool) {
	inlineBody := entry.Body
	imageBlock := ""
	imageRows := 0

	if entry.Image != nil {
		if block, rows, ok := cli.renderInlineImageBlock(target, entry, width, cfg); ok {
			imageBlock = block
			imageRows = rows
			inlineBody = imagePreviewText(entry.Image, imageCaptionText(entry.Body))
		}
	}

	bubble := renderTranscriptBubbleForWidth(entry.Outgoing, entry.Header, inlineBody, entry.Meta, width)
	rows := len(strings.Split(bubble, "\n"))
	text := bubble

	if imageBlock != "" && imageRows > 0 {
		text += "\n" + imageBlock
		rows += imageRows
	}

	return legacyRenderBlock{
		Text: text,
		Rows: rows,
	}, true
}
