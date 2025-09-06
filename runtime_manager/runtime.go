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

package runtime_manager

import (
	"errors"

	"github.com/spectrum-mc/bootstrap/models"
)

func GuessRequiredRuntime(settings *models.BootstrapSettings, manifest *models.LauncherManifest) (Manager, error) {
	if manifest.Java != nil {
		return GetJvmManager(settings, *manifest.Java)
	}

	return nil, errors.New("no runtime configuration found in launcher manifest")
}

func BuildRequiredFileDownloadList(settings *models.BootstrapSettings, launcherManager *LauncherManager) (Manager, []models.Downloadable, error) {
	runtimeManager, err := GuessRequiredRuntime(settings, launcherManager.LauncherManifest)
	if err != nil {
		return runtimeManager, nil, err
	}

	runtimeFilesToDownload, err := runtimeManager.ValidateInstallation()
	if err != nil {
		return runtimeManager, nil, err
	}

	launcherFilesToDownload, err := launcherManager.ValidateInstallation()
	if err != nil {
		return runtimeManager, nil, err
	}

	return runtimeManager, append(runtimeFilesToDownload, launcherFilesToDownload...), nil
}
