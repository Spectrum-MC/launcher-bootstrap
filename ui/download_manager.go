/**
 * Spectrum-Bootstrap - A bootstrap for Minecraft launchers
 * Copyright (C) 2023-2024 - Oxodao
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 **/

package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/utils"
)

type DownloadTable struct {
	FilesToDownload []models.Downloadable
	ProgressBars    []*widget.ProgressBar

	filePrefix     string
	processedFiles int
	amtFiles       int
	table          *widget.Table
	scroll         *container.Scroll
	startedAt      time.Time
}

func (dt *DownloadTable) Refresh() {
	fyne.Do(func() {
		dt.table.Refresh()
	})
}

func (dt *DownloadTable) SetFileProgress(idx int, currSize int64, filesize int) {
	dt.ProgressBars[idx].SetValue(float64(currSize) / float64(filesize))
}

func (dt *DownloadTable) OnFileComplete(idx int) {
	dt.processedFiles = dt.processedFiles + 1
	dt.FilesToDownload = append(dt.FilesToDownload[:idx], dt.FilesToDownload[idx+1:]...)
	dt.Refresh()
}

func (dt *DownloadTable) GetProgressionStr() string {
	return fmt.Sprintf(
		"%v (%v/%v)",
		utils.FormatDuration(time.Since(dt.startedAt)),
		dt.processedFiles,
		dt.amtFiles,
	)
}

func (dt *DownloadTable) init() {
	progressBars := make([]*widget.ProgressBar, len(dt.FilesToDownload))

	table := widget.NewTable(
		func() (int, int) {
			return dt.amtFiles, 2
		},
		func() fyne.CanvasObject {
			return container.NewStack(widget.NewLabel(""), widget.NewProgressBar())
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			label := stack.Objects[0].(*widget.Label)
			pbar := stack.Objects[1].(*widget.ProgressBar)

			// Fyne is a joke
			// Apparently this is the recommended way to do it
			// https://stackoverflow.com/questions/73571674/how-to-make-a-table-with-different-object-types-in-fyne-io
			if id.Col == 0 {
				label.Show()
				pbar.Hide()
				label.Truncation = fyne.TextTruncateEllipsis
				label.Wrapping = fyne.TextWrapOff
				file := dt.FilesToDownload[id.Row]
				filepath := strings.TrimPrefix(file.Path, dt.filePrefix)
				fyne.Do(func() {
					label.SetText(filepath)
				})
			} else {
				label.Hide()
				pbar.Show()
				progressBars[id.Row] = pbar
				fyne.Do(func() {
					pbar.SetValue(dt.ProgressBars[id.Row].Value)
				})
			}
		},
	)

	table.SetColumnWidth(0, 500)
	table.SetColumnWidth(1, 200)

	vscroll := container.NewVScroll(table)
	vscroll.SetMinSize(fyne.NewSize(400, 400))

	dt.ProgressBars = progressBars
	dt.table = table
	dt.scroll = vscroll
}

func (dt *DownloadTable) GetScroll() *container.Scroll {
	return dt.scroll
}

func NewDownloadTable(filesToDownload []models.Downloadable, filePrefix string) *DownloadTable {
	filesToDownloadCopy := make([]models.Downloadable, len(filesToDownload))
	copy(filesToDownloadCopy, filesToDownload)

	dt := &DownloadTable{
		FilesToDownload: filesToDownloadCopy,
		filePrefix:      filePrefix,
		amtFiles:        len(filesToDownloadCopy),
		processedFiles:  0,
		startedAt:       time.Now(),
	}

	dt.init()

	return dt
}
