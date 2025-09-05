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
