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
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spectrum-mc/bootstrap/localize"
)

type MainUi struct {
	App        fyne.App
	MainWindow fyne.Window
}

func (ui *MainUi) CenterOnScreen() {
	ui.MainWindow.CenterOnScreen()
}

func (ui *MainUi) SetContent(content ...fyne.CanvasObject) {
	ui.MainWindow.SetContent(container.NewVBox(content...))
}

func (ui *MainUi) showWidgets(widgets ...fyne.CanvasObject) {
	ui.SetContent(widgets...)
	ui.MainWindow.CenterOnScreen()
}

func (ui *MainUi) ShowInfo(localizeKey string, data map[string]string) {
	ui.showWidgets(widget.NewLabel(localize.Localize(localizeKey, data)))
}

func (ui *MainUi) ShowError(localizeKey string, err error) bool {
	if err != nil {
		ui.ShowInfo(localizeKey, map[string]string{"Err": err.Error()})

		return true
	}

	return false
}

func (ui *MainUi) ShowFailedDownloadError(err error) bool {
	if err != nil {
		ui.showWidgets(
			widget.NewLabel(localize.Localize("fail_download", nil)),
			widget.NewLabel(err.Error()),
		)

		return true
	}

	return false
}

func (ui *MainUi) SetTitle(title string) {
	ui.MainWindow.SetTitle(title + " - Bootstrap")
}

func (ui *MainUi) Show() {
	ui.MainWindow.Show()
}

func (ui *MainUi) Hide() {
	ui.MainWindow.Hide()
}

func (ui *MainUi) Start() {
	ui.MainWindow.ShowAndRun()
}

func New() *MainUi {
	app := app.New()

	window := app.NewWindow("SpectrumBootstrap")
	window.SetFixedSize(true)

	return &MainUi{
		App:        app,
		MainWindow: window,
	}
}
