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
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//go:embed bs_settings.json
var BOOTSTRAP_SETTINGS_STR []byte

var basepath *string

var BOOTSTRAP_VERSION = "1"

func init() {
	basepath = flag.String("path", "", "The path to store launcher data (i.e. portable-mode)")
}

func main() {
	bsVersion, err := strconv.Atoi(BOOTSTRAP_VERSION)
	if err != nil {
		fmt.Println("Failed to parse bootstrap version to an int!")
		fmt.Println("Version found: ", BOOTSTRAP_VERSION)

		panic(err)
	}

	flag.Parse()

	app := app.New()
	window := app.NewWindow("SpectrumBootstrap")
	window.SetFixedSize(true)

	go func() {
		window.SetContent(
			container.NewVBox(
				widget.NewLabel(Localize("fetching_launcher_updates", nil)),
			),
		)
		window.CenterOnScreen()

		settings := BootstrapSettings{}
		err := json.Unmarshal(BOOTSTRAP_SETTINGS_STR, &settings)
		if err != nil {
			window.SetContent(
				container.NewVBox(
					widget.NewLabel(Localize("failed_load_bs_settings", map[string]string{"Err": err.Error()})),
				),
			)
			window.CenterOnScreen()

			return
		}

		if len(*basepath) > 0 {
			settings.LauncherPath = *basepath
		}

		settings.LauncherPath, err = GetLauncherDirectory(&settings)
		if err != nil {
			window.SetContent(
				container.NewVBox(
					widget.NewLabel(Localize("failed_init", map[string]string{"Err": err.Error()})),
				),
			)
			window.CenterOnScreen()

			return
		}

		window.SetTitle(settings.Brand + " - Bootstrap")

		launcherManager, err := GetLauncherManager(&settings)
		if err != nil {
			window.SetContent(
				container.NewVBox(
					widget.NewLabel(Localize("failed_init", map[string]string{"Err": err.Error()})),
				),
			)
			window.CenterOnScreen()

			return
		}

		jvmManager, err := GetJvmManager(&settings, launcherManager.launcherManifest.Java)
		if err != nil {
			window.SetContent(
				container.NewVBox(
					widget.NewLabel(Localize("failed_init", map[string]string{"Err": err.Error()})),
				),
			)
			window.CenterOnScreen()

			return
		}

		jvmFilesToDownload, err := jvmManager.ValidateInstallation()
		if err != nil {
			window.SetContent(
				container.NewVBox(
					widget.NewLabel(Localize("failed_init", map[string]string{"Err": err.Error()})),
				),
			)
			window.CenterOnScreen()

			return
		}

		launcherFilesToDownload, err := launcherManager.ValidateInstallation()
		if err != nil {
			window.SetContent(
				container.NewVBox(
					widget.NewLabel(Localize("failed_init", map[string]string{"Err": err.Error()})),
				),
			)
			window.CenterOnScreen()

			return
		}

		filesToDownload := append(jvmFilesToDownload, launcherFilesToDownload...)

		timeLabel := widget.NewLabel("00:00:00")
		mainProgressBar := widget.NewProgressBar()

		// @TODO Make this base on goroutine to download multiple file at once
		// @TODO which will be hard to display properly like SKCraft
		filenameLabel := widget.NewLabel("-")
		fileProgressBar := widget.NewProgressBar()

		window.SetContent(container.NewVBox(
			widget.NewLabel(Localize("downloading", nil)),
			container.NewHBox(
				widget.NewLabel(Localize("elapsed_time", nil)),
				timeLabel,
			),
			mainProgressBar,
			filenameLabel,
			fileProgressBar,
		))

		start := time.Now()
		amtFiles := len(filesToDownload)
		processedFiles := 0
		for _, f := range filesToDownload {
			err := os.MkdirAll(filepath.Dir(f.Path), os.ModePerm)
			if err != nil {
				window.SetContent(
					container.NewVBox(
						widget.NewLabel(Localize("fail_download", map[string]string{"Err": err.Error()})),
					),
				)
				window.CenterOnScreen()

				return
			}

			out, err := os.Create(f.Path)
			if ShowError(window, "fail_download", err) {
				return
			}

			done := make(chan int64)
			go func(f Downloadable) {
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

						fileProgressBar.SetValue(float64(currSize) / float64(f.Size))

						duration := time.Since(start).Round(time.Second)
						hours := duration / time.Hour
						duration -= hours * time.Hour
						minutes := duration / time.Minute
						duration -= minutes * time.Minute
						seconds := duration / time.Second

						timeLabel.SetText(fmt.Sprintf("%02d:%02d:%02d (%v/%v)", hours, minutes, seconds, processedFiles, amtFiles))
					}

					if stop {
						break
					}

					time.Sleep(time.Second)
				}
			}(f)

			dlFilePath := strings.TrimPrefix(
				f.Path,
				settings.LauncherPath,
			)
			if len(dlFilePath) > 20 {
				dlFilePath = "..." + dlFilePath[len(dlFilePath)-20:]
			}
			filenameLabel.SetText(dlFilePath)

			window.CenterOnScreen()

			// @TODO: 3 Retries per file
			req, err := http.NewRequest("GET", f.Url, nil)
			if ShowError(window, "fail_download", err) {
				return
			}

			req.Header.Set("User-Agent", "SpectrumBootstrap/"+BOOTSTRAP_VERSION)

			resp, err := http.DefaultClient.Do(req)
			if ShowError(window, "fail_download", err) {
				return
			}
			defer resp.Body.Close()

			n, err := io.Copy(out, resp.Body)
			if ShowError(window, "fail_download", err) {
				return
			}

			out.Close()

			if f.Executable {
				err := os.Chmod(f.Path, os.ModePerm)
				if ShowError(window, "fail_download", err) {
					return
				}
			}

			done <- n

			processedFiles += 1
			mainProgressBar.SetValue(float64(processedFiles) / float64(len(filesToDownload)))
		}

		// Launching the launcher
		// @TODO: Handle other than java
		executablePath := ""
		classpathSeparator := ":"
		if runtime.GOOS == "darwin" {
			executablePath = "jre.bundle/Contents/Home/bin/java"
		} else if runtime.GOOS == "linux" {
			executablePath = "bin/java"
		} else if runtime.GOOS == "windows" {
			executablePath = "bin/javaw.exe"
			classpathSeparator = ";"
		} else {
			// I don't currently handle BSD/Solaris/whatever people try to use it on
			panic("How did we get here?")
		}

		classpath := []string{}
		for _, f := range launcherManager.launcherManifest.Files {
			if f.Type == "classpath" {
				classpath = append(classpath, filepath.Join(settings.LauncherPath, "launcher", f.Path))
			}
		}

		variables := map[string]any{
			"osArch":     jvmManager.os,
			"rootPath":   settings.LauncherPath,
			"bsVersion":  bsVersion,
			"isPortable": len(*basepath) > 0,
		}

		cmdStrArr := []string{
			"-classpath",
			strings.Join(classpath, classpathSeparator),
			launcherManager.launcherManifest.MainClass,
		}

		for _, arg := range launcherManager.launcherManifest.Args {
			val := arg

			for k, v := range variables {
				val = strings.ReplaceAll(val, "${"+k+"}", fmt.Sprintf("%v", v))
			}

			cmdStrArr = append(cmdStrArr, val)
		}

		cmd := exec.Command(
			filepath.Join(jvmManager.GetPath(), executablePath),
			cmdStrArr...,
		)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Dir = settings.LauncherPath

		if err = cmd.Start(); err != nil {
			fmt.Println("Failed to run the launcher:")
			fmt.Println(err)
			os.Exit(1)
		}

		window.Hide()

		if err = cmd.Wait(); err != nil {
			fmt.Println("Failed to run the launcher:")
			fmt.Println(err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	window.ShowAndRun()
}

func ShowError(w fyne.Window, translation string, err error) bool {
	if err != nil {
		w.SetContent(
			container.NewVBox(
				widget.NewLabel(Localize(translation, nil)),
				widget.NewLabel(err.Error()),
			),
		)
		w.CenterOnScreen()

		return true
	}

	return false
}
