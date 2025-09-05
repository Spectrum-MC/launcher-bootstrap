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

// mjm["linux"]["java-runtime-gamma"]
type MainJavaManifest map[string]map[string][]MainJavaManifestVersion

type MainJavaManifestVersion struct {
	Availability struct {
		Group    int `json:"group"`
		Progress int `json:"progress"`
	} `json:"availability"`
	Manifest JavaManifestFileDownload `json:"manifest"`
	Version  struct {
		Name     string `json:"name"`
		Released string `json:"released"`
	} `json:"version"`
}
