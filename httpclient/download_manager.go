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
	filenameLabel := widget.NewLabel("-")
	fileProgressBar := widget.NewProgressBar()

	dm.ui.SetContent(
		widget.NewLabel(localize.Localize("downloading", nil)),
		container.NewHBox(
			widget.NewLabel(localize.Localize("elapsed_time", nil)),
			timeLabel,
		),
		mainProgressBar,
		filenameLabel,
		fileProgressBar,
	)

	start := time.Now()
	amtFiles := len(filesToDownload)
	processedFiles := 0
	for _, f := range filesToDownload {
		err := os.MkdirAll(filepath.Dir(f.Path), os.ModePerm)
		if dm.ui.ShowFailedDownloadError(err) {
			return
		}

		out, err := os.Create(f.Path)
		if dm.ui.ShowFailedDownloadError(err) {
			return
		}

		done := make(chan int64)
		go func(f models.Downloadable) {
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
						fileProgressBar.SetValue(float64(currSize) / float64(f.Size))
					})

					duration := time.Since(start).Round(time.Second)
					hours := duration / time.Hour
					duration -= hours * time.Hour
					minutes := duration / time.Minute
					duration -= minutes * time.Minute
					seconds := duration / time.Second

					fyne.Do(func() {
						timeLabel.SetText(fmt.Sprintf("%02d:%02d:%02d (%v/%v)", hours, minutes, seconds, processedFiles, amtFiles))
					})
				}

				if stop {
					break
				}

				time.Sleep(time.Second)
			}
		}(f)

		dlFilePath := strings.TrimPrefix(
			f.Path,
			dm.settings.LauncherPath,
		)
		if len(dlFilePath) > 20 {
			dlFilePath = "..." + dlFilePath[len(dlFilePath)-20:]
		}
		fyne.Do(func() {
			filenameLabel.SetText(dlFilePath)
		})

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
