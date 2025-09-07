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
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"

	"github.com/spectrum-mc/bootstrap/httpclient"
	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/utils"
)

type LauncherManager struct {
	LauncherManifest *models.LauncherManifest
	bSettings        *models.BootstrapSettings
}

func GetLauncherManager(bs *models.BootstrapSettings) (*LauncherManager, error) {
	launcherManager := &LauncherManager{
		bSettings: bs,
	}

	// We load the main manifest
	mainManifest, err := httpclient.GetOrCached[models.LauncherManifest](
		filepath.Join(bs.LauncherPath, ".cache", "launcher_manifest.json"),
		bs.ManifestURL,
	)
	if err != nil {
		return nil, err
	}

	launcherManager.LauncherManifest = mainManifest

	return launcherManager, nil
}

func (m *LauncherManager) GetPath() string {
	return path.Join(m.bSettings.LauncherPath, "launcher")
}

// Returns a list of files to re-download
func (m *LauncherManager) ValidateInstallation() ([]models.Downloadable, error) {
	bp := m.GetPath()
	os.MkdirAll(bp, os.ModePerm)

	filesToDownload := []models.Downloadable{}
	fileList := []string{}

	for _, v := range m.LauncherManifest.Files {
		file := filepath.Join(bp, v.Path)
		fileList = append(fileList, file)

		if v.Type == "directory" {
			err := os.MkdirAll(file, os.ModePerm)
			if err != nil {
				return nil, err
			}
		} else if v.Type == "file" || v.Type == "classpath" {
			_, err := os.Stat(file)
			if !os.IsNotExist(err) {
				hash := utils.GetHash(file)
				if hash == v.Hash {
					// The file exists and has the correct hash
					// No need to redownload
					continue
				}
			}

			filesToDownload = append(filesToDownload, models.Downloadable{
				Url:        v.Url,
				Path:       file,
				Sha256:     v.Hash,
				Size:       v.Size,
				Executable: false,
				// @TODO Maybe later, but there should no need to have an executable
				// Unless we want to support Java in other languages
				// Like go which produces direct executables or python
				// Maybe really later
				// This could lead this bootstrap to be more generic
				// instead of a Minecraft focused thing
			})
		}
	}

	// Removing the files that should not exist
	err := filepath.Walk(bp, func(currPath string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		if !slices.Contains(fileList, currPath) {
			fmt.Printf("File / dir %v should not exist. Removing it.\n", currPath)
			if err := os.RemoveAll(currPath); err != nil {
				return err
			}
		}

		return nil
	})

	return filesToDownload, err
}

func (m *LauncherManager) GetCommand(launcherManager *LauncherManager) (*exec.Cmd, error) {
	return nil, errors.New("you should not call GetCommand on the launcher manager, use the runtime manager instead")
}
