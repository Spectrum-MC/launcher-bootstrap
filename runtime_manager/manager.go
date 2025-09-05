package runtime_manager

import (
	"os/exec"

	"github.com/spectrum-mc/bootstrap/models"
)

type Manager interface {
	GetPath() string
	ValidateInstallation() ([]models.Downloadable, error)
	GetCommand(launcherManager *LauncherManager) (*exec.Cmd, error)
}
