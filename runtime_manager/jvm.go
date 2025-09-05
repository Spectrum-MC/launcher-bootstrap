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
	"path"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/utils"
)

var (
	ErrFailedDetermineOs  = errors.New("failed to determine os/arch")
	ErrNoJavaForOs        = errors.New("no java found for this os")
	ErrNoJavaVersionForOs = errors.New("java version for this os doesn't include the required component")
)

type JvmManager struct {
	cachedMainManifest    *models.MainJavaManifest
	cachedVersionManifest *models.JavaManifest

	launcherManifest models.LauncherJavaManifest
	Os               string
	bSettings        *models.BootstrapSettings
}

func GetJvmManager(bs *models.BootstrapSettings, launcherManifest models.LauncherJavaManifest) (*JvmManager, error) {
	//#region Detecting os
	// runtime.GOARCH = 386 amd64 amd64p32 arm arm64
	os := runtime.GOOS
	arch := runtime.GOARCH
	switch os {
	case "linux":
		os = "linux"
		if arch == "386" {
			os += "-i386"
		} else if arch != "amd64" && arch != "amd64p32" {
			return nil, ErrFailedDetermineOs
		}
	case "darwin":
		os = "mac-os"
		if arch == "arm64" {
			os += "-arm64"
		} else if arch != "amd64" {
			return nil, ErrFailedDetermineOs
		}
	case "windows":
		os = "windows"
		switch arch {
		case "386":
			os += "-x86"
		case "amd64", "amd64p32":
			os += "-x64"
		case "arm64":
			os += "-arm64"
		default:
			return nil, ErrFailedDetermineOs
		}
	default:
		return nil, ErrFailedDetermineOs
	}
	//#endregion

	jvmManager := &JvmManager{
		launcherManifest: launcherManifest,
		bSettings:        bs,
		Os:               os,
	}

	// We load the main manifest
	mainManifest, err := utils.GetOrCached[models.MainJavaManifest](
		bs,
		filepath.Join(bs.LauncherPath, ".cache", "main_java_manifest.json"),
		launcherManifest.ManifestURL,
	)
	if err != nil {
		return nil, err
	}

	jvmManager.cachedMainManifest = mainManifest

	// We load the manifest for the os/version
	versions, ok := (*jvmManager.cachedMainManifest)[os]
	if !ok {
		return nil, ErrNoJavaForOs
	}

	version, ok := versions[launcherManifest.Component]
	if !ok {
		return nil, ErrNoJavaVersionForOs
	}
	versionManifest, err := utils.GetOrCached[models.JavaManifest](
		bs,
		filepath.Join(bs.LauncherPath, ".cache", "java_"+os+"_"+launcherManifest.Component+".json"),
		version[0].Manifest.Url, // @TODO: Check how versions are handled, should we DL the first or the last?
	)
	if err != nil {
		return nil, err
	}

	jvmManager.cachedVersionManifest = versionManifest

	return jvmManager, nil
}

func (m *JvmManager) GetPath() string {
	return path.Join(m.bSettings.LauncherPath, "runtime", m.launcherManifest.Component, m.Os)
}

// Returns a list of files to re-download
func (m *JvmManager) ValidateInstallation() ([]models.Downloadable, error) {
	bp := m.GetPath()

	filesToDownload := []models.Downloadable{}
	fileList := []string{}

	for k, v := range m.cachedVersionManifest.Files {
		file := filepath.Join(bp, k)
		fileList = append(fileList, file)

		if v.Type == "directory" {
			err := os.MkdirAll(file, os.ModePerm)
			if err != nil {
				return nil, err
			}
		} else if v.Type == "file" {
			_, err := os.Stat(file)
			if !os.IsNotExist(err) {
				sha1 := utils.GetHashSha1(file)
				if sha1 == v.Downloads.Raw.Hash {
					// The file exists and has the correct hash
					// No need to redownload

					// Just checking the executable flag
					if v.Executable {
						err := os.Chmod(file, os.ModePerm)
						if err != nil {
							return nil, err
						}
					}
					continue
				}
			}

			filesToDownload = append(filesToDownload, models.Downloadable{
				Url:        v.Downloads.Raw.Url,
				Path:       file,
				Sha1:       v.Downloads.Raw.Hash,
				Size:       v.Downloads.Raw.Size,
				Executable: v.Executable,
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
