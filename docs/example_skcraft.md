# Example: SKCraft

As of today, SKCraft cannot download JRE by itself so the game won't run from them and will keep using system ones. This is an issue discussed [here](https://github.com/SKCraft/Launcher/issues/521)

Lets use this as an example anyway.

## Creating a SKCraft launcher

Read the SKCraft launcher and make one working as usual. Nothing specific here.

## Creating the manifest

The boostrap uses a manifest file which lists everything it needs to know. That's a JSON file stored on a publicly accessible URL. Lets say we have a domain name on which we will put files that is called `launcher.example.com`.

```
{
    "version": "4.6-SNAPSHOT",
    "files": [
        {
            "type": "classpath",
            "path": "launcher.jar",
            "hash": "SHA256 OF THE LAUNCHER.JAR",
            "url": "https://launcher.example.com/4.6-SNAPSHOT.jar"
        }
    ],
    "main_class": "com.skcraft.launcher.FancyLauncher",
    "app_args": [
        "--dir",
        "${base_path}",
        "--bootstrap-version",
        "${int_bs_version}"
    ],
    "jre": {
        "manifest": "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json",
        "component": "java-runtime-gamma"
    }
}
```

This example will download a single file (launcher.jar), adding it to the classpath and running the "com.skcraft.launcher.FancyLauncher" class.

It will also download the "java-runtime-gamma", which is Java 17, from Mojang's server (You probably want to upload your own JVM to not abuse their bandwidth).

The `app_args` key refers to what is passed to the launcher. We do not support JRE args nor conditional args (Such as the real launcher is doing to the game) since there is no use case found for them yet. This could change in the future.

Here are the available variables to be used in the arguments:
| Arg name | Info |
|----------|------|
|base_path|The folder in which the game should be installed|
|launcher_path|The launcher folder in the base_path|
|jre_path|The path to the used JRE (e.g. simple launcher that uses the same java version)
|runtime_path|The folder that contains all the installed Java available|
|bs_version|The bootstrap version|
|int_bs_version|The bootstrap major version, as an integer, specifically for SKCraft compat|

The result will be a folder on your user's computer with the following structure:
```
.my_launcher
├─ runtime
|  ├─ java-runtime-gamma
|  |  ├─ linux
|  |  |  ├─ java-runtime-gamma.sha1
|  |  |  ├─ java-runtime-gamma
|  |  |  |  ├─ bin
|  |  |  |  |  ├─ java
|  |  |  |  |  ├─ [...]
|  |  |  |  ├─ lib
|  |  |  |  |  ├─ [...]
|  |  |  |  ├─ [...]
├─ launcher
|  ├─ java_linux_java-runtime-gamma.json (Cache for the files for the java instance)
|  ├─ launcher.jar
|  ├─ launcher_manifest.json (Cache for the launcher manifest)
|  ├─ main_java_manifest.json (Cache for the list of java versions)
```

The allowed file types are: `file`, `directory`, `classpath`.

The path can contain folder names & such (e.g. `lib/javafx.jar`).

## Compiling the bootstrap

You'll need Golang for this. I'm using 1.21 but I'm not sure which features I used so an older one could be ok.

Copy the `bs_settings.json.dist` file to `bs_settings.json` and open it.

Fill the `launcher_brand` with the name of your launcher, lets say `MyLauncher`, the `launcher_foldername` with a slugged name like `my_launcher`. Finally, the `launcher_manifest` key corresponds to the direct url for the JSON file we created in the previous part.

Once everything is ready, try the bootstrap to be sure that it works as intended:
```sh
$ go run .
$ go run . --path ./test_launcher
```

If everything is good you can put your icon as `icon.png` at the root of the project and build it! We'll use a neat too that's called "fyne-cross" for this so that you don't have to handle everything by yourself. It just requires docker installed.

```sh
$ go install github.com/fyne-io/fyne-cross@latest
```

Please be nice to your user and provide executable + JRE for at least the following platforms:
```
windows-x64
osx-x64
osx-arm64
linux-x64
```