# Spectrum Bootstrap

This project is the bootstrap for the launcher. While it is made for Minecraft, it can theorically be used with any Java software and easily adapted for other languages.

## Basic overview

The bootstrap aims at doing a few things:
- Pre-install Java to the launcher directory
- Install & update the launcher (Can be multiple files)
- Start it
- Be able to be used in portable mode or standard mode
- Be working on Linux, Mac OS and Windows
- Work offline (After the initial download of course)
- Parallel downloads (Hardcoded to 5 files at the same time for now)
- Retry on failure (Not implemented yet)

## How does it do this ?

We use the `//go:embed` directive to embed a settings file in the executable. This contains a few useful informations:
- Launcher brand: The name that is displayed to the user
- Launcher folder name: The name of the folder in which the launcher is installed
- Launcher manifest: The link to the launcher manifest

The launcher manifest is a file stored on a publicly accessible URL. It tells the launcher what to do with a few parameters:
- Version: the currently available launcher version
- JRE: the url for a Mojang-like formatted URL to download the Java Virtual Machine alongside with which version should be used
- Files: A list of file to be downloaded. Those can be either "file" for a simple file, "directory" for a simple folder, or classpath for files that should be put in the classpath. This lets you use JavaFX as it doesn't like being fat-jarred.
- Main class: The class to run

Note: these files are revalidated on each startup, so you might not want to have a lot of file (sha256 of each one every startup) and you don't want to have user-ediable files in there (They will be replaced). This could change in the future if we have a use-case but thats not planned for now.

## TODO

As of now the bootstrap is working-ish.

It has been only tested on Linux and do not support symlinks (some file in the JRE are symlinks and are not made, but this does not seem to impact for now).

The UI framework is a bit of a pain due to how manual it is so the list of downloaded file is ugly and badly refreshed. This will have to be dealt with, so that used have a correct view of what's being done.