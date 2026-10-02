package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/safetext"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

const progressBarWidth = 16

func RenderProgress(palette uikit.Palette, width int, icon, message, errText string, running bool, current, total int) string {
	if width <= 0 {
		return ""
	}

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
	line := strings.Join(segments, gap)
	if text != "" {
		if line != "" {
			line += gap
		}
		// A status owns one terminal row. Clip before styling/joining the
		// screen, otherwise a scanner's stderr can wrap the entire TUI away.
		available := uikit.Max(width-lipgloss.Width(line), 0)
		text = ansi.Truncate(safetext.Plain(text), available, "…")
		line += uikit.Text(palette, color, text)
	}
	// The counter and bar may themselves exceed a very narrow terminal.
	line = ansi.Truncate(line, width, "…")
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
