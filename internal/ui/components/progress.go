package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

const progressBarWidth = 16

func RenderProgress(palette uikit.Palette, width int, icon, message, errText string, running bool, current, total int) string {
	color := palette.Info
	text := message
	if errText != "" {
		color = palette.Error
		text = errText
	}

	gap := uikit.BackgroundSpaces(palette, 1)
	segments := make([]string, 0, 4)
	if icon != "" {
		segments = append(segments, uikit.Text(palette, color, icon))
	}
	if total > 0 {
		counter := "[" + strconv.Itoa(current) + "/" + strconv.Itoa(total) + "]"
		segments = append(segments, uikit.Text(palette, palette.Hint, counter))
		segments = append(segments, progressBar(palette, current, total))
	}
	if text != "" {
		segments = append(segments, uikit.Text(palette, color, text))
	}

	line := strings.Join(segments, gap)
	pad := uikit.Max(width-lipgloss.Width(line), 0)

	return line + uikit.BackgroundSpaces(palette, pad)
}

func progressBar(palette uikit.Palette, current, total int) string {
	if total <= 0 {
		return ""
	}
	filled := current * progressBarWidth / total
	filled = uikit.Max(uikit.Min(filled, progressBarWidth), 0)

	bar := uikit.Text(palette, palette.Primary, strings.Repeat("█", filled)) +
		uikit.Text(palette, palette.Disable, strings.Repeat("█", progressBarWidth-filled))

	return uikit.Text(palette, palette.Hint, "▕") + bar + uikit.Text(palette, palette.Hint, "▏")
}
