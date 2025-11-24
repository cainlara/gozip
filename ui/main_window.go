// Package ui provides the terminal-based user interface for goZip.
// It uses the tview library to create an interactive interface that allows
// users to view and filter the contents of ZIP files.
package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cainlara/gozip/core"
	"github.com/cainlara/gozip/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// BuildUI constructs and configures the complete user interface for viewing ZIP files.
//
// The interface includes:
//   - A header with the title and keyboard shortcuts
//   - An interactive table displaying the ZIP file contents
//   - Filtering functionality activated with the 'f' key
//   - File extraction with the Enter key
//   - Navigation with arrow keys
//   - Exit with 'q' or Ctrl+C
//
// Parameters:
//   - fileName: name of the ZIP file to display in the title
//   - zipPath: full path to the ZIP file for extraction
//   - content: slice of ZippedFile with the ZIP file contents
//
// Returns:
//   - *tview.Application: configured tview application ready to run
//
// Usage:
//
//	app := BuildUI("archive.zip", "/path/to/archive.zip", contents)
//	app.Run()
func BuildUI(fileName string, zipPath string, content []core.ZippedFile) *tview.Application {
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

	table := buildContentTable(fileName, zipPath, footer, filterInput, layout, app, content)

	layout.AddItem(table, 0, 1, true)

	return app.SetRoot(layout, true)
}

func buildHeader() *tview.TextView {
	header := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	header.SetText("[::b]goZip! [gray]• Up/Down select • Enter extract • f filter • q exit[gray]")
	header.SetBackgroundColor(tcell.ColorReset)

	return header
}

func buildContentTable(fileName string, zipPath string, filterFooter *tview.Flex, filterInput *tview.InputField, layout *tview.Flex, app *tview.Application, content []core.ZippedFile) *tview.Table {
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

	var lastExtractedRow int = -1
	var extractionMessage string = ""

	filterInput.SetChangedFunc(func(text string) {
		populateTable(text)
	})

	filterInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape || key == tcell.KeyEnter {
			filterMode = false
			if key == tcell.KeyEscape {
				filterInput.SetText("")
				populateTable("")
			}
			layout.RemoveItem(filterFooter)
			app.SetFocus(table)
		}
	})

	table.SetSelectionChangedFunc(func(row, column int) {
		if lastExtractedRow != -1 && row != lastExtractedRow {
			table.SetTitle(fileName)
			lastExtractedRow = -1
			extractionMessage = ""
		}
	})

	table.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyEnter:
			row, _ := table.GetSelection()
			if row < 1 {
				return nil
			}

			fileNameCell := table.GetCell(row, 0)
			isDirCell := table.GetCell(row, 1)
			if fileNameCell == nil || isDirCell == nil {
				return nil
			}

			targetName := fileNameCell.Text
			isDir := isDirCell.Text == "true"

			if isDir {
				showConfirmationModal(app, layout, table, zipPath, targetName, &lastExtractedRow, &extractionMessage)
			} else {
				extractItem(table, zipPath, targetName, false, row, &lastExtractedRow, &extractionMessage)
			}
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

// showConfirmationModal displays a modal dialog asking for confirmation before extracting a folder.
func showConfirmationModal(app *tview.Application, layout *tview.Flex, table *tview.Table, zipPath, folderName string, lastExtractedRow *int, extractionMessage *string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Extract folder '%s' and all its contents?\n\nThis will extract all files within this folder recursively.", folderName)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				row, _ := table.GetSelection()
				extractItem(table, zipPath, folderName, true, row, lastExtractedRow, extractionMessage)
			}
			app.SetRoot(layout, true)
			app.SetFocus(table)
		})

	app.SetRoot(modal, true)
}

// extractItem performs the actual extraction and updates the table title with status.
func extractItem(table *tview.Table, zipPath, targetName string, isFolder bool, row int, lastExtractedRow *int, extractionMessage *string) {
	destDir, err := os.Getwd()
	if err != nil {
		table.SetTitle(fmt.Sprintf("[red]Error: %s[-]", err.Error()))
		return
	}

	count, err := util.ExtractFile(zipPath, targetName, destDir)
	if err != nil {
		table.SetTitle(fmt.Sprintf("[red]Error: %s[-]", err.Error()))
		*lastExtractedRow = -1
		*extractionMessage = ""
	} else {
		*lastExtractedRow = row

		if isFolder {
			*extractionMessage = fmt.Sprintf("[green]Extracted folder: %d files[-]", count)
		} else {
			*extractionMessage = fmt.Sprintf("[green]Extracted: %s[-]", targetName)
		}
		table.SetTitle(*extractionMessage)
	}
}
