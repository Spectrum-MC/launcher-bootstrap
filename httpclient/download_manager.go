package httpclient

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spectrum-mc/bootstrap/localize"
	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/ui"
	"github.com/spectrum-mc/bootstrap/utils"
)

type DownloadManager struct {
	ui         *ui.MainUi
	settings   *models.BootstrapSettings
	onComplete func()
}

func (dm *DownloadManager) SetOnComplete(f func()) {
	dm.onComplete = f
}

func (dm *DownloadManager) Download(filesToDownload []models.Downloadable) {
	timeLabel := widget.NewLabel("00:00:00")
	mainProgressBar := widget.NewProgressBar()

	// @TODO Make this base on goroutine to download multiple file at once
	// @TODO which will be hard to display properly like SKCraft

	filesToDownloadCopy := make([]models.Downloadable, len(filesToDownload))
	copy(filesToDownloadCopy, filesToDownload)

	size := fyne.NewSize(300, 200)
	if len(filesToDownload) > 0 {
		size = fyne.NewSize(750, 500)
	}

	fyne.Do(func() {
		dm.ui.MainWindow.Resize(size)
	})
	dm.ui.CenterOnScreen()

	progressBars := make([]*widget.ProgressBar, len(filesToDownloadCopy))

	table := widget.NewTable(
		func() (int, int) {
			return len(filesToDownloadCopy), 2
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
				file := filesToDownloadCopy[id.Row]
				filepath := strings.TrimPrefix(file.Path, dm.settings.LauncherPath)
				fyne.Do(func() {
					label.SetText(filepath)
				})
			} else {
				label.Hide()
				pbar.Show()
				progressBars[id.Row] = pbar
				fyne.Do(func() {
					pbar.SetValue(progressBars[id.Row].Value)
				})
			}
		},
	)

	table.SetColumnWidth(0, 500)
	table.SetColumnWidth(1, 200)

	vscroll := container.NewVScroll(table)
	vscroll.SetMinSize(fyne.NewSize(400, 400))

	dm.ui.SetContent(
		widget.NewLabel(localize.Localize("downloading", nil)),
		container.NewHBox(
			widget.NewLabel(localize.Localize("elapsed_time", nil)),
			timeLabel,
		),
		mainProgressBar,
		vscroll,
	)

	start := time.Now()
	amtFiles := len(filesToDownload)
	processedFiles := 0
	for i := 0; i < len(filesToDownloadCopy); {
		f := filesToDownloadCopy[i]
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
						progressBars[idx].SetValue(float64(currSize) / float64(f.Size))
					})
					fyne.Do(func() {
						timeLabel.SetText(
							fmt.Sprintf(
								"%v (%v/%v)",
								utils.FormatDuration(time.Since(start)),
								processedFiles,
								amtFiles,
							),
						)
					})
				}
				if stop {
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
		}(i, f)

		dm.ui.CenterOnScreen()

		// @TODO: 3 Retries per file
		req, err := http.NewRequest("GET", f.Url, nil)
		if dm.ui.ShowFailedDownloadError(err) {
			return
		}

		SetUserAgent((*dm).settings, req)

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

		processedFiles += 1
		fyne.Do(func() {
			mainProgressBar.SetValue(float64(processedFiles) / float64(len(filesToDownload)))
		})

		filesToDownloadCopy = append(filesToDownloadCopy[:i], filesToDownloadCopy[i+1:]...)
		fyne.Do(func() {
			table.Refresh()
		})
	}

	if dm.onComplete != nil {
		dm.onComplete()
	}
}

func NewDownloadManager(settings *models.BootstrapSettings, ui *ui.MainUi) *DownloadManager {
	return &DownloadManager{
		ui:       ui,
		settings: settings,
	}
}
