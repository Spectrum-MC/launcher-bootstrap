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

package httpclient

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spectrum-mc/bootstrap/localize"
	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/ui"
)

type DownloadManager struct {
	ui         *ui.MainUi
	onComplete func()
}

func (dm *DownloadManager) SetOnComplete(f func()) {
	dm.onComplete = f
}

func (dm *DownloadManager) downloadFile(
	table *ui.DownloadTable,
	timeLabel *widget.Label,
	i int,
) {
	f := table.FilesToDownload[i]

	err := os.MkdirAll(filepath.Dir(f.Path), os.ModePerm)
	if dm.ui.ShowFailedDownloadError(err) {
		return
	}

	out, err := os.Create(f.Path)
	if dm.ui.ShowFailedDownloadError(err) {
		return
	}

	done := make(chan int64)
	go func(idx int, f models.Downloadable) {
		var stop bool = false
		for {
			select {
			case <-done:
				stop = true
			default:
				fi, err := os.Stat(f.Path)
				if err != nil {
					log.Fatal(err)
				}

				currSize := fi.Size()
				if currSize == 0 {
					currSize = 1
				}

				fyne.Do(func() {
					table.SetFileProgress(idx, currSize, f.Size)
					timeLabel.SetText(table.GetProgressionStr())
				})
			}

			if stop {
				break
			}

			time.Sleep(50 * time.Millisecond)
		}
	}(i, f)

	dm.ui.CenterOnScreen()

	// @TODO: 3 Retries per file
	req, err := NewRequest("GET", f.Url, nil)
	if dm.ui.ShowFailedDownloadError(err) {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if dm.ui.ShowFailedDownloadError(err) {
		return
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if dm.ui.ShowFailedDownloadError(err) {
		return
	}

	out.Close()

	if f.Executable {
		err := os.Chmod(f.Path, os.ModePerm)
		if dm.ui.ShowFailedDownloadError(err) {
			return
		}
	}

	done <- n
}

/**
 * @TODO: At some point
 * we need to get the ui out tf of this package
 * DM should only DM and tell the UI everything it needs to know
 **/
func (dm *DownloadManager) Download(filesToDownload []models.Downloadable) {
	timeLabel := widget.NewLabel("00:00:00")
	mainProgressBar := widget.NewProgressBar()

	size := fyne.NewSize(300, 200)
	if len(filesToDownload) > 0 {
		size = fyne.NewSize(750, 500)
	}

	dm.ui.Resize(size)

	dlTable := ui.NewDownloadTable(
		filesToDownload,
		filepath.Join(BOOTSTRAP_SETTINGS.LauncherPath, "runtime"),
	) // @TODO: add /{component}/{os} in the prefix

	dm.ui.SetContent(
		widget.NewLabel(localize.Localize("downloading", nil)),
		container.NewHBox(
			widget.NewLabel(localize.Localize("elapsed_time", nil)),
			timeLabel,
		),
		mainProgressBar,
		dlTable.GetScroll(),
	)

	processedFiles := 0
	for i := 0; i < len(dlTable.FilesToDownload); {
		dm.downloadFile(dlTable, timeLabel, i)

		processedFiles += 1
		fyne.Do(func() {
			mainProgressBar.SetValue(float64(processedFiles) / float64(len(filesToDownload)))
		})

		dlTable.OnFileComplete(i)
	}

	if dm.onComplete != nil {
		dm.onComplete()
	}
}

func NewDownloadManager(ui *ui.MainUi) *DownloadManager {
	return &DownloadManager{
		ui: ui,
	}
}
