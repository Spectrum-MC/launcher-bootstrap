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

package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spectrum-mc/bootstrap/models"
	"github.com/spectrum-mc/bootstrap/utils"
)

var BOOTSTRAP_SETTINGS *models.BootstrapSettings

func GetOrCached[T any](cachePath, url string) (*T, error) {
	cached, cachedErr := LoadFromCache[T](cachePath)
	// There is no error for file not found or file corrupted
	// So if we have an error here, there is a deeper issue and we need to raise
	if cachedErr != nil {
		return nil, cachedErr
	}

	live, liveErr := DoGetRequest[T](url)
	// If we can't get it but the cache is loaded, no issue
	// If we can't get it and no cache: CRASH
	if liveErr != nil && cached != nil {
		return cached, nil
	} else if liveErr != nil {
		return nil, liveErr
	}

	// We got it, lets cache it while we're at it!
	err := os.MkdirAll(filepath.Dir(cachePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, _ := json.MarshalIndent(live, "", "  ")
	_, err = f.Write(data)

	return live, err
}

func LoadFromCache[T interface{}](filepath string) (*T, error) {
	_, err := os.Stat(filepath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if err != nil {
		return nil, nil
	}

	out, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	manifest := new(T)
	err = json.Unmarshal(out, manifest)
	if err != nil {
		// If the file is corrupted
		// We want to download the new one directly
		fmt.Println(err)
		return nil, nil
	}

	return manifest, nil
}

func DoGetRequest[T any](url string) (*T, error) {
	client := &http.Client{}

	req, err := NewRequest(
		"GET",
		url,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	manifest := new(T)
	err = json.NewDecoder(resp.Body).Decode(manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func SetUserAgent(req *http.Request) {
	req.Header.Set(
		"User-Agent",
		BOOTSTRAP_SETTINGS.Brand+" (SpectrumBootstrap v"+utils.BOOTSTRAP_VERSION+", "+runtime.GOOS+", "+runtime.GOARCH+")",
	)
}

func NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	SetUserAgent(req)

	return req, nil
}
