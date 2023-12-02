package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UiFileDownload struct {
	Progress float64
	File     Downloadable

	Started         bool
	Errored         string
	Completed       bool
	progressChannel chan (int64)
}

func (ui *UiFileDownload) updateUi(table *widget.Table) {
	var stop bool = false

	for {
		select {
		case <-ui.progressChannel:
			stop = true
		default:
			fi, err := os.Stat(ui.File.Path)
			if err == nil {
				currSize := fi.Size()
				if currSize == 0 {
					currSize = 1
				}

				ui.Progress = float64(currSize) / float64(ui.File.Size)
				table.Refresh()
			} else if !os.IsNotExist(err) {
				log.Fatal(err)
			}
		}

		if stop {
			fi, err := os.Stat(ui.File.Path)
			if err == nil {
				currSize := fi.Size()
				if currSize == 0 {
					currSize = 1
				}

				ui.Progress = float64(currSize) / float64(ui.File.Size)
				table.Refresh()
			}
			break
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (ui *UiFileDownload) ShowError(err error) bool {
	if err == nil {
		return false
	}

	ui.Errored = Localize("fail_download", nil) + err.Error()
	ui.Completed = true
	ui.progressChannel <- 0

	return true
}

func (ui *UiFileDownload) Download(table *widget.Table) {
	ui.progressChannel = make(chan int64)
	go ui.updateUi(table)

	err := os.MkdirAll(filepath.Dir(ui.File.Path), os.ModePerm)
	if ui.ShowError(err) {
		return
	}

	out, err := os.Create(ui.File.Path)
	if ui.ShowError(err) {
		return
	}

	req, err := http.NewRequest("GET", ui.File.Url, nil)
	if ui.ShowError(err) {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if ui.ShowError(err) {
		return
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if ui.ShowError(err) {
		return
	}

	out.Close()

	// @TODO: Verify hash & restart if not ok, max 3 times

	if ui.File.Executable {
		err := os.Chmod(ui.File.Path, os.ModePerm)
		if ui.ShowError(err) {
			return
		}
	}

	ui.Completed = true
	ui.progressChannel <- n
}

func worker(dlChan chan *UiFileDownload, wg *sync.WaitGroup, table *widget.Table) {
	defer wg.Done()

	for dl := range dlChan {
		dl.Download(table)
	}
}

type Downloader struct {
	FilesToDownload   []Downloadable
	Window            fyne.Window
	ParallelDownloads int

	startedAt time.Time
	running   bool

	timeLabel       *widget.Label
	mainProgressBar *widget.ProgressBar
	uiFileDownloads []*UiFileDownload
}

func (d *Downloader) Start() {
	d.timeLabel = widget.NewLabel("00:00:00")
	d.mainProgressBar = widget.NewProgressBar()

	for _, f := range d.FilesToDownload {
		d.uiFileDownloads = append(d.uiFileDownloads, &UiFileDownload{
			File:      f,
			Started:   false,
			Completed: false,
		})
	}

	clearDoneDownloads := func() {
		downloads := []*UiFileDownload{}
		for _, dl := range d.uiFileDownloads {
			if dl.Progress < 1.0 && len(dl.Errored) == 0 {
				downloads = append(downloads, dl)
			}
		}
		d.uiFileDownloads = downloads
	}

	// @TODO Make this base on goroutine to download multiple file at once

	d.startedAt = time.Now()
	d.running = true
	go d.updateUi(clearDoneDownloads)

	table := widget.NewTable(
		func() (int, int) {
			val := len(d.uiFileDownloads)
			if val > 15 {
				val = 15
			}

			return 15, 3
		},
		func() fyne.CanvasObject {
			return container.NewStack(widget.NewLabel("filepath"), widget.NewProgressBar(), widget.NewLabel("err"))
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			if id.Row >= len(d.uiFileDownloads) {
				return
			}

			// Wtf is this library seriously
			// This is the RECOMMENDED way of doing it
			// https://fynelabs.com/2022/12/30/building-complex-tables-with-fyne/
			path := o.(*fyne.Container).Objects[0].(*widget.Label)
			progress := o.(*fyne.Container).Objects[1].(*widget.ProgressBar)
			err := o.(*fyne.Container).Objects[2].(*widget.Label)

			switch id.Col {
			case 0:
				path.Show()
				progress.Hide()
				err.Hide()

				path.SetText(d.uiFileDownloads[id.Row].File.Path)
			case 1:
				path.Hide()
				progress.Show()
				err.Hide()

				progress.SetValue(d.uiFileDownloads[id.Row].Progress)
			case 2:
				path.Hide()
				progress.Hide()
				err.Show()

				err.SetText(d.uiFileDownloads[id.Row].Errored)
			}
		},
	)

	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 100)
	table.SetColumnWidth(2, 200)
	table.ShowHeaderRow = true
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		switch id.Col {
		case 0:
			template.(*widget.Label).SetText("File")
		case 1:
			template.(*widget.Label).SetText("Progress")
		case 2:
			template.(*widget.Label).SetText("Error")
		}
	}

	d.Window.SetContent(container.NewBorder(
		container.NewVBox(
			widget.NewLabel(Localize("downloading", nil)),
			container.NewHBox(
				widget.NewLabel(Localize("elapsed_time", nil)),
				d.timeLabel,
			),
			d.mainProgressBar,
		),
		nil,
		nil,
		nil,
		table,
	))
	d.Window.Resize(fyne.NewSize(560, 500))
	d.Window.CenterOnScreen()

	dlChan := make(chan *UiFileDownload)
	wg := new(sync.WaitGroup)

	for i := 0; i < d.ParallelDownloads; i++ {
		wg.Add(1)
		go worker(dlChan, wg, table)
	}

	for _, f := range d.uiFileDownloads {
		// f.Download(table)
		dlChan <- f
	}

	close(dlChan)
	wg.Wait()

	d.running = false
}

func (d *Downloader) updateUi(clearDoneDownloads func()) {
	for d.running {
		duration := time.Since(d.startedAt).Round(time.Second)
		hours := duration / time.Hour
		duration -= hours * time.Hour
		minutes := duration / time.Minute
		duration -= minutes * time.Minute
		seconds := duration / time.Second

		amtFiles := len(d.FilesToDownload)

		// @TODO: Maybe use channels to skip the loop every seconds
		clearDoneDownloads()
		processedFiles := len(d.FilesToDownload) - len(d.uiFileDownloads)

		d.timeLabel.SetText(fmt.Sprintf("%02d:%02d:%02d (%v/%v)", hours, minutes, seconds, processedFiles, amtFiles))
		d.mainProgressBar.SetValue(float64(float64(processedFiles) / float64(amtFiles)))

		time.Sleep(250 * time.Millisecond)
	}
}
