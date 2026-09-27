// AUTO-RECONSTRUCTED TYPES — DOMAIN: app
// 研究用途
package main

type AppEntry struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	EntryType             string   `json:"entryType,omitempty"`
	Icon                  string   `json:"icon"`
	Favorite              bool     `json:"favorite,omitempty"`
	AutoIcon              bool     `json:"autoIcon,omitempty"`
	IconData              string   `json:"iconData,omitempty"`
	IconRef               string   `json:"iconRef,omitempty"`
	IconURL               string   `json:"iconUrl,omitempty"`
	CustomIconData        string   `json:"customIconData,omitempty"`
	CustomIconRef         string   `json:"customIconRef,omitempty"`
	IconDataVersion       int      `json:"iconDataVersion,omitempty"`
	Path                  string   `json:"path"`
	WorkingDir            string   `json:"workingDir"`
	Args                  []string `json:"args"`
	Tags                  []string `json:"tags"`
	LaunchCount           int      `json:"launchCount,omitempty"`
	LastLaunchedAt        string   `json:"lastLaunchedAt,omitempty"`
	LaunchPrivilege       string   `json:"launchPrivilege,omitempty"`
	ShortcutMode          string   `json:"shortcutMode,omitempty"`
	ShortcutPath          string   `json:"shortcutPath,omitempty"`
	ShortcutTargetPath    string   `json:"shortcutTargetPath,omitempty"`
	ShortcutWorkingDir    string   `json:"shortcutWorkingDir,omitempty"`
	ShortcutArgumentsText string   `json:"shortcutArgumentsText,omitempty"`
}

type AppImportPathDescriptor struct {
	Path      string `json:"path"`
	EntryType string `json:"entryType"`
}

type ConsoleItem struct {
	ID         string `json:"id,omitempty"`
	Kind       string `json:"kind"`
	SectionID  string `json:"sectionId"`
	TargetID   string `json:"targetId"`
	SourceID   string `json:"sourceId,omitempty"`
	FolderPath string `json:"folderPath,omitempty"`
	Title      string `json:"title,omitempty"`
	Subtitle   string `json:"subtitle,omitempty"`
	Icon       string `json:"icon,omitempty"`
	IconData   string `json:"iconData,omitempty"`
	IconRef    string `json:"iconRef,omitempty"`
	IconURL    string `json:"iconUrl,omitempty"`
	LayoutX    int    `json:"layoutX,omitempty"`
	LayoutY    int    `json:"layoutY,omitempty"`
	LayoutW    int    `json:"layoutW,omitempty"`
	LayoutH    int    `json:"layoutH,omitempty"`
}

type DetectedLinkBrowser struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type DragLaunchRule struct {
	Pattern      string   `json:"pattern"`
	DefaultAppID string   `json:"defaultAppId,omitempty"`
	AppIDs       []string `json:"appIds,omitempty"`
}

type FileEntry struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Path string   `json:"path"`
	Tags []string `json:"tags"`
}

type LinkBrowser struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Kind string   `json:"kind,omitempty"`
	Path string   `json:"path"`
	Args []string `json:"args,omitempty"`
}

type LinkEntry struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	TitleLocked    bool     `json:"titleLocked,omitempty"`
	Icon           string   `json:"icon"`
	Favorite       bool     `json:"favorite,omitempty"`
	IconMode       string   `json:"iconMode,omitempty"`
	IconData       string   `json:"iconData,omitempty"`
	IconRef        string   `json:"iconRef,omitempty"`
	IconURL        string   `json:"iconUrl,omitempty"`
	URL            string   `json:"url"`
	Browser        string   `json:"browser"`
	Args           []string `json:"args"`
	Tags           []string `json:"tags"`
	LaunchCount    int      `json:"launchCount,omitempty"`
	LastLaunchedAt string   `json:"lastLaunchedAt,omitempty"`
}

type ShortcutResolution struct {
	Path           string `json:"path"`
	TargetPath     string `json:"targetPath"`
	Arguments      string `json:"arguments"`
	WorkingDir     string `json:"workingDir"`
	Description    string `json:"description"`
	IconLocation   string `json:"iconLocation"`
	IconData       string `json:"iconData,omitempty"`
	CustomIconData string `json:"customIconData,omitempty"`
	Resolved       bool   `json:"resolved"`
}

type StartMenuApp struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	TargetPath     string `json:"targetPath,omitempty"`
	Arguments      string `json:"arguments,omitempty"`
	WorkingDir     string `json:"workingDir,omitempty"`
	IconData       string `json:"iconData,omitempty"`
	IconURL        string `json:"iconUrl,omitempty"`
	CustomIconData string `json:"customIconData,omitempty"`
	CustomIconURL  string `json:"customIconUrl,omitempty"`
	Source         string `json:"source"`
	Group          string `json:"group"`
}

type StartMenuLaunchLog struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	ShortcutPath   string `json:"shortcutPath,omitempty"`
	TargetPath     string `json:"targetPath,omitempty"`
	LaunchCount    int    `json:"launchCount,omitempty"`
	LastLaunchedAt string `json:"lastLaunchedAt,omitempty"`
}

type startMenuRoot struct {
	Path   string
	Source string
}
