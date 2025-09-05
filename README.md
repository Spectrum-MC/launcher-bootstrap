# Launcher Bootstrap

This tool is a bootstrap for your launcher.

Even though it's built for Minecraft, the goal is to be generic-enough to be used for any launcher that requires downloading files and keep them up-to-date, at least for executable / Java / Python apps.

Its what your players will download to run your launcher.

Its goal is to:
1. Download the JVM / Python runtime required by your launcher
2. Update the launcher
3. Start the launcher with the requested runtime

It is made to work with Spectrum and SKCraft. As Spectrum is currently a distant dream it only targets SKCraft.

The bootstrap also features a portable mode that can be started by doing `./bootstrap --path ./launcherfolder`. This will put everything the bootstrap, the launcher then the game is doing in the given folder.

Useful notes:
- The basepath is the "path" argument if it's filled, or the XDG path to `launcher_foldername` otherwise.
- The JVM runtimes are stored at `$basepath/runtime/{component}/{os-arch}`. The launcher can use this folder to store its JVM as long as its in a compatible state. See `jvm_manager.go` if you want to know how they're stored.
- The launcher files are stored at `$basepath/launcher`. This folder is entierly controlled by the bootstrap, don't touch it.

**Note**: While this is made for SKCraft, this won't work properly with the upstream one as it still checks for installed JREs, use [our fork](https://github.com/spectrum-mc/skcraft) for now. [This issue](https://github.com/SKCraft/Launcher/issues/521) relates our effort to upstream it, but for now it's not merged yet.


## Usage

For this guide, we'll make a launcher called "Spectrum Indev", with its related files available on "https://mc.example.com/".

### Setting up the webserver

The webserver needs to have a `launcher_manifest.json` file containing the following:
```json
{
    "version": "v1.0.0",
    "files": [
        {
            "type": "classpath",
            "path": "launcher.jar",
            "hash": "sha256 of the launcher.jar",
            "url": "https://mc.example.com/1.0.0.jar"
        }
    ],
    "main_class": "com.skcraft.launcher.FancyLauncher",
    "args": [
        "--portable",
        "--dir",
        "${rootPath}"
        "--bootstrap-version",
        "${bsVersion}"
    ],
    "jre": {
        "manifest": "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json",
        "component": "java-runtime-gamma"
    }
}
```

Lets dig what's going on there.

- `version`: This represents the version of your launcher, this will be used to compare whether the launcher needs to be updated or not.
- `files`: A list of file to download and how they will be used.
- `files.type`: For now, allowed values are: `directory` => A folder will be created at this path, `file` => The file will be downloaded at this path, `classpath` => Same as file but it will be added to the classpath when running a Java application.
- `files.path`: The path where the file should be downloaded relative to the launcher folder.
- `files.hash`: The sha256 of the file, used to re-download it when corrupted / not completely downloaded / tampered with.
- `files.url`: The path to download your file.
- `main_class`: Only useful for Java softwares, this specifies the main class to be run.
- `jre`: The manifest to know where to download Java and which version to use for the launcher.
- `jre.manifest`: The manifest URL. This one is Mojang's one but you should use the [Java Manifest Builder](https://github.com/spectrum-mc/java-manifest-builder) to download them and provide them from your server.
- `jre.component`: The Java version used. Check the JSON in the `manifest` key to find the correct value here.

The `args` key should be an array letting you specify the argument to the launcher to be used. It features special placeholder variables which will be replaced by the bootstrap when running the final command and should be put like this: `${VARIABLE_NAME}`

Here are the allowed values:
| Value | Description |
|-------|-------------|
| osArch | The os/arch string for downloading a JVM |
| rootPath | The path for your launcher to use as its root |
| bsVersion | The bootstrap version |
| isPortable | Is the bootstrap running in portable mode |

### Building the bootstrap

The bootstrap needs to be compiled for every architecture because you need to embed a file that contains the metadata for it.

Lets clone the repository:
```sh
$ git clone git@github.com:spectrum-mc/launcher-bootstrap.git
$ cd launcher-bootstrap
```

Then create a `bs_settings.json` and let's fill it together:
```json
{
	"launcher_manifest": "https://mc.example.com/launcher_manifest.json",
	"launcher_brand": "Spectrum Indev",
	"launcher_foldername": "spectrumlauncher"
}
```

- `launcher_manifest`: should point to the manifest we created in the previous step
- `launcher_brand`: The name displayed everywhere for your launcher
- `launcher_foldername`: The folder name that will be used

N.B. The folder name tries to respect the XDG specs, thus it will store your launcher and its file to `$HOME/.local/share/launchername` on Linux, `@TODO` on OSX and `%APPDATA%/launchername` on Windows.

Please make sure this file is also accessible on `https://mc.example.com/bs_settings.json`. This is not required but if you can't compile one launcher or the other (I'm talking about osx for no particular reason :unamused:) that your user can do it themselves without having to reverse engineer the executable.

Just before building it, we can change the `Icon.png` file to put our custom icon for the bootstrap.

Finally compile your executables:
```sh
$ go install github.com/fyne-io/fyne-cross@latest # Do it only once, to install fyne-cross
$ fyne-cross windows -arch=amd64,arm64 --app-id fr.oxodao.spectrumbs.YOUR_LAUNCHER_NAME
$ fyne-cross linux -arch=amd64,arm64
$ fyne-cross darwin -macosx-sdk-path="PATH/TO/EXTRACTED/SDK" --arch=amd64,arm64 --app-id fr.oxodao.spectrumbs.YOUR_LAUNCHER_NAME >xc14.3.log # Requires xcode /!\
```

See the [fyne-cross repo](https://github.com/fyne-io/fyne-cross) for more info

**Note**: Currently, fyne-cross do not support building for Windows ARM64 nor OSX ARM64.

Your compiled bootstrap, are available in `fyne-cross/bin/` ready to be uploaded to your webhost to distribute to your players.

## Build

Required packages to build (debian):
- libxxf86vm-dev
- libxcursor-dev
- libxrandr-dev
- libxinerama-dev
- libxi-dev
- libgl1-mesa-dev
- libxrandr-dev

## ROADMAP

- Retry downloads when failed
- Multi-"threaded" download (multi-goroutines)
- Implement Python
- Implement generic executable thing

## License

> Spectrum-Bootstrap - A bootstrap for Minecraft launchers
>
>Copyright (C) 2023-2024 - Oxodao
>
>
> This program is free software: you can redistribute it and/or modify
>
> it under the terms of the GNU General Public License as published by
>
> the Free Software Foundation, either version 3 of the License, or
>
> (at your option) any later version.
>
>
> This program is distributed in the hope that it will be useful,
>
> but WITHOUT ANY WARRANTY; without even the implied warranty of
>
> MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
>
> GNU General Public License for more details.
>
>
> You should have received a copy of the GNU General Public License
>
> along with this program.  If not, see <https://www.gnu.org/licenses/>.
