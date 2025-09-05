//go:build !linux

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

package utils

import (
	"os"

	"github.com/kirsle/configdir"
	"github.com/spectrum-mc/bootstrap/models"
)

func GetLauncherDirectory(s *models.BootstrapSettings) (string, error) {
	lp := s.LauncherPath
	if len(lp) == 0 {
		lp = configdir.LocalConfig(s.FolderName)
	}

	_, err := os.Stat(lp)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	} else if err != nil {
		err = os.MkdirAll(lp, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return lp, nil
}
