package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//go:embed bs_settings.json
var BOOTSTRAP_SETTINGS_STR []byte

var basepath *string

var BOOTSTRAP_VERSION = "1.0.0"

func init() {
	basepath = flag.String("path", "", "The path to store launcher data (i.e. portable-mode)")
}

func main() {
	flag.Parse()

	app := app.New()
	window := app.NewWindow("SpectrumBootstrap")
	//window.SetFixedSize(true)

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
			settings.LauncherPath, err = filepath.Abs(settings.LauncherPath)
			if ShowError(window, "failed_init", err) {
				return
			}
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

		downloader := Downloader{
			FilesToDownload:   filesToDownload,
			Window:            window,
			ParallelDownloads: 5,
		}

		downloader.Start()

		// Launching the launcher
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
			panic("How did we get here?")
		}

		classpath := []string{}
		for _, f := range launcherManager.launcherManifest.Files {
			if f.Type == "classpath" {
				classpath = append(classpath, filepath.Join(settings.LauncherPath, "launcher", f.Path))
			}
		}

		cmdArgs := []string{
			"-classpath",
			strings.Join(classpath, classpathSeparator),
			launcherManager.launcherManifest.MainClass,
		}

		cmdArgs = append(cmdArgs, launcherManager.SubstituteVariables(jvmManager.GetPath())...)

		cmd := exec.Command(filepath.Join(jvmManager.GetPath(), executablePath), cmdArgs...)

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
		window.Close()
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
