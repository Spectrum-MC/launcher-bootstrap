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

package localize

import (
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/jeandeaual/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var LocalesFS embed.FS

var localizer *i18n.Localizer = nil

func LoadTranslations(fs embed.FS) {
	LocalesFS = fs
}

func GetLocalizer() *i18n.Localizer {
	userLocales, _ := locale.GetLocales()

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	available_langs := []string{"en", "fr"}

	for _, lang := range available_langs {
		_, err := bundle.LoadMessageFileFS(LocalesFS, "lang/locale."+lang+".toml")
		if err != nil {
			fmt.Println(err)
		}
	}

	return i18n.NewLocalizer(bundle, userLocales...)
}

func Localize(id string, data map[string]string) string {
	if localizer == nil {
		localizer = GetLocalizer()
	}

	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id,
		},
		TemplateData: data,
	})
}
