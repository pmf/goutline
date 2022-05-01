module goutline

go 1.13

replace github.com/charmbracelet/bubbletea => ../bubbletea
replace github.com/charmbracelet/bubbles => ../bubbles
replace github.com/muesli/cancelreader v0.2.0 => ../cancelreader

require (
	github.com/charmbracelet/bubbles v0.10.3
	github.com/charmbracelet/bubbletea v0.19.3
	github.com/charmbracelet/glamour v0.5.0
	github.com/charmbracelet/lipgloss v0.5.0
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-isatty v0.0.14
	github.com/muesli/reflow v0.3.0
	github.com/muesli/termenv v0.11.1-0.20220212125758-44cd13922739
)

