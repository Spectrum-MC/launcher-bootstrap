package main

type BootstrapSettings struct {
	ManifestURL string `json:"launcher_manifest"`
	Brand       string `json:"launcher_brand"`
	FolderName  string `json:"launcher_foldername"`

	LauncherPath string `json:"-"`
}

type LauncherVersion struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
}

type ManifestFile struct {
	Type string `json:"type"`
	Path string `json:"path"`
	Hash string `json:"hash"`
	Url  string `json:"url"`
	Size int    `json:"size"`
}

type LauncherJavaManifest struct {
	ManifestURL string `json:"manifest"`
	Component   string `json:"component"`
}

type LauncherManifest struct {
	Version   string               `json:"version"`
	Files     []ManifestFile       `json:"files"`
	MainClass string               `json:"main_class"`
	Args      []string             `json:"app_args"`
	Java      LauncherJavaManifest `json:"jre"`
}

type JavaManifestFileDownload struct {
	Hash string `json:"sha1"`
	Size int    `json:"size"`
	Url  string `json:"url"`
}

type JavaManifestFile struct {
	Type       string `json:"type"`
	Executable bool   `json:"executable"`
	Downloads  struct {
		LZMA JavaManifestFileDownload `json:"lzma"`
		Raw  JavaManifestFileDownload `json:"raw"`
	} `json:"downloads"`
}

type JavaManifest struct {
	Files map[string]JavaManifestFile `json:"files"`
}

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

// mjm["linux"]["java-runtime-gamma"]
type MainJavaManifest map[string]map[string][]MainJavaManifestVersion

type Downloadable struct {
	Url        string
	Path       string
	Sha1       string
	Sha256     string
	Size       int
	Executable bool
}
