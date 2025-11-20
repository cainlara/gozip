package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cainlara/gozip/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func BuildUI(fileName string, content []core.ZippedFile) *tview.Application {
	app := tview.NewApplication()

	header := buildHeader()

	filterInput := tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)

	footer := tview.NewFlex().
		AddItem(filterInput, 0, 1, true)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false)

	table := buildContentTable(fileName, footer, filterInput, layout, app, content)

	layout.AddItem(table, 0, 1, true)

	return app.SetRoot(layout, true)
}

func buildHeader() *tview.TextView {
	header := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	header.SetText("[::b]goZip! [gray]• Up/Down select file • f filter • q exit[gray]")
	header.SetBackgroundColor(tcell.ColorReset)

	return header
}

func buildContentTable(fileName string, filterFooter *tview.Flex, filterInput *tview.InputField, layout *tview.Flex, app *tview.Application, content []core.ZippedFile) *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 0).
		SetSelectable(true, false)

	table.
		SetBorder(true).
		SetTitle(fileName).
		SetTitleAlign(tview.AlignCenter)

	allRows := make([][]string, 0, len(content))

	for _, zf := range content {
		row := []string{
			zf.GetName(),
			strconv.FormatBool(zf.IsDir()),
			strconv.FormatUint(zf.GetSize(), 10),
			zf.GetModifiedDate(),
			strconv.FormatUint(uint64(zf.GetCrc()), 10)}
		allRows = append(allRows, row)
	}

	headers := []string{"NAME", "IS FOLDER", "SIZE", "MODIFIED ON", "CRC"}

	populateTable := func(filterText string) {
		table.Clear()

		for c, h := range headers {
			cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", h)).
				SetSelectable(false).
				SetAlign(tview.AlignCenter)
			table.SetCell(0, c, cell)
		}

		rowIndex := 1
		filterLower := strings.ToLower(filterText)
		for _, row := range allRows {
			matches := filterText == ""
			if !matches {
				for _, val := range row {
					if strings.Contains(strings.ToLower(val), filterLower) {
						matches = true
						break
					}
				}
			}

			if matches {
				for c, val := range row {
					table.SetCell(rowIndex, c, tview.NewTableCell(val))
				}
				rowIndex++
			}
		}

		if rowIndex > 1 {
			table.Select(1, 0)
		}
	}

	populateTable("")

	table.Select(1, 0)

	filterMode := false

	filterInput.SetChangedFunc(func(text string) {
		populateTable(text)
	})

	filterInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape || key == tcell.KeyEnter {
			filterMode = false
			layout.RemoveItem(filterFooter)
			app.SetFocus(table)
		}
	})

	table.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		if ev.Key() == tcell.KeyRune {
			switch ev.Rune() {
			case 'q', 'Q':
				app.Stop()
				return nil
			case 'f', 'F':
				if !filterMode {
					filterMode = true
					filterInput.SetText("")
					layout.AddItem(filterFooter, 1, 0, true)
					app.SetFocus(filterInput)
					return nil
				}
			}
		}

		return ev
	})

	return table
}
