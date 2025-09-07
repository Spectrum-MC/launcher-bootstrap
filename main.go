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

package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/spectrum-mc/bootstrap/httpclient"
	"github.com/spectrum-mc/bootstrap/localize"
	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/runtime_manager"
	"github.com/spectrum-mc/bootstrap/ui"
	"github.com/spectrum-mc/bootstrap/utils"
)

//go:embed bs_settings.json
var BOOTSTRAP_SETTINGS_STR []byte

//go:embed lang/locale.*.toml
var LocalesFS embed.FS

var basepath *string

func init() {
	basepath = flag.String("path", "", "The path to store launcher data (i.e. portable-mode)")
}

func main() {
	localize.LoadTranslations(LocalesFS)

	flag.Parse()

	mainUi := ui.New()

	mainUi.App.Lifecycle().SetOnStarted(func() {
		go func() {
			mainUi.ShowInfo("fetching_launcher_updates", nil)

			settings := initializeBootstrap(mainUi)
			if settings == nil {
				os.Exit(1)

				return
			}

			launcherManager, err := runtime_manager.GetLauncherManager(settings)
			if mainUi.ShowError("failed_init", err) {
				return
			}

			runtimeManager, filesToDownload, err := runtime_manager.BuildRequiredFileDownloadList(settings, launcherManager)
			if mainUi.ShowError("failed_init", err) {
				return
			}

			dm := httpclient.NewDownloadManager(mainUi)
			dm.SetOnComplete(func() {
				err = runLauncher(runtimeManager, launcherManager, settings, mainUi)
				if err != nil {
					// @TODO: If it fails it should show a GUI message instead of this
					// A new window
					fmt.Println("Failed to run the launcher:")
					fmt.Println(err)
					os.Exit(1)
				}

				os.Exit(0)
			})

			go func() {
				dm.Download(filesToDownload)
			}()
		}()
	})

	mainUi.Start()
}

func initializeBootstrap(mainUi *ui.MainUi) *models.BootstrapSettings {
	settings := models.BootstrapSettings{}
	err := json.Unmarshal(BOOTSTRAP_SETTINGS_STR, &settings)
	if mainUi.ShowError("failed_load_bs_settings", err) {
		return nil
	}

	httpclient.BOOTSTRAP_SETTINGS = &settings

	bsVersion, err := strconv.Atoi(utils.BOOTSTRAP_VERSION)
	if err != nil {
		fmt.Println("Failed to parse bootstrap version to an int!")
		fmt.Println("Version found: ", utils.BOOTSTRAP_VERSION)

		panic(err)
	}

	settings.BootstrapVersion = bsVersion
	settings.Portable = false

	if len(*basepath) > 0 {
		settings.LauncherPath = *basepath
		settings.Portable = true
	}

	settings.LauncherPath, err = utils.GetLauncherDirectory(&settings)
	if mainUi.ShowError("failed_init", err) {
		return nil
	}

	mainUi.SetTitle(settings.Brand)

	return &settings
}

func runLauncher(
	runtimeManager runtime_manager.Manager,
	launcherManager *runtime_manager.LauncherManager,
	settings *models.BootstrapSettings,
	mainUi *ui.MainUi,
) error {
	cmd, err := runtimeManager.GetCommand(launcherManager)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = settings.LauncherPath

	if err = cmd.Start(); err != nil {
		return err
	}

	mainUi.Hide()

	return cmd.Wait()
}
