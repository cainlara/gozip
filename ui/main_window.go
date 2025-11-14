package ui

import (
	"fmt"
	"strconv"

	"github.com/cainlara/gozip/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func BuildUI(fileName string, content []core.ZippedFile) *tview.Application {
	app := tview.NewApplication()

	header := buildHeader()

	footer := tview.NewTextView().
		SetText("selected: {{name}}").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	table := buildContentTable(fileName, footer, app, content)

	// Layout: Header | Table (expand) | Footer
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(table, 0, 1, true) //.
		// AddItem(footer, 1, 0, false)

	return app.SetRoot(layout, true)
}

func updateFooter(table *tview.Table, footer *tview.TextView, row int) {
	if row == 0 {
		table.Select(1, 0)
		row = 1
	}
	userCell := table.GetCell(row, 2)
	if userCell != nil {
		footer.SetText(fmt.Sprintf("selected: %s", userCell.Text))
	}
}

func buildHeader() *tview.TextView {
	header := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	header.SetText("[::b]goZip! [gray]• Up/Down select file • q exit[gray]")
	header.SetBackgroundColor(tcell.ColorReset)

	return header
}

func buildContentTable(fileName string, footer *tview.TextView, app *tview.Application, content []core.ZippedFile) *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 0).
		SetSelectable(true, false)

	table.
		SetBorder(true).
		SetTitle(fileName).
		SetTitleAlign(tview.AlignCenter)

	rows := make([][]string, 0, 10)

	for _, zf := range content {
		row := []string{
			zf.GetName(),
			strconv.FormatBool(zf.IsDir()),
			strconv.FormatUint(zf.GetSize(), 10),
			zf.GetModifiedDate(),
			strconv.FormatUint(uint64(zf.GetCrc()), 10)}
		rows = append(rows, row)
	}

	headers := []string{"NAME", "IS FOLDER", "SIZE", "MODIFIED ON", "CRC"}

	for c, h := range headers {
		cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", h)).
			SetSelectable(false).
			SetAlign(tview.AlignCenter)
		table.SetCell(0, c, cell)
	}

	for r, row := range rows {
		for c, val := range row {
			table.SetCell(r+1, c, tview.NewTableCell(val))
		}
	}

	// Callback cuando cambia la selección
	table.SetSelectionChangedFunc(func(row, column int) {
		updateFooter(table, footer, row)
	})

	// Selección inicial en la primera fila de datos
	table.Select(1, 0)
	updateFooter(table, footer, 1)

	table.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		if ev.Key() == tcell.KeyRune && (ev.Rune() == 'q' || ev.Rune() == 'Q') {
			app.Stop()
			return nil
		}
		return ev
	})

	return table
}
