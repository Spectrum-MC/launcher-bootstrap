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

package models

type BootstrapSettings struct {
	BootstrapVersion int  `json:"-"`
	Portable         bool `json:"-"`

	ManifestURL string `json:"launcher_manifest"`
	Brand       string `json:"launcher_brand"`
	FolderName  string `json:"launcher_foldername"`

	LauncherPath string `json:"-"`
}
