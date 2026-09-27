// AUTO-RECONSTRUCTED TYPES（研究用途）
package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/text/collate"

	"modernc.org/sqlite"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"
	"github.com/wailsapp/go-webview2/pkg/combridge"
)

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

type AudioDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Flow      string `json:"flow"`
	IsDefault bool   `json:"isDefault"`
}

type AudioSession struct {
	ID                     string `json:"id"`
	AppName                string `json:"appName"`
	ProcessID              uint32 `json:"processId"`
	ProcessPath            string `json:"processPath"`
	IconData               string `json:"iconData"`
	VolumePercent          int    `json:"volumePercent"`
	Muted                  bool   `json:"muted"`
	Active                 bool   `json:"active"`
	SystemSounds           bool   `json:"systemSounds"`
	OutputDeviceID         string `json:"outputDeviceId"`
	InputDeviceID          string `json:"inputDeviceId"`
	DeviceRoutingSupported bool   `json:"deviceRoutingSupported"`
}

type AudioState struct {
	OutputDevices                    []AudioDevice  `json:"outputDevices"`
	InputDevices                     []AudioDevice  `json:"inputDevices"`
	DefaultOutputDeviceID            string                 `json:"defaultOutputDeviceId"`
	DefaultInputDeviceID             string                 `json:"defaultInputDeviceId"`
	DefaultOutputDeviceVolumePercent int                    `json:"defaultOutputDeviceVolumePercent"`
	DefaultInputDeviceVolumePercent  int                    `json:"defaultInputDeviceVolumePercent"`
	DefaultOutputDeviceMuted         bool                   `json:"defaultOutputDeviceMuted"`
	DefaultInputDeviceMuted          bool                   `json:"defaultInputDeviceMuted"`
	Sessions                         []AudioSession `json:"sessions"`
}

type BackgroundPreference struct {
	Enabled                   bool     `json:"enabled,omitempty"`
	ImagePath                 string   `json:"imagePath,omitempty"`
	ImageURL                  string   `json:"imageUrl,omitempty"`
	Layout                    string   `json:"layout,omitempty"`
	ReadabilityOverlayEnabled *bool    `json:"readabilityOverlayEnabled,omitempty"`
	ReadabilityOverlayOpacity *float64 `json:"readabilityOverlayOpacity,omitempty"`
	ImageOpacity              *float64 `json:"imageOpacity,omitempty"`
	ImageBlur                 int      `json:"imageBlur,omitempty"`
	SidebarTransparency       *float64 `json:"sidebarTransparency,omitempty"`
	ContentTransparency       *float64 `json:"contentTransparency,omitempty"`
	ShellOpacity              *float64 `json:"shellOpacity,omitempty"`
}

type BookmarkFavoritePath struct {
	SourcePath string `json:"sourcePath"`
	FolderPath string `json:"folderPath"`
}

type BookmarkSection struct {
	Custom  []LinkEntry      `json:"custom"`
	Sources []BookmarkSource `json:"sources"`
}

type BookmarkSource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Browser string `json:"browser"`
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

type BookmarkState struct {
	Sources []ResolvedBookmarkSource `json:"sources"`
	Items   []ResolvedBookmarkItem   `json:"items"`
}

type BootstrapSnapshot struct {
	Workspace WorkspaceLayout    `json:"workspace"`
	Languages []LanguageManifest `json:"languages"`
	Plugins   []PluginManifest   `json:"plugins"`
}

type ConfigImportPreview struct {
	SourceFile      string                  `json:"sourceFile"`
	SourceDirectory string                  `json:"sourceDirectory"`
	Workspace       WorkspaceLayout `json:"workspace"`
}

type ConfigMigrationResult struct {
	OldConfig       string `json:"oldConfig"`
	NewConfig       string `json:"newConfig"`
	OldDataRoot     string `json:"oldDataRoot,omitempty"`
	NewDataRoot     string `json:"newDataRoot,omitempty"`
	RestartRequired bool   `json:"restartRequired"`
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

type DesktopAirQuality struct {
	Standard         string  `json:"standard"`
	Value            int     `json:"value"`
	Category         string  `json:"category"`
	PrimaryPollutant string  `json:"primaryPollutant,omitempty"`
	PM25             float64 `json:"pm2_5"`
	PM10             float64 `json:"pm10"`
	ObservedAt       string  `json:"observedAt"`
}

type DesktopCalendarDay struct {
	Date           string   `json:"date"`
	Day            int      `json:"day"`
	Weekday        int      `json:"weekday"`
	LunarYear      string   `json:"lunarYear"`
	LunarMonth     string   `json:"lunarMonth"`
	LunarDay       string   `json:"lunarDay"`
	SolarTerm      string   `json:"solarTerm,omitempty"`
	SolarFestivals []string `json:"solarFestivals,omitempty"`
	LunarFestivals []string `json:"lunarFestivals,omitempty"`
}

type DesktopCalendarMonth struct {
	Year      int                          `json:"year"`
	Month     int                          `json:"month"`
	ServerNow string                       `json:"serverNow"`
	Days      []DesktopCalendarDay `json:"days"`
}

type DesktopNote struct {
	WidgetID  string `json:"widgetId"`
	Body      string `json:"body"`
	Revision  int64  `json:"revision"`
	UpdatedAt string `json:"updatedAt"`
}

type DesktopNoteDraftInput struct {
	ID               string `json:"id"`
	Body             string `json:"body"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

type DesktopNotificationRecord struct {
	DeliveryKey string `json:"deliveryKey"`
	EntityID    string `json:"entityId"`
	DueAtUtc    string `json:"dueAtUtc"`
	AttemptedAt string `json:"attemptedAt,omitempty"`
	Status      string `json:"status"`
	ErrorCode   string `json:"errorCode,omitempty"`
}

type DesktopProtectedSecret struct {
	Provider      string            `json:"provider"`
	SecretKind    string            `json:"secretKind"`
	ProtectedBlob string            `json:"protectedBlob"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	UpdatedAt     string            `json:"updatedAt"`
}

type DesktopReminder struct {
	WidgetID        string `json:"widgetId"`
	Enabled         bool   `json:"enabled"`
	ScheduleKind    string `json:"scheduleKind"`
	IntervalSeconds int64  `json:"intervalSeconds,omitempty"`
	DueAtUtc        string `json:"dueAtUtc,omitempty"`
	TimeOfDay       string `json:"timeOfDay,omitempty"`
	DayOfMonth      int    `json:"dayOfMonth,omitempty"`
	TimezoneID      string `json:"timezoneId,omitempty"`
	ActiveStart     string `json:"activeStart,omitempty"`
	ActiveEnd       string `json:"activeEnd,omitempty"`
	DaysMask        int    `json:"daysMask,omitempty"`
	NextDueAtUtc    string `json:"nextDueAtUtc,omitempty"`
	SnoozedUntilUtc string `json:"snoozedUntilUtc,omitempty"`
	LastFiredAtUtc  string `json:"lastFiredAtUtc,omitempty"`
	CompletedAtUtc  string `json:"completedAtUtc,omitempty"`
	Status          string `json:"status"`
	Revision        int64  `json:"revision"`
}

type DesktopReminderDefinition struct {
	ScheduleKind     string `json:"scheduleKind"`
	IntervalSeconds  int64  `json:"intervalSeconds,omitempty"`
	DueAtUtc         string `json:"dueAtUtc,omitempty"`
	TimeOfDay        string `json:"timeOfDay,omitempty"`
	DayOfMonth       int    `json:"dayOfMonth,omitempty"`
	TimezoneID       string `json:"timezoneId,omitempty"`
	ActiveStart      string `json:"activeStart,omitempty"`
	ActiveEnd        string `json:"activeEnd,omitempty"`
	DaysMask         int    `json:"daysMask,omitempty"`
	Preset           string `json:"preset,omitempty"`
	AudioMode        string `json:"audioMode,omitempty"`
	AudioPath        string `json:"audioPath,omitempty"`
	AudioRepeatCount int    `json:"audioRepeatCount,omitempty"`
}

type DesktopReminderNotificationTestInput struct {
	WidgetID         string `json:"widgetId"`
	Title            string `json:"title,omitempty"`
	Message          string `json:"message,omitempty"`
	AudioMode        string `json:"audioMode,omitempty"`
	AudioPath        string `json:"audioPath,omitempty"`
	AudioRepeatCount int    `json:"audioRepeatCount,omitempty"`
}

type DesktopStopwatch struct {
	WidgetID      string `json:"widgetId"`
	AccumulatedMs int64  `json:"accumulatedMs"`
	Status        string `json:"status"`
	StartedAtUtc  string `json:"startedAtUtc,omitempty"`
	Revision      int64  `json:"revision"`
}

type DesktopStopwatchLap struct {
	Sequence  int    `json:"sequence"`
	ElapsedMs int64  `json:"elapsedMs"`
	CreatedAt string `json:"createdAt"`
}

type DesktopTimer struct {
	WidgetID       string `json:"widgetId"`
	DurationMs     int64  `json:"durationMs"`
	RemainingMs    int64  `json:"remainingMs"`
	Status         string `json:"status"`
	StartedAtUtc   string `json:"startedAtUtc,omitempty"`
	DueAtUtc       string `json:"dueAtUtc,omitempty"`
	CompletedAtUtc string `json:"completedAtUtc,omitempty"`
	NotifiedAtUtc  string `json:"notifiedAtUtc,omitempty"`
	Revision       int64  `json:"revision"`
}

type DesktopWeatherCurrent struct {
	ObservedAt    string  `json:"observedAt"`
	Temperature   float64 `json:"temperature"`
	ApparentTemp  float64 `json:"apparentTemperature"`
	Humidity      int     `json:"humidity"`
	ConditionCode int     `json:"conditionCode"`
	WindSpeed     float64 `json:"windSpeed"`
}

type DesktopWeatherDay struct {
	Date              string  `json:"date"`
	MinTemperature    float64 `json:"minTemperature"`
	MaxTemperature    float64 `json:"maxTemperature"`
	ConditionCode     int     `json:"conditionCode"`
	PrecipProbability int     `json:"precipProbability"`
	UVMax             float64 `json:"uvMax"`
	Sunrise           string  `json:"sunrise,omitempty"`
	Sunset            string  `json:"sunset,omitempty"`
}

type DesktopWeatherDomainState struct {
	Provider  string `json:"provider"`
	FetchedAt string `json:"fetchedAt,omitempty"`
	Stale     bool   `json:"stale"`
	ErrorCode string `json:"errorCode,omitempty"`
}

type DesktopWeatherHour struct {
	Time              string  `json:"time"`
	Temperature       float64 `json:"temperature"`
	ConditionCode     int     `json:"conditionCode"`
	PrecipProbability int     `json:"precipProbability"`
	PrecipAmount      float64 `json:"precipAmount"`
	UV                float64 `json:"uv"`
}

type DesktopWeatherLocation struct {
	ID              string  `json:"id"`
	DisplayName     string  `json:"displayName"`
	DisplayLanguage string  `json:"displayLanguage,omitempty"`
	CountryCode     string  `json:"countryCode,omitempty"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TimezoneID      string  `json:"timezoneId"`
}

type DesktopWeatherLocationResult struct {
	ID              string  `json:"id"`
	DisplayName     string  `json:"displayName"`
	DisplayLanguage string  `json:"displayLanguage,omitempty"`
	CountryCode     string  `json:"countryCode,omitempty"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TimezoneID      string  `json:"timezoneId"`
	Provider        string  `json:"provider"`
}

type DesktopWeatherProviderCredentialInput struct {
	Provider   string            `json:"provider"`
	SecretKind string            `json:"secretKind"`
	Secret     string            `json:"secret"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type DesktopWeatherProviderState struct {
	Provider   string            `json:"provider"`
	Configured bool              `json:"configured"`
	SecretKind string            `json:"secretKind,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	UpdatedAt  string            `json:"updatedAt,omitempty"`
	Available  bool              `json:"available"`
	LastError  string            `json:"lastError,omitempty"`
}

type DesktopWeatherProviderStatus struct {
	Providers []DesktopWeatherProviderState `json:"providers"`
}

type DesktopWeatherSnapshot struct {
	WidgetID    string                                       `json:"widgetId,omitempty"`
	Location    DesktopWeatherLocation               `json:"location"`
	Current     DesktopWeatherCurrent                `json:"current"`
	Hourly      []DesktopWeatherHour                 `json:"hourly"`
	Daily       []DesktopWeatherDay                  `json:"daily"`
	AirQuality  DesktopAirQuality                    `json:"airQuality"`
	Attribution []string                                     `json:"attribution"`
	Domains     map[string]DesktopWeatherDomainState `json:"domains"`
	FetchedAt   string                                       `json:"fetchedAt"`
	ExpiresAt   string                                       `json:"expiresAt"`
	StaleUntil  string                                       `json:"staleUntil"`
}

type DesktopWidget struct {
	ID             string                      `json:"id"`
	Type           string                      `json:"type"`
	Title          string                      `json:"title"`
	Config         DesktopWidgetConfig `json:"config"`
	Revision       int64                       `json:"revision"`
	LifecycleState string                      `json:"lifecycleState,omitempty"`
	CreatedAt      string                      `json:"createdAt"`
	UpdatedAt      string                      `json:"updatedAt"`
}

type DesktopWidgetConfig struct {
	WeatherLocation      *DesktopWeatherLocation    `json:"weatherLocation,omitempty"`
	WeatherProviderOrder []string                           `json:"weatherProviderOrder,omitempty"`
	Timezones            []DesktopWorldClockZone    `json:"timezones,omitempty"`
	DurationMs           int64                              `json:"durationMs,omitempty"`
	Reminder             *DesktopReminderDefinition `json:"reminder,omitempty"`
	HideTitleOnHome      bool                               `json:"hideTitleOnHome,omitempty"`
	HourFormat           string                             `json:"hourFormat,omitempty"`
}

type DesktopWidgetCreateInput struct {
	Type   string                      `json:"type"`
	Title  string                      `json:"title"`
	Config DesktopWidgetConfig `json:"config"`
}

type DesktopWidgetDocument struct {
	Version                int                                          `json:"version"`
	Revision               int64                                        `json:"revision"`
	UpdatedAt              string                                       `json:"updatedAt"`
	Widgets                map[string]DesktopWidget             `json:"widgets"`
	Notes                  map[string]DesktopNote               `json:"notes"`
	Reminders              map[string]DesktopReminder           `json:"reminders"`
	Timers                 map[string]DesktopTimer              `json:"timers"`
	Stopwatches            map[string]DesktopStopwatch          `json:"stopwatches"`
	StopwatchLaps          map[string][]DesktopStopwatchLap     `json:"stopwatchLaps"`
	WeatherCache           map[string]DesktopWeatherSnapshot    `json:"weatherCache"`
	NotificationDeliveries map[string]DesktopNotificationRecord `json:"notificationDeliveries"`
	ProtectedSecrets       map[string]DesktopProtectedSecret    `json:"protectedSecrets"`
}

type DesktopWidgetSnapshot struct {
	Status               string                                       `json:"status"`
	ErrorCode            string                                       `json:"errorCode,omitempty"`
	ErrorMessage         string                                       `json:"errorMessage,omitempty"`
	Revision             int64                                        `json:"revision"`
	ServerNow            string                                       `json:"serverNow"`
	Widgets              []DesktopWidget                      `json:"widgets"`
	Notes                map[string]DesktopNote               `json:"notes"`
	Reminders            map[string]DesktopReminder           `json:"reminders"`
	Timers               map[string]DesktopTimer              `json:"timers"`
	Stopwatches          map[string]DesktopStopwatch          `json:"stopwatches"`
	StopwatchLaps        map[string][]DesktopStopwatchLap     `json:"stopwatchLaps"`
	Weather              map[string]DesktopWeatherSnapshot    `json:"weather"`
	NotificationFailures map[string]DesktopNotificationRecord `json:"notificationFailures"`
}

type DesktopWidgetUpdateInput struct {
	ID               string                      `json:"id"`
	Title            string                      `json:"title"`
	Config           DesktopWidgetConfig `json:"config"`
	ExpectedRevision int64                       `json:"expectedRevision"`
}

type DesktopWorldClockSnapshot struct {
	ServerNow string                          `json:"serverNow"`
	Clocks    []DesktopWorldClockTime `json:"clocks"`
}

type DesktopWorldClockTime struct {
	TimezoneID    string `json:"timezoneId"`
	Label         string `json:"label"`
	LocalTime     string `json:"localTime"`
	Offset        string `json:"offset"`
	OffsetSeconds int    `json:"offsetSeconds"`
	DST           bool   `json:"dst"`
}

type DesktopWorldClockZone struct {
	TimezoneID string `json:"timezoneId"`
	Label      string `json:"label"`
}

type DetectedBookmarkSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BrowserName string `json:"browserName"`
	Profile     string `json:"profile"`
	Path        string `json:"path"`
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

type FileLocatorConfig struct {
	FileNameQuery         string                      `json:"fileNameQuery,omitempty"`
	FileNameMode          string                      `json:"fileNameMode,omitempty"`
	FileNameMatchCase     bool                        `json:"fileNameMatchCase,omitempty"`
	FileNameHistory       []string                    `json:"fileNameHistory,omitempty"`
	ContainsTextQuery     string                      `json:"containsTextQuery,omitempty"`
	ContainsTextMode      string                      `json:"containsTextMode,omitempty"`
	ContainsTextMatchCase bool                        `json:"containsTextMatchCase,omitempty"`
	ContainsTextHistory   []string                    `json:"containsTextHistory,omitempty"`
	BooleanScope          string                      `json:"booleanScope,omitempty"`
	SearchRoot            string                      `json:"searchRoot,omitempty"`
	SearchRootHistory     []string                    `json:"searchRootHistory,omitempty"`
	MaxSearchFileSizeMB   int                         `json:"maxSearchFileSizeMB,omitempty"`
	IncludeSubfolders     bool                        `json:"includeSubfolders"`
	ActiveFilterID        string                      `json:"activeFilterId,omitempty"`
	SavedFilters          []FileLocatorFilter `json:"savedFilters,omitempty"`
}

type FileLocatorFilter struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
	Value  string `json:"value"`
	Type   string `json:"type"`
}

type FileLocatorLineMatch struct {
	LineNumber int                              `json:"lineNumber"`
	Text       string                           `json:"text"`
	Ranges     []FileLocatorTextRange   `json:"ranges"`
	Before     []FileLocatorPreviewLine `json:"before"`
	After      []FileLocatorPreviewLine `json:"after"`
}

type FileLocatorPreviewLine struct {
	LineNumber int                            `json:"lineNumber"`
	Text       string                         `json:"text"`
	Ranges     []FileLocatorTextRange `json:"ranges,omitempty"`
}

type FileLocatorResultDetail struct {
	Path          string                         `json:"path"`
	Name          string                         `json:"name"`
	Directory     string                         `json:"directory"`
	Size          int64                          `json:"size"`
	ModifiedAt    string                         `json:"modifiedAt"`
	LineCount     int                            `json:"lineCount"`
	MatchCount    int                            `json:"matchCount"`
	StoredMatches int                            `json:"storedMatches"`
	Truncated     bool                           `json:"truncated"`
	TextAvailable bool                           `json:"textAvailable"`
	Reason        string                         `json:"reason"`
	Matches       []FileLocatorLineMatch `json:"matches"`
}

type FileLocatorResultSummary struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Directory  string `json:"directory"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
	MatchCount int    `json:"matchCount"`
	Preview    string `json:"preview"`
}

type FileLocatorState struct {
	Request              FileLocatorConfig          `json:"request"`
	Running              bool                               `json:"running"`
	Paused               bool                               `json:"paused"`
	Completed            bool                               `json:"completed"`
	Cancelled            bool                               `json:"cancelled"`
	StartedAt            string                             `json:"startedAt"`
	FinishedAt           string                             `json:"finishedAt"`
	ElapsedMilliseconds  int64                              `json:"elapsedMilliseconds"`
	CurrentPath          string                             `json:"currentPath"`
	CheckedItemCount     int                                `json:"checkedItemCount"`
	CheckedBytes         int64                              `json:"checkedBytes"`
	SearchedItemCount    int                                `json:"searchedItemCount"`
	SearchedBytes        int64                              `json:"searchedBytes"`
	MatchedFileCount     int                                `json:"matchedFileCount"`
	MatchedBytes         int64                              `json:"matchedBytes"`
	TextMatchCount       int                                `json:"textMatchCount"`
	SkippedBinaryCount   int                                `json:"skippedBinaryCount"`
	SkippedOversizeCount int                                `json:"skippedOversizeCount"`
	ResultLimit          int                                `json:"resultLimit"`
	ResultTruncated      bool                               `json:"resultTruncated"`
	LastError            string                             `json:"lastError"`
	Results              []FileLocatorResultSummary `json:"results"`
}

type FileLocatorTextRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type FileMetadataHit struct {
	Path           string `json:"path"`
	Size           int64  `json:"size"`
	SizeReady      bool   `json:"sizeReady"`
	ModifiedAt     string `json:"modifiedAt"`
	ModifiedAtUnix int64  `json:"modifiedAtUnix"`
}

type FileSearchConfig struct {
	Enabled             bool                           `json:"enabled"`
	PinyinSearchEnabled bool                           `json:"pinyinSearchEnabled,omitempty"`
	RecentItemsEnabled  *bool                          `json:"recentItemsEnabled,omitempty"`
	RecentItemsLimit    int                            `json:"recentItemsLimit,omitempty"`
	RecentItems         []FileSearchRecentItem `json:"recentItems,omitempty"`
	Volumes             []string                       `json:"volumes"`
	MaxResults          int                            `json:"maxResults"`
	IgnoredDirectories  []string                       `json:"ignoredDirectories,omitempty"`
	ResourceMode        string                         `json:"resourceMode,omitempty"`
	FileTypeFilters     []FileSearchTypeFilter `json:"fileTypeFilters"`
	LegacyMode          string                         `json:"mode,omitempty"`
	LegacyRoots         []string                       `json:"roots,omitempty"`
	LegacyRules         []string                       `json:"rules,omitempty"`
}

type FileSearchHit struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	Directory      string  `json:"directory"`
	IsDirectory    bool    `json:"isDirectory"`
	Extension      string  `json:"extension"`
	Snippet        string  `json:"snippet"`
	ModifiedAt     string  `json:"modifiedAt"`
	ModifiedAtUnix int64   `json:"modifiedAtUnix"`
	IconData       string  `json:"iconData"`
	Score          float64 `json:"score"`
	Root           string  `json:"root"`
	NodeIndex      int32   `json:"nodeIndex"`
	PathReady      bool    `json:"pathReady"`
}

type FileSearchPathHit struct {
	ID        string `json:"id"`
	Root      string `json:"root"`
	NodeIndex int32  `json:"nodeIndex"`
	Path      string `json:"path"`
}

type FileSearchPathRequest struct {
	ID        string `json:"id"`
	Root      string `json:"root"`
	NodeIndex int32  `json:"nodeIndex"`
}

type FileSearchRecentItem struct {
	Path        string `json:"path"`
	Name        string `json:"name,omitempty"`
	IsDirectory bool   `json:"isDirectory,omitempty"`
}

type FileSearchResult struct {
	Hits                 []FileSearchHit `json:"hits"`
	TotalMatchCount      int                     `json:"totalMatchCount"`
	SnapshotStale        bool                    `json:"snapshotStale"`
	Syncing              bool                    `json:"syncing"`
	RuntimeWarming       bool                    `json:"runtimeWarming"`
	PinyinRefreshPending bool                    `json:"pinyinRefreshPending"`
}

type FileSearchState struct {
	Config                 FileSearchConfig   `json:"config"`
	Volumes                []FileSearchVolume `json:"volumes"`
	Provider               string                     `json:"provider"`
	Ready                  bool                       `json:"ready"`
	Watching               bool                       `json:"watching"`
	RuntimeIndexCount      int                        `json:"runtimeIndexCount"`
	RuntimeIndexEntries    int                        `json:"runtimeIndexEntries"`
	WatcherCount           int                        `json:"watcherCount"`
	USNFollowerCount       int                        `json:"usnFollowerCount"`
	USNFollowerBufferBytes int64                      `json:"usnFollowerBufferBytes"`
	GoHeapAllocBytes       uint64                     `json:"goHeapAllocBytes"`
	GoHeapSysBytes         uint64                     `json:"goHeapSysBytes"`
	GoHeapReleasedBytes    uint64                     `json:"goHeapReleasedBytes"`
	GoHeapRetainedBytes    uint64                     `json:"goHeapRetainedBytes"`
	ProcessWorkingSetBytes uint64                     `json:"processWorkingSetBytes"`
	ProcessPrivateBytes    uint64                     `json:"processPrivateBytes"`
	IndexedCount           int                        `json:"indexedCount"`
	LastIndexedAt          string                     `json:"lastIndexedAt"`
	LastJournalReadAt      string                     `json:"lastJournalReadAt"`
	LastJournalApplyAt     string                     `json:"lastJournalApplyAt"`
	JournalNextUSN         int64                      `json:"journalNextUsn"`
	JournalLastUSN         int64                      `json:"journalLastUsn"`
	JournalLag             int64                      `json:"journalLag"`
	WatcherRestarting      bool                       `json:"watcherRestarting"`
	LastUsedAt             string                     `json:"lastUsedAt"`
	LastError              string                     `json:"lastError"`
	IndexPath              string                     `json:"indexPath"`
	RuntimeState           string                     `json:"runtimeState"`
	IdleUnloadAfter        int                        `json:"idleUnloadAfter"`
	Syncing                bool                       `json:"syncing"`
	Lagging                bool                       `json:"lagging"`
	PendingBatches         int                        `json:"pendingBatches"`
	PendingBytes           int64                      `json:"pendingBytes"`
	PendingAgeSeconds      int                        `json:"pendingAgeSeconds"`
	StaticMmap             bool                       `json:"staticMmap"`
	StaticMmapActive       bool                       `json:"staticMmapActive"`
	OverlayEntries         int                        `json:"overlayEntries"`
	OverlayRatio           float64                    `json:"overlayRatio"`
	MergePending           int                        `json:"mergePending"`
	MergeRunning           bool                       `json:"mergeRunning"`
	MergeRetries           int                        `json:"mergeRetries"`
	MergeError             string                     `json:"mergeError"`
	CheckpointPending      int                        `json:"checkpointPending"`
	CheckpointRunning      bool                       `json:"checkpointRunning"`
	CheckpointRetries      int                        `json:"checkpointRetries"`
	CheckpointError        string                     `json:"checkpointError"`
	WarmingVolumes         int                        `json:"warmingVolumes"`
	WarmedVolumes          int                        `json:"warmedVolumes"`
	WarmingTarget          string                     `json:"warmingTarget"`
	LastReleaseReason      string                     `json:"lastReleaseReason"`
	RestaticPending        int                        `json:"restaticPending"`
	RestaticRunning        bool                       `json:"restaticRunning"`
	RestaticRetries        int                        `json:"restaticRetries"`
	RestaticError          string                     `json:"restaticError"`
	PinyinIndexWarming     bool                       `json:"pinyinIndexWarming"`
	PinyinReadyVolumes     int                        `json:"pinyinReadyVolumes"`
	PinyinTargetVolumes    int                        `json:"pinyinTargetVolumes"`
	PinyinRecordCount      int                        `json:"pinyinRecordCount"`
	PinyinIndexFileBytes   int64                      `json:"pinyinIndexFileBytes"`
	PinyinMappedBytes      int64                      `json:"pinyinMappedBytes"`
	PinyinHeapBytes        int64                      `json:"pinyinHeapBytes"`
	PinyinIndexError       string                     `json:"pinyinIndexError"`
	RealtimeEnabled        bool                       `json:"realtimeEnabled"`
	RequiresAdmin          bool                       `json:"requiresAdmin"`
}

type FileSearchTypeFilter struct {
	ID      string   `json:"id"`
	Label   string   `json:"label,omitempty"`
	Enabled bool     `json:"enabled"`
	Rules   []string `json:"rules"`
}

type FileSearchVolume struct {
	Root            string `json:"root"`
	FileSystem      string `json:"fileSystem"`
	SupportsJournal bool   `json:"supportsJournal"`
	Included        bool   `json:"included"`
	IndexedCount    int    `json:"indexedCount"`
	IndexFileBytes  int64  `json:"indexFileBytes"`
}

type FileSearchWindowResult struct {
	Hits                 []FileSearchHit `json:"hits"`
	TotalMatchCount      int                     `json:"totalMatchCount"`
	Offset               int                     `json:"offset"`
	Limit                int                     `json:"limit"`
	NextOffset           int                     `json:"nextOffset"`
	HasMore              bool                    `json:"hasMore"`
	SnapshotStale        bool                    `json:"snapshotStale"`
	Syncing              bool                    `json:"syncing"`
	RuntimeWarming       bool                    `json:"runtimeWarming"`
	PinyinRefreshPending bool                    `json:"pinyinRefreshPending"`
}

type GPUPreferenceAdapterInfo struct {
	PowerSavingName     string `json:"powerSavingName"`
	HighPerformanceName string `json:"highPerformanceName"`
}

type GPUPreferenceEntry struct {
	Path            string                         `json:"path"`
	Name            string                         `json:"name"`
	Icon            string                         `json:"icon"`
	IconData        string                         `json:"iconData"`
	PreferenceMode  string                         `json:"preferenceMode"`
	PreferenceValue int                            `json:"preferenceValue"`
	Settings        []GPUPreferenceSetting `json:"settings"`
	RawValue        string                         `json:"rawValue"`
}

type GPUPreferenceSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GPUPreferenceState struct {
	Entries     []GPUPreferenceEntry     `json:"entries"`
	AdapterInfo GPUPreferenceAdapterInfo `json:"adapterInfo"`
}

type GestureAction struct {
	Kind           string   `json:"kind,omitempty"`
	Hotkey         string   `json:"hotkey,omitempty"`
	KeySequence    []string `json:"keySequence,omitempty"`
	Text           string   `json:"text,omitempty"`
	Path           string   `json:"path,omitempty"`
	Args           []string `json:"args,omitempty"`
	WorkingDir     string   `json:"workingDir,omitempty"`
	URL            string   `json:"url,omitempty"`
	WindowCommand  string   `json:"windowCommand,omitempty"`
	WindowMove     string   `json:"windowMove,omitempty"`
	ScreenID       string   `json:"screenId,omitempty"`
	ScreenIndex    int      `json:"screenIndex,omitempty"`
	ScreenName     string   `json:"screenName,omitempty"`
	LauncherAction string   `json:"launcherAction,omitempty"`
	AppID          string   `json:"appId,omitempty"`
	ProfileID      string   `json:"profileId,omitempty"`
}

type GestureAppMatch struct {
	Path                string `json:"path,omitempty"`
	ProcessName         string `json:"processName,omitempty"`
	WindowTitleContains string `json:"windowTitleContains,omitempty"`
	IconData            string `json:"iconData,omitempty"`
	IconRef             string `json:"iconRef,omitempty"`
	IconURL             string `json:"iconUrl,omitempty"`
}

type GestureAppProfile struct {
	ID                 string                    `json:"id"`
	Order              int                       `json:"order,omitempty"`
	Enabled            bool                      `json:"enabled"`
	InheritGlobal      bool                      `json:"inheritGlobal"`
	GlobalRulePriority string                    `json:"globalRulePriority,omitempty"`
	Blacklisted        bool                      `json:"blacklisted,omitempty"`
	Match              GestureAppMatch   `json:"match,omitempty"`
	Matches            []GestureAppMatch `json:"matches,omitempty"`
	Name               string                    `json:"name,omitempty"`
	DisplayName        string                    `json:"displayName,omitempty"`
	IconData           string                    `json:"iconData,omitempty"`
	IconRef            string                    `json:"iconRef,omitempty"`
	IconURL            string                    `json:"iconUrl,omitempty"`
	Rules              []GestureRule     `json:"rules,omitempty"`
}

type GesturePattern struct {
	Button     string   `json:"button,omitempty"`
	Directions []string `json:"directions,omitempty"`
	Modifier   string   `json:"modifier,omitempty"`
}

type GestureProfile struct {
	Enabled bool                  `json:"enabled,omitempty"`
	Rules   []GestureRule `json:"rules,omitempty"`
}

type GestureRule struct {
	ID      string                 `json:"id"`
	Order   int                    `json:"order,omitempty"`
	Enabled bool                   `json:"enabled"`
	Name    string                 `json:"name,omitempty"`
	Gesture GesturePattern `json:"gesture,omitempty"`
	Action  GestureAction  `json:"action,omitempty"`
}

type HotCornerConfig struct {
	Enabled               bool                    `json:"enabled,omitempty"`
	TriggerSizePx         int                     `json:"triggerSizePx,omitempty"`
	TriggerDelayMs        int                     `json:"triggerDelayMs,omitempty"`
	CooldownMs            int                     `json:"cooldownMs,omitempty"`
	AutoDisableFullscreen bool                    `json:"autoDisableFullscreen,omitempty"`
	Corners               []HotCornerRule `json:"corners,omitempty"`
}

type HotCornerRule struct {
	Corner  string `json:"corner,omitempty"`
	Enabled bool   `json:"enabled"`
	Hotkey  string `json:"hotkey,omitempty"`
}

type IndexNode struct {
	FRN        uint64
	NameOffset uint32
	ParentIdx  int32
	ModTime    uint32
	NameLen    uint16
	Flags      uint16
}

type InitializationImportRequest struct {
	SourceFile string `json:"sourceFile"`
	Mode       string `json:"mode"`
}

type InputMonitorEvent struct {
	ID         uint64 `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Device     string `json:"device"`
	Type       string `json:"type"`
	Action     string `json:"action"`
	Key        string `json:"key,omitempty"`
	Code       string `json:"code,omitempty"`
	VKCode     uint32 `json:"vkCode,omitempty"`
	ScanCode   uint32 `json:"scanCode,omitempty"`
	Button     string `json:"button,omitempty"`
	WheelDelta int    `json:"wheelDelta,omitempty"`
	WheelAxis  string `json:"wheelAxis,omitempty"`
	X          int    `json:"x,omitempty"`
	Y          int    `json:"y,omitempty"`
	Injected   bool   `json:"injected,omitempty"`
	Repeat     bool   `json:"repeat,omitempty"`
	AltKey     bool   `json:"altKey,omitempty"`
	CtrlKey    bool   `json:"ctrlKey,omitempty"`
	ShiftKey   bool   `json:"shiftKey,omitempty"`
	MetaKey    bool   `json:"metaKey,omitempty"`
}

type InputMonitorSnapshot struct {
	Supported       bool                        `json:"supported"`
	Platform        string                      `json:"platform"`
	Running         bool                        `json:"running"`
	Keyboard        bool                        `json:"keyboard"`
	Mouse           bool                        `json:"mouse"`
	ThreadHealthy   bool                        `json:"threadHealthy"`
	HookHealthy     bool                        `json:"hookHealthy"`
	StartedAt       string                      `json:"startedAt,omitempty"`
	LastError       string                      `json:"lastError,omitempty"`
	LastEventID     uint64                      `json:"lastEventId"`
	EarliestEventID uint64                      `json:"earliestEventId"`
	DroppedEvents   uint64                      `json:"droppedEvents"`
	Gap             bool                        `json:"gap"`
	BufferSize      int                         `json:"bufferSize"`
	BufferLimit     int                         `json:"bufferLimit"`
	Events          []InputMonitorEvent `json:"events"`
	MonitorStrategy string                      `json:"monitorStrategy,omitempty"`
}

type InputMonitorStartRequest struct {
	Keyboard bool `json:"keyboard,omitempty"`
	Mouse    bool `json:"mouse,omitempty"`
}

type LanguageManifest struct {
	Code       string                 `json:"code"`
	Name       string                 `json:"name"`
	NativeName string                 `json:"nativeName"`
	SourceFile string                 `json:"sourceFile"`
	Messages   map[string]interface{} `json:"messages,omitempty"`
}

type LauncherAdvertisementConfig struct {
	Enabled       bool   `json:"enabled"`
	ImageURL      string `json:"img"`
	ImageAssetURL string `json:"imageAssetUrl,omitempty"`
	URL           string `json:"url"`
	Seconds       int    `json:"sec"`
}

type LauncherBackgroundMetrics struct {
	Supported bool `json:"supported"`
	VirtualX  int  `json:"virtualX"`
	VirtualY  int  `json:"virtualY"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
}

type LauncherBackgroundSelectionResult struct {
	Preference BackgroundPreference `json:"preference"`
	FileName   string                       `json:"fileName,omitempty"`
}

type LauncherConfig struct {
	Revision         string                         `json:"revision,omitempty"`
	Initialized      bool                           `json:"initialized,omitempty"`
	Version          int                            `json:"version"`
	Storage          StorageConfig          `json:"storage,omitempty"`
	Preferences      Preferences            `json:"preferences"`
	Apps             []AppEntry             `json:"apps"`
	SpeedDial        []LinkEntry            `json:"speedDial"`
	Bookmarks        BookmarkSection        `json:"bookmarks"`
	FileSearch       FileSearchConfig       `json:"fileSearch"`
	FileLocator      FileLocatorConfig      `json:"fileLocator"`
	Files            []FileEntry            `json:"files"`
	MemoryRelease    MemoryReleaseConfig    `json:"memoryRelease"`
	OLEDBlackout     OLEDBlackoutConfig     `json:"oledBlackout"`
	WindowManagement WindowManagementConfig `json:"windowManagement"`
	MouseGestures    MouseGestureConfig     `json:"mouseGestures"`
	TwoFactor        TwoFactorConfig        `json:"twoFactor"`
}

type LauncherConfigOptions struct {
	DefaultTagCatalog []TagCatalogItem
}

type LauncherIconResource struct {
	IconData string `json:"iconData,omitempty"`
	IconRef  string `json:"iconRef,omitempty"`
	IconURL  string `json:"iconUrl,omitempty"`
}

type LauncherLatestVersionState struct {
	Version       string                              `json:"version"`
	SourceURL     string                              `json:"sourceUrl"`
	CheckedAt     string                              `json:"checkedAt"`
	ReleaseURL    string                              `json:"releaseUrl,omitempty"`
	Update        LauncherUpdatePackageConfig `json:"update"`
	Advertisement LauncherAdvertisementConfig `json:"advertisement"`
}

type LauncherState struct {
	Workspace                WorkspaceLayout    `json:"workspace"`
	Languages                []LanguageManifest `json:"languages"`
	Plugins                  []PluginManifest   `json:"plugins"`
	Config                   LauncherConfig     `json:"config"`
	StartupTrayMode          bool                       `json:"startupTrayMode"`
	ConfigOnly               bool                       `json:"configOnly,omitempty"`
	ContentRuntimeSyncNeeded bool                       `json:"contentRuntimeSyncNeeded,omitempty"`
	HotkeyRegistrationErrors map[string]string          `json:"hotkeyRegistrationErrors,omitempty"`
}

type LauncherUpdateInstallResult struct {
	Started bool   `json:"started"`
	Version string `json:"version"`
}

type LauncherUpdatePackageConfig struct {
	Available  bool   `json:"available"`
	PackageURL string `json:"packageUrl,omitempty"`
	SHA256     string `json:"sha256,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Root       string `json:"root,omitempty"`
}

type LauncherUpdateProgressState struct {
	Active          bool   `json:"active"`
	Stage           string `json:"stage"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
	Percent         int    `json:"percent"`
	Error           string `json:"error,omitempty"`
	UpdatedAt       string `json:"updatedAt"`
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

type MemoryReleaseConfig struct {
	Mode            string `json:"mode,omitempty"`
	TimerEnabled    bool   `json:"timerEnabled,omitempty"`
	IntervalMinutes int    `json:"intervalMinutes,omitempty"`
}

type MemoryReleaseState struct {
	Supported       bool   `json:"supported"`
	IsAdmin         bool   `json:"isAdmin"`
	Running         bool   `json:"running"`
	SchedulerActive bool   `json:"schedulerActive"`
	NextRunAt       string `json:"nextRunAt,omitempty"`
	LastRunAt       string `json:"lastRunAt,omitempty"`
	LastMode        string `json:"lastMode,omitempty"`
	LastError       string `json:"lastError,omitempty"`
}

type MouseGestureActionTarget struct {
	HWND        uintptr
	ProcessID   uint32
	Path        string
	ProcessName string
	Title       string
	Activate    bool
}

type MouseGestureConfig struct {
	Enabled    bool                        `json:"enabled,omitempty"`
	Settings   GestureSettings     `json:"settings,omitempty"`
	Global     GestureProfile      `json:"global,omitempty"`
	Apps       []GestureAppProfile `json:"apps,omitempty"`
	HotCorners HotCornerConfig     `json:"hotCorners,omitempty"`
}

type MouseGesturePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MouseGestureScreenBounds struct {
	ID     string `json:"id,omitempty"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MouseGestureState struct {
	Loading           bool                       `json:"loading"`
	Supported         bool                       `json:"supported"`
	Platform          string                     `json:"platform"`
	ModuleEnabled     bool                       `json:"moduleEnabled"`
	RuntimeEnabled    bool                       `json:"runtimeEnabled"`
	HotCornersEnabled bool                       `json:"hotCornersEnabled"`
	RuntimeActive     bool                       `json:"runtimeActive"`
	RuntimeWanted     bool                       `json:"runtimeWanted"`
	Config            MouseGestureConfig `json:"config"`
	LastAction        string                     `json:"lastAction,omitempty"`
	LastGesture       string                     `json:"lastGesture,omitempty"`
	LastError         string                     `json:"lastError,omitempty"`
}

type MouseGestureTarget struct {
	HWND        uintptr `json:"hwnd,omitempty"`
	ProcessID   uint32  `json:"processId,omitempty"`
	Path        string  `json:"path,omitempty"`
	ProcessName string  `json:"processName,omitempty"`
	Title       string  `json:"title,omitempty"`
}

type OLEDBlackoutConfig struct {
	ExitTriggers         OLEDBlackoutExitTriggers          `json:"exitTriggers,omitempty"`
	MediaPauseExclusions []OLEDBlackoutMediaPauseExclusion `json:"mediaPauseExclusions,omitempty"`
	Profiles             []OLEDBlackoutProfileConfig       `json:"profiles,omitempty"`
}

type OLEDBlackoutExitTriggers struct {
	Escape    bool `json:"escape,omitempty"`
	MouseMove bool `json:"mouseMove,omitempty"`
	AnyKey    bool `json:"anyKey,omitempty"`
}

type OLEDBlackoutMediaPauseExclusion struct {
	Enabled     bool   `json:"enabled"`
	ProcessName string `json:"processName,omitempty"`
	Path        string `json:"path,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	IconData    string `json:"iconData,omitempty"`
	IconRef     string `json:"iconRef,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

type OLEDBlackoutProfileConfig struct {
	ID            string   `json:"id"`
	ScreenIDs     []string `json:"screenIds,omitempty"`
	ScreenIndexes []int    `json:"screenIndexes,omitempty"`
	ScreenNames   []string `json:"screenNames,omitempty"`
	Hotkey        string   `json:"hotkey,omitempty"`
	IdleEnabled   bool     `json:"idleEnabled,omitempty"`
	IdleMinutes   int      `json:"idleMinutes,omitempty"`
	QuickIdle     bool     `json:"quickIdle,omitempty"`
	MediaPause    bool     `json:"mediaPause,omitempty"`
}

type OLEDBlackoutProfileState struct {
	ID                string   `json:"id"`
	ScreenIDs         []string `json:"screenIds"`
	ScreenIndexes     []int    `json:"screenIndexes,omitempty"`
	ScreenNames       []string `json:"screenNames,omitempty"`
	Hotkey            string   `json:"hotkey,omitempty"`
	IdleEnabled       bool     `json:"idleEnabled,omitempty"`
	IdleMinutes       int      `json:"idleMinutes,omitempty"`
	QuickIdle         bool     `json:"quickIdle,omitempty"`
	MediaPause        bool     `json:"mediaPause,omitempty"`
	Registered        bool     `json:"registered"`
	RegistrationError string   `json:"registrationError,omitempty"`
	MissingScreenIDs  []string `json:"missingScreenIds,omitempty"`
	Active            bool     `json:"active"`
	PartiallyActive   bool     `json:"partiallyActive,omitempty"`
}

type OLEDBlackoutScreen struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Index     int     `json:"index"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	IsPrimary bool    `json:"isPrimary"`
	Scale     float32 `json:"scale"`
}

type OLEDBlackoutState struct {
	Loading         bool                               `json:"loading"`
	Supported       bool                               `json:"supported"`
	Enabled         bool                               `json:"enabled"`
	Config          OLEDBlackoutConfig         `json:"config"`
	ActiveProfileID string                             `json:"activeProfileId,omitempty"`
	CursorHidden    bool                               `json:"cursorHidden"`
	LastError       string                             `json:"lastError,omitempty"`
	DebugLogPath    string                             `json:"debugLogPath,omitempty"`
	Screens         []OLEDBlackoutScreen       `json:"screens"`
	Profiles        []OLEDBlackoutProfileState `json:"profiles"`
}

type PluginEnvironment struct {
	PluginID         string                 `json:"pluginId"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	APIVersion       string                 `json:"apiVersion"`
	ThemeMode        string                 `json:"themeMode"`
	ResolvedTheme    string                 `json:"resolvedTheme"`
	Language         string                 `json:"language"`
	FallbackLanguage string                 `json:"fallbackLanguage"`
	UIScalePercent   int                    `json:"uiScalePercent"`
	Permissions      []string               `json:"permissions"`
	UI               PluginUIConfig `json:"ui"`
	Messages         map[string]interface{} `json:"messages,omitempty"`
}

type PluginInstallResult struct {
	State     LauncherState            `json:"state"`
	Catalog   PluginUpdateCatalogState `json:"catalog"`
	Installed []string                         `json:"installed,omitempty"`
	Updated   []string                         `json:"updated,omitempty"`
	Skipped   []string                         `json:"skipped,omitempty"`
}

type PluginLocalization struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Messages    map[string]interface{} `json:"messages,omitempty"`
}

type PluginManifest struct {
	ID              string                                `json:"id"`
	Name            string                                `json:"name"`
	Description     string                                `json:"description"`
	Version         string                                `json:"version"`
	MinAppVersion   string                                `json:"minAppVersion,omitempty"`
	Entry           string                                `json:"entry"`
	Category        string                                `json:"category"`
	APIVersion      string                                `json:"apiVersion,omitempty"`
	Icon            string                                `json:"icon,omitempty"`
	IconURL         string                                `json:"iconUrl,omitempty"`
	UI              PluginUIConfig                `json:"ui,omitempty"`
	Permissions     []string                              `json:"permissions,omitempty"`
	I18N            map[string]PluginLocalization `json:"i18n,omitempty"`
	SourceDir       string                                `json:"sourceDir"`
	Installed       bool                                  `json:"installed"`
	RemoteAvailable bool                                  `json:"remoteAvailable,omitempty"`
	RemoteVersion   string                                `json:"remoteVersion,omitempty"`
	PackageFile     string                                `json:"packageFile,omitempty"`
	PackageURL      string                                `json:"packageUrl,omitempty"`
	PackageSHA256   string                                `json:"packageSha256,omitempty"`
	PackageSize     int64                                 `json:"packageSize,omitempty"`
	UpdateAvailable bool                                  `json:"updateAvailable,omitempty"`
}

type PluginRemoteCatalog struct {
	SchemaVersion int                                `json:"schemaVersion"`
	GeneratedAt   string                             `json:"generatedAt,omitempty"`
	BaseURL       string                             `json:"baseUrl,omitempty"`
	Plugins       []PluginRemoteCatalogEntry `json:"plugins"`
	CDNPurgeURLs  []string                           `json:"cdnPurgeUrls,omitempty"`
}

type PluginRemoteCatalogEntry struct {
	ID            string                                `json:"id"`
	Name          string                                `json:"name,omitempty"`
	Description   string                                `json:"description,omitempty"`
	Version       string                                `json:"version,omitempty"`
	MinAppVersion string                                `json:"minAppVersion,omitempty"`
	Entry         string                                `json:"entry,omitempty"`
	Category      string                                `json:"category,omitempty"`
	APIVersion    string                                `json:"apiVersion,omitempty"`
	Icon          string                                `json:"icon,omitempty"`
	UI            PluginUIConfig                `json:"ui,omitempty"`
	Permissions   []string                              `json:"permissions,omitempty"`
	I18N          map[string]PluginLocalization `json:"i18n,omitempty"`
	PackageFile   string                                `json:"package,omitempty"`
	PackageURL    string                                `json:"packageUrl,omitempty"`
	PackageSHA256 string                                `json:"sha256,omitempty"`
	PackageSize   int64                                 `json:"size,omitempty"`
}

type PluginScreenInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Index     int     `json:"index"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	IsPrimary bool    `json:"isPrimary"`
	Scale     float32 `json:"scale"`
}

type PluginUIConfig struct {
	Theme       string `json:"theme,omitempty"`
	DefaultMode string `json:"defaultMode,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	MinWidth    int    `json:"minWidth,omitempty"`
	MinHeight   int    `json:"minHeight,omitempty"`
	Resizable   *bool  `json:"resizable,omitempty"`
	Frameless   bool   `json:"frameless,omitempty"`
	Fullscreen  bool   `json:"fullscreen,omitempty"`
	FluidWidth  bool   `json:"fluidWidth,omitempty"`
}

type PluginUpdateCatalogState struct {
	CatalogURL       string                   `json:"catalogUrl"`
	BaseURL          string                   `json:"baseUrl"`
	GeneratedAt      string                   `json:"generatedAt,omitempty"`
	CheckedAt        string                   `json:"checkedAt"`
	Plugins          []PluginManifest `json:"plugins"`
	UpdateCount      int                      `json:"updateCount"`
	InstallableCount int                      `json:"installableCount"`
}

type PluginWindowOpenRequest struct {
	Name         string `json:"name,omitempty"`
	Entry        string `json:"entry,omitempty"`
	Title        string `json:"title,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	MinWidth     int    `json:"minWidth,omitempty"`
	MinHeight    int    `json:"minHeight,omitempty"`
	Resizable    *bool  `json:"resizable,omitempty"`
	Frameless    *bool  `json:"frameless,omitempty"`
	Fullscreen   *bool  `json:"fullscreen,omitempty"`
	AlwaysOnTop  bool   `json:"alwaysOnTop,omitempty"`
	HideOnEscape *bool  `json:"hideOnEscape,omitempty"`
	ScreenID     string `json:"screenId,omitempty"`
	Query        string `json:"query,omitempty"`
}

type PluginWindowState struct {
	Name        string `json:"name"`
	PluginID    string `json:"pluginId"`
	Entry       string `json:"entry"`
	Title       string `json:"title"`
	Mode        string `json:"mode"`
	Visible     bool   `json:"visible"`
	Focused     bool   `json:"focused"`
	Frameless   bool   `json:"frameless"`
	Fullscreen  bool   `json:"fullscreen"`
	AlwaysOnTop bool   `json:"alwaysOnTop"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	ScreenID    string `json:"screenId,omitempty"`
	URL         string `json:"url"`
}

type Preferences struct {
	ThemeMode                              string                       `json:"themeMode"`
	UIScalePercent                         int                          `json:"uiScalePercent,omitempty"`
	HTTPProxy                              string                       `json:"httpProxy,omitempty"`
	AdDisplayEnabled                       *bool                        `json:"adDisplayEnabled,omitempty"`
	AppLaunchPrivilegeDefault              string                       `json:"appLaunchPrivilegeDefault,omitempty"`
	AppLaunchHideToTrayEnabled             *bool                        `json:"appLaunchHideToTrayEnabled,omitempty"`
	StartupLaunchEnabled                   *bool                        `json:"startupLaunchEnabled,omitempty"`
	StartupLaunchDelaySeconds              int                          `json:"startupLaunchDelaySeconds,omitempty"`
	DynamicStartMenuEnabled                *bool                        `json:"dynamicStartMenuEnabled,omitempty"`
	DynamicDesktopEnabled                  *bool                        `json:"dynamicDesktopEnabled,omitempty"`
	WebViewMemorySaverV8LimitEnabled       *bool                        `json:"webviewMemorySaverV8LimitEnabled,omitempty"`
	WebViewMemorySaverEdgeExtrasDisabled   *bool                        `json:"webviewMemorySaverEdgeExtrasDisabled,omitempty"`
	WebViewMemorySaverRendererLimitEnabled *bool                        `json:"webviewMemorySaverRendererLimitEnabled,omitempty"`
	WebViewMemoryReleaseStrategy           string                       `json:"webviewMemoryReleaseStrategy,omitempty"`
	WebViewDestroyOnTrayHideEnabled        *bool                        `json:"webviewDestroyOnTrayHideEnabled,omitempty"`
	ConfigBackupLimit                      int                          `json:"configBackupLimit,omitempty"`
	Language                               string                       `json:"language"`
	Background                             BackgroundPreference `json:"background,omitempty"`
	Hotkey                                 string                       `json:"hotkey"`
	HotkeyEnabled                          *bool                        `json:"hotkeyEnabled,omitempty"`
	SummonOnlyHotkey                       string                       `json:"summonOnlyHotkey,omitempty"`
	SummonOnlyHotkeyEnabled                *bool                        `json:"summonOnlyHotkeyEnabled,omitempty"`
	ScreenshotHotkey                       string                       `json:"screenshotHotkey"`
	ScreenshotHotkeyEnabled                *bool                        `json:"screenshotHotkeyEnabled,omitempty"`
	ScreenshotQRCodeHotkey                 string                       `json:"screenshotQRCodeHotkey"`
	ScreenshotQRCodeHotkeyEnabled          *bool                        `json:"screenshotQRCodeHotkeyEnabled,omitempty"`
	ScreenshotAllScreensHotkey             string                       `json:"screenshotAllScreensHotkey"`
	ScreenshotAllScreensHotkeyEnabled      *bool                        `json:"screenshotAllScreensHotkeyEnabled,omitempty"`
	ScreenshotScrollingHotkey              string                       `json:"screenshotScrollingHotkey,omitempty"`
	ScreenshotScrollingHotkeyEnabled       *bool                        `json:"screenshotScrollingHotkeyEnabled,omitempty"`
	ScreenshotActiveWindowHotkey           string                       `json:"screenshotActiveWindowHotkey"`
	ScreenshotActiveWindowHotkeyEnabled    *bool                        `json:"screenshotActiveWindowHotkeyEnabled,omitempty"`
	ScreenshotHotkeysInitialized           bool                         `json:"screenshotHotkeysInitialized,omitempty"`
	ScreenshotCaptureControls              *bool                        `json:"screenshotCaptureControlsEnabled,omitempty"`
	ScreenshotCaptureCursor                *bool                        `json:"screenshotCaptureCursorEnabled,omitempty"`
	ScreenshotSelectionConfirm             *bool                        `json:"screenshotSelectionConfirmEnabled,omitempty"`
	ScreenshotPreviewEnabled               *bool                        `json:"screenshotPreviewEnabled,omitempty"`
	ScreenshotSoundEnabled                 *bool                        `json:"screenshotSoundEnabled,omitempty"`
	ScreenshotSaveFormat                   string                       `json:"screenshotSaveFormat,omitempty"`
	ScreenshotHistoryDisplayLimit          int                          `json:"screenshotHistoryDisplayLimit,omitempty"`
	ScreenshotAnnotationLineWidth          int                          `json:"screenshotAnnotationLineWidth,omitempty"`
	ScreenshotCornerRadius                 int                          `json:"screenshotCornerRadius"`
	SearchCategoryShortcuts                []string                     `json:"searchCategoryShortcuts"`
	DefaultSearchCategory                  string                       `json:"defaultSearchCategory"`
	SearchScope                            string                       `json:"searchScope,omitempty"`
	SearchScopeCycle                       []string                     `json:"searchScopeCycle,omitempty"`
	SearchScopeCycleDefault                string                       `json:"searchScopeCycleDefault,omitempty"`
	MatchCase                              bool                         `json:"matchCase"`
	PinyinSearchEnabled                    *bool                        `json:"pinyinSearchEnabled,omitempty"`
	SearchHistory                          []string                     `json:"searchHistory,omitempty"`
	SearchCategoryOrder                    []string                     `json:"searchCategoryOrder"`
	NavigationVisibleModules               []string                     `json:"navigationVisibleModules"`
	FeatureModules                         map[string]bool              `json:"featureModules,omitempty"`
	BookmarkFavoritePath                   BookmarkFavoritePath `json:"bookmarkFavoritePath"`
	LinkBrowsers                           []LinkBrowser        `json:"linkBrowsers,omitempty"`
	DefaultLinkBrowserID                   string                       `json:"defaultLinkBrowserId,omitempty"`
	GlobalTags                             []string                     `json:"globalTags"`
	TagCatalog                             []TagCatalogItem     `json:"tagCatalog,omitempty"`
	SpeedDialTileSize                      int                          `json:"speedDialTileSize,omitempty"`
	SpeedDialSortMode                      string                       `json:"speedDialSortMode,omitempty"`
	SpeedDialCategoryOrder                 []string                     `json:"speedDialCategoryOrder,omitempty"`
	SpeedDialExpandedGroups                []string                     `json:"speedDialExpandedGroups,omitempty"`
	AppSortMode                            string                       `json:"appSortMode,omitempty"`
	AppRecentLimit                         int                          `json:"appRecentLimit,omitempty"`
	AppCategoryOrder                       []string                     `json:"appCategoryOrder,omitempty"`
	AppExpandedGroups                      []string                     `json:"appExpandedGroups,omitempty"`
	DynamicStartMenuLaunchHistory          []StartMenuLaunchLog `json:"dynamicStartMenuLaunchHistory,omitempty"`
	DragLaunchAppIDs                       []string                     `json:"dragLaunchAppIds,omitempty"`
	ConsoleItems                           []ConsoleItem        `json:"consoleItems,omitempty"`
	LegacyDragLaunchRules                  []DragLaunchRule     `json:"dragLaunchRules,omitempty"`
}

type QRCodeDecodeResult struct {
	ImageData                    string                       `json:"imageData"`
	ImageWidth                   int                          `json:"imageWidth"`
	ImageHeight                  int                          `json:"imageHeight"`
	Entries                      []QRCodeDecodedEntry `json:"entries"`
	Cancelled                    bool                         `json:"cancelled"`
	SelectedEntryKey             string                       `json:"selectedEntryKey,omitempty"`
	TriggeredFromLauncherVisible bool                         `json:"triggeredFromLauncherVisible,omitempty"`
	TriggeredFromLauncherFocused bool                         `json:"triggeredFromLauncherFocused,omitempty"`
}

type QRCodeDecodedEntry struct {
	Text         string  `json:"text"`
	Format       string  `json:"format"`
	Selectable   bool    `json:"selectable"`
	MarkerX      float64 `json:"markerX"`
	MarkerY      float64 `json:"markerY"`
	BoundsLeft   float64 `json:"boundsLeft"`
	BoundsTop    float64 `json:"boundsTop"`
	BoundsRight  float64 `json:"boundsRight"`
	BoundsBottom float64 `json:"boundsBottom"`
}

type RemoteIconSearchResult struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Collection     string `json:"collection"`
	CollectionName string `json:"collectionName"`
	IconData       string `json:"iconData"`
}

type ResolvedBookmarkItem struct {
	ID         string `json:"id"`
	SourceID   string `json:"sourceId"`
	SourceName string `json:"sourceName"`
	SourcePath string `json:"sourcePath"`
	Browser    string `json:"browser"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Folder     string `json:"folder"`
}

type ResolvedBookmarkSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Browser     string `json:"browser"`
	BrowserName string `json:"browserName"`
	Profile     string `json:"profile"`
	Path        string `json:"path"`
	Enabled     bool   `json:"enabled"`
	Exists      bool   `json:"exists"`
	ItemCount   int    `json:"itemCount"`
	Error       string `json:"error"`
}

type ScreenshotCaptureResult struct {
	ImageData        string `json:"imageData"`
	ImageURL         string `json:"imageUrl,omitempty"`
	ThumbnailURL     string `json:"thumbnailUrl,omitempty"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	Mode             string `json:"mode"`
	Path             string `json:"path"`
	Cancelled        bool   `json:"cancelled"`
	Truncated        bool   `json:"truncated,omitempty"`
	TruncationReason string `json:"truncationReason,omitempty"`
}

type ScreenshotHistoryEntry struct {
	ID         string `json:"id"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Mode       string `json:"mode"`
	Path       string `json:"path"`
	CapturedAt int64  `json:"capturedAt"`
}

type ScreenshotImageRef struct {
	ImageURL  string `json:"imageUrl,omitempty"`
	Path      string `json:"path,omitempty"`
	ImageData string `json:"imageData,omitempty"`
}

type ScreenshotPinState struct {
	WindowName   string `json:"windowName"`
	Visible      bool   `json:"visible"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ImageData    string `json:"imageData,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Scale        string `json:"scale"`
	Opacity      string `json:"opacity"`
	ClickThrough bool   `json:"clickThrough"`
	AutoShow     bool   `json:"autoShow"`
	Order        int64  `json:"order"`
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

type StartupState struct {
	Workspace    WorkspaceLayout    `json:"workspace"`
	Languages    []LanguageManifest `json:"languages"`
	Plugins      []PluginManifest   `json:"plugins"`
	ConfigExists bool                       `json:"configExists"`
	Initialized  bool                       `json:"initialized"`
}

type StorageConfig struct {
	DataRoot      string `json:"dataRoot,omitempty"`
	IconDir       string `json:"iconDir,omitempty"`
	IndexDir      string `json:"indexDir,omitempty"`
	ScreenshotDir string `json:"screenshotDir,omitempty"`
	WebView2Dir   string `json:"webview2Dir,omitempty"`
}

type TagCatalogItem struct {
	Name     string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	IconData string `json:"iconData,omitempty"`
	IconRef  string `json:"iconRef,omitempty"`
	IconURL  string `json:"iconUrl,omitempty"`
}

type TwoFactorCommandResult struct {
	Config TwoFactorConfig `json:"config"`
	State  TwoFactorState  `json:"state"`
}

type TwoFactorConfig struct {
	Password TwoFactorPasswordConfig `json:"password"`
	Entries  []TwoFactorEntryConfig  `json:"entries,omitempty"`
}

type TwoFactorDraftPreviewResult struct {
	Entry     TwoFactorEntryState `json:"entry"`
	TimeState TwoFactorTimeState  `json:"timeState"`
}

type TwoFactorEntryConfig struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Name             string `json:"name"`
	Issuer           string `json:"issuer,omitempty"`
	AccountName      string `json:"accountName,omitempty"`
	Icon             string `json:"icon,omitempty"`
	IconData         string `json:"iconData,omitempty"`
	IconRef          string `json:"iconRef,omitempty"`
	IconURL          string `json:"iconUrl,omitempty"`
	Algorithm        string `json:"algorithm,omitempty"`
	Digits           int    `json:"digits,omitempty"`
	Period           int    `json:"period,omitempty"`
	SecretNonce      string `json:"secretNonce"`
	SecretCiphertext string `json:"secretCiphertext"`
}

type TwoFactorEntryDraft struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Issuer      string `json:"issuer"`
	AccountName string `json:"accountName"`
	Icon        string `json:"icon,omitempty"`
	IconData    string `json:"iconData,omitempty"`
	IconRef     string `json:"iconRef,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Secret      string `json:"secret"`
	Algorithm   string `json:"algorithm,omitempty"`
	Digits      int    `json:"digits,omitempty"`
	Period      int    `json:"period,omitempty"`
}

type TwoFactorEntryState struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Name             string `json:"name"`
	Issuer           string `json:"issuer,omitempty"`
	AccountName      string `json:"accountName,omitempty"`
	Icon             string `json:"icon,omitempty"`
	IconData         string `json:"iconData,omitempty"`
	IconRef          string `json:"iconRef,omitempty"`
	IconURL          string `json:"iconUrl,omitempty"`
	Algorithm        string `json:"algorithm,omitempty"`
	Digits           int    `json:"digits"`
	Period           int    `json:"period"`
	Code             string `json:"code,omitempty"`
	SecondsRemaining int    `json:"secondsRemaining"`
}

type TwoFactorExportResult struct {
	Content string `json:"content"`
	Count   int    `json:"count"`
}

type TwoFactorImportResult struct {
	Config         TwoFactorConfig `json:"config"`
	State          TwoFactorState  `json:"state"`
	ImportedCount  int                     `json:"importedCount"`
	DuplicateCount int                     `json:"duplicateCount"`
	SkippedCount   int                     `json:"skippedCount"`
}

type TwoFactorPasswordConfig struct {
	KDF      string `json:"kdf,omitempty"`
	Salt     string `json:"salt,omitempty"`
	Verifier string `json:"verifier,omitempty"`
	Blank    bool   `json:"blank,omitempty"`
}

type TwoFactorProvisioningParseResult struct {
	Draft TwoFactorEntryDraft `json:"draft"`
}

type TwoFactorRevealResult struct {
	Entries   []TwoFactorEntryState `json:"entries"`
	TimeState TwoFactorTimeState    `json:"timeState"`
}

type TwoFactorState struct {
	PasswordConfigured bool                          `json:"passwordConfigured"`
	PasswordBlank      bool                          `json:"passwordBlank"`
	Unlocked           bool                          `json:"unlocked"`
	EntryCount         int                           `json:"entryCount"`
	Entries            []TwoFactorEntryState `json:"entries"`
	TimeState          TwoFactorTimeState    `json:"timeState"`
}

type TwoFactorTimeState struct {
	Source             string `json:"source"`
	SourceName         string `json:"sourceName,omitempty"`
	OffsetMilliseconds int64  `json:"offsetMilliseconds"`
	SyncedAt           string `json:"syncedAt,omitempty"`
	SampleCount        int    `json:"sampleCount,omitempty"`
	ErrorMessage       string `json:"errorMessage,omitempty"`
}

type WebView2ProcessInfo struct {
	ProcessID               int     `json:"processId"`
	Kind                    string  `json:"kind"`
	WorkingSetBytes         uint64  `json:"workingSetBytes"`
	PrivateBytes            uint64  `json:"privateBytes"`
	WorkingSetMB            float64 `json:"workingSetMB"`
	PrivateMB               float64 `json:"privateMB"`
	HasJSFlags              bool    `json:"hasJsFlags"`
	HasDisableFeatures      bool    `json:"hasDisableFeatures"`
	HasRendererProcessLimit bool    `json:"hasRendererProcessLimit"`
	CommandLine             string  `json:"commandLine,omitempty"`
}

type WebView2ProcessSnapshot struct {
	UserDataDir          string                        `json:"userDataDir,omitempty"`
	Processes            []WebView2ProcessInfo `json:"processes"`
	TotalWorkingSetBytes uint64                        `json:"totalWorkingSetBytes"`
	TotalPrivateBytes    uint64                        `json:"totalPrivateBytes"`
	TotalWorkingSetMB    float64                       `json:"totalWorkingSetMB"`
	TotalPrivateMB       float64                       `json:"totalPrivateMB"`
}

type WindowFullscreenSnapshot struct {
	HWND      uintptr                      `json:"hwnd,omitempty"`
	ProcessID uint32                       `json:"processId,omitempty"`
	Rect      WindowManagementRect `json:"rect,omitempty"`
	Style     uintptr                      `json:"style,omitempty"`
	ExStyle   uintptr                      `json:"exStyle,omitempty"`
	TopMost   bool                         `json:"topMost,omitempty"`
}

type WindowManagementInfo struct {
	HWND               uintptr                      `json:"hwnd,omitempty"`
	WindowText         string                       `json:"windowText,omitempty"`
	WindowClass        string                       `json:"windowClass,omitempty"`
	WindowStyle        string                       `json:"windowStyle,omitempty"`
	WindowExStyle      string                       `json:"windowExStyle,omitempty"`
	WindowStyleValue   string                       `json:"windowStyleValue,omitempty"`
	WindowExStyleValue string                       `json:"windowExStyleValue,omitempty"`
	WindowID           uintptr                      `json:"windowId,omitempty"`
	ParentHWND         uintptr                      `json:"parentHwnd,omitempty"`
	ParentText         string                       `json:"parentText,omitempty"`
	ParentClass        string                       `json:"parentClass,omitempty"`
	ThreadID           uint32                       `json:"threadId,omitempty"`
	ProcessID          uint32                       `json:"processId,omitempty"`
	ProcessName        string                       `json:"processName,omitempty"`
	ProcessPath        string                       `json:"processPath,omitempty"`
	WindowRect         WindowManagementRect `json:"windowRect,omitempty"`
	ClientRect         WindowManagementRect `json:"clientRect,omitempty"`
}

type WindowManagementRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WindowManagementSize struct {
	Width      int  `json:"width,omitempty"`
	Height     int  `json:"height,omitempty"`
	ClientArea bool `json:"clientArea,omitempty"`
}

type WindowManagementState struct {
	Supported            bool                           `json:"supported"`
	Platform             string                         `json:"platform"`
	ModuleEnabled        bool                           `json:"moduleEnabled"`
	Config               WindowManagementConfig `json:"config"`
	Target               WindowManagementTarget `json:"target,omitempty"`
	TargetInfo           WindowManagementInfo   `json:"targetInfo,omitempty"`
	TargetValid          bool                           `json:"targetValid"`
	DisplayCount         int                            `json:"displayCount"`
	CursorWrapActive     bool                           `json:"cursorWrapActive"`
	TargetTopMost        bool                           `json:"targetTopMost"`
	TargetBorderless     bool                           `json:"targetBorderless"`
	TargetFullscreen     bool                           `json:"targetFullscreen"`
	TargetOpacityPercent int                            `json:"targetOpacityPercent,omitempty"`
	LastError            string                         `json:"lastError,omitempty"`
}

type WindowManagementTarget struct {
	HWND        uintptr `json:"hwnd,omitempty"`
	ProcessID   uint32  `json:"processId,omitempty"`
	ProcessName string  `json:"processName,omitempty"`
	Path        string  `json:"path,omitempty"`
	Title       string  `json:"title,omitempty"`
	DisplayName string  `json:"displayName,omitempty"`
	IconData    string  `json:"iconData,omitempty"`
	IconRef     string  `json:"iconRef,omitempty"`
	IconURL     string  `json:"iconUrl,omitempty"`
}

type WindowProcessPickResult struct {
	Path        string  `json:"path"`
	ProcessID   uint32  `json:"processId"`
	ProcessName string  `json:"processName"`
	DisplayName string  `json:"displayName"`
	Title       string  `json:"title,omitempty"`
	HWND        uintptr `json:"hwnd,omitempty"`
	IconData    string  `json:"iconData,omitempty"`
	IconURL     string  `json:"iconUrl,omitempty"`
}

type WorkspaceLayout struct {
	Root           string `json:"root"`
	ConfigFile     string `json:"configFile"`
	LanguageDir    string `json:"languageDir"`
	AppLanguageDir string `json:"appLanguageDir,omitempty"`
	PluginDir      string `json:"pluginDir"`
	IconDir        string `json:"iconDir"`
	IndexDir       string `json:"indexDir"`
	ScreenshotDir  string `json:"screenshotDir"`
	WebView2Dir    string `json:"webview2Dir"`
	BackgroundDir  string `json:"backgroundDir"`
}

type appIconInfo struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type audioIMMDeviceEnumeratorVtbl struct {
	QueryInterface                         uintptr
	AddRef                                 uintptr
	Release                                uintptr
	EnumAudioEndpoints                     uintptr
	GetDefaultAudioEndpoint                uintptr
	GetDevice                              uintptr
	RegisterEndpointNotificationCallback   uintptr
	UnregisterEndpointNotificationCallback uintptr
}

type audioIMMDeviceVtbl struct {
	QueryInterface    uintptr
	AddRef            uintptr
	Release           uintptr
	Activate          uintptr
	OpenPropertyStore uintptr
	GetID             uintptr
	GetState          uintptr
}

type audioPropVariant struct {
	VT        uint16
	Reserved1 uint16
	Reserved2 uint16
	Reserved3 uint16
	Data      [16]uint8
}

type audioSessionSnapshot struct {
	GroupID                string
	AppName                string
	ProcessID              uint32
	ProcessPath            string
	IconData               string
	VolumePercent          int
	Muted                  bool
	Active                 bool
	SystemSounds           bool
	OutputDeviceID         string
	InputDeviceID          string
	DeviceRoutingSupported bool
}

type chromiumBookmarkFile struct {
	Roots map[string]chromiumBookmarkNode `json:"roots"`
}

type chromiumBookmarkNode struct {
	Children []chromiumBookmarkNode `json:"children"`
	Name     string                         `json:"name"`
	Type     string                         `json:"type"`
	URL      string                         `json:"url"`
}

type chromiumBookmarkRoot struct {
	BrowserName string
	Path        string
}

type desktopReminderAudio struct {
	Mode        string
	Path        string
	RepeatCount int
}

type desktopWidgetChangeEvent struct {
	IDs      []string `json:"ids"`
	Revision int64    `json:"revision"`
	Kind     string   `json:"kind"`
}

type directLauncherNetworkAccess struct{}

type displayconfigDeviceInfoHeader struct {
	Type      uint32
	Size      uint32
	AdapterID windows.LUID
	ID        uint32
}

type displayconfigModeInfo struct {
	InfoType  uint32
	ID        uint32
	AdapterID windows.LUID
	Info      [6]uint64
}

type displayconfigPathInfo struct {
	SourceInfo displayconfigPathSourceInfo
	TargetInfo displayconfigPathTargetInfo
	Flags      uint32
}

type displayconfigPathSourceInfo struct {
	AdapterID   windows.LUID
	ID          uint32
	ModeInfoIdx uint32
	StatusFlags uint32
}

type displayconfigPathTargetInfo struct {
	AdapterID        windows.LUID
	ID               uint32
	ModeInfoIdx      uint32
	OutputTechnology uint32
	Rotation         uint32
	Scaling          uint32
	RefreshRate      displayconfigRational
	ScanLineOrdering uint32
	TargetAvailable  int32
	StatusFlags      uint32
}

type displayconfigRational struct {
	Numerator   uint32
	Denominator uint32
}

type displayconfigSDRWhiteLevel struct {
	Header        displayconfigDeviceInfoHeader
	SDRWhiteLevel uint32
}

type displayconfigSourceDeviceName struct {
	Header            displayconfigDeviceInfoHeader
	ViewGDIDeviceName [32]uint16
}

type dxgiAdapter struct {
	Vtbl *dxgiAdapterVtbl
}

type dxgiAdapterVtbl struct {
	QueryInterface          uintptr
	AddRef                  uintptr
	Release                 uintptr
	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	GetPrivateData          uintptr
	GetParent               uintptr
	EnumOutputs             uintptr
	GetDesc                 uintptr
	CheckInterfaceSupport   uintptr
}

type dxgiIUnknown struct {
	Vtbl *dxgiIUnknownVtbl
}

type dxgiIUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type dxgiOutput6 struct {
	Vtbl *dxgiOutput6Vtbl
}

type dxgiOutput6Vtbl struct {
	QueryInterface                uintptr
	AddRef                        uintptr
	Release                       uintptr
	SetPrivateData                uintptr
	SetPrivateDataInterface       uintptr
	GetPrivateData                uintptr
	GetParent                     uintptr
	GetDesc                       uintptr
	GetDisplayModeList            uintptr
	FindClosestMatchingMode       uintptr
	WaitForVBlank                 uintptr
	TakeOwnership                 uintptr
	ReleaseOwnership              uintptr
	GetGammaControlCapabilities   uintptr
	SetGammaControl               uintptr
	GetGammaControl               uintptr
	SetDisplaySurface             uintptr
	GetDisplaySurfaceData         uintptr
	GetFrameStatistics            uintptr
	GetDisplayModeList1           uintptr
	FindClosestMatchingMode1      uintptr
	GetDisplaySurfaceData1        uintptr
	DuplicateOutput               uintptr
	SupportsOverlays              uintptr
	CheckOverlaySupport           uintptr
	CheckOverlayColorSpaceSupport uintptr
	DuplicateOutput1              uintptr
	GetDesc1                      uintptr
	CheckHardwareComposition      uintptr
}

type fileMetadataResolveTask struct {
	Path string
	Name string
	Key  string
}

type fileSearchProcessMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
}

type fileSearchSortSpec struct {
	Key       string
	Direction string
}

type fileSearchUSNFollower struct {
	Root      string
	Handle    uintptr
	JournalID uint64
	Buffer    []uint8
	Stop      chan struct{}
	Done      chan struct{}
}

type fileSearchUSNFollowerMeta struct {
	Version             uint32 `json:"version"`
	Root                string `json:"root"`
	JournalID           uint64 `json:"journalId"`
	LastAppliedUSN      int64  `json:"lastAppliedUsn"`
	LastReadUSN         int64  `json:"lastReadUsn"`
	LastReadAtUnixNano  int64  `json:"lastReadAtUnixNano"`
	LastPersistedAtNano int64  `json:"lastPersistedAtUnixNano"`
	Recovering          bool   `json:"recovering"`
	LastError           string `json:"lastError,omitempty"`
}

type fileSearchVolumeMeta struct {
	IndexPath           string
	IndexedCount        int
	LastIndexedAt       time.Time
	LastJournalReadAt   time.Time
	LastJournalApplyAt  time.Time
	Ready               bool
	Watching            bool
	RealtimeEnabled     bool
	WatcherRestarting   bool
	JournalNextUsn      int64
	JournalID           uint64
	LastUsn             int64
	RootReferenceNumber uint64
}

type firefoxBookmarkNode struct {
	ID        int64
	ParentID  int64
	Type      int64
	Title     string
	PageTitle string
	URL       string
	GUID      string
}

type firefoxProfileSource struct {
	BrowserName string
	Profile     string
	Path        string
}

type gpuPreferenceScreenPoint struct {
	X int32
	Y int32
}

type indexSearchTerm struct {
	Original       string
	Parts          [][]uint8
	IsPath         bool
	IsDrive        bool
	AllowHierarchy bool
	AllowNameMatch bool
	PinyinFuzzy    *fileSearchPinyinFuzzyMatcher
}

type launcherAssetRef struct {
	ID          string
	URL         string
	Version     int64
	ContentType string
}

type launcherConfigConflictError struct {
	Expected string
	Actual   string
}

type launcherConfigIconAssetCacheEntry struct {
	URL      string
	Sequence uint64
}

type launcherConfigIconDataError struct {
	Path   string
	Reason string
}

type launcherConfigIconExtractionResult struct {
	Extracted int
	Added     int
	Reused    int
	Removed   int
}

type launcherConfigIconHydrationError struct {
	Path string
	Ref  string
	Err  error
}

type launcherConfigIconLibrary struct {
	Version int                                         `json:"version"`
	Icons   map[string]launcherConfigIconRecord `json:"icons"`
}

type launcherConfigIconRecord struct {
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
}

type launcherConfigIconSlot struct {
	Path string
	Data *string
	Ref  *string
	URL  *string
}

type launcherHotkeyBindings struct {
	SummonSearch                  string
	SummonSearchEnabled           bool
	SummonOnly                    string
	SummonOnlyEnabled             bool
	ScreenshotFeatureEnabled      bool
	Screenshot                    string
	ScreenshotEnabled             bool
	ScreenshotQRCode              string
	ScreenshotQRCodeEnabled       bool
	ScreenshotAllScreens          string
	ScreenshotAllScreensEnabled   bool
	ScreenshotScrolling           string
	ScreenshotScrollingEnabled    bool
	ScreenshotActiveWindow        string
	ScreenshotActiveWindowEnabled bool
}

type launcherHotkeyRegistrationError struct {
	Action  string
	Binding string
	Err     error
}

type launcherStartupTaskActionsXML struct {
	Context string                             `xml:"Context,attr"`
	Exec    launcherStartupTaskExecXML `xml:"Exec"`
}

type launcherStartupTaskEnabledXML struct {
	Enabled bool   `xml:"Enabled"`
	Delay   string `xml:"Delay,omitempty"`
}

type launcherStartupTaskExecXML struct {
	Command          string `xml:"Command"`
	Arguments        string `xml:"Arguments,omitempty"`
	WorkingDirectory string `xml:"WorkingDirectory"`
}

type launcherStartupTaskIdleSettingsXML struct {
	Duration      string `xml:"Duration"`
	WaitTimeout   string `xml:"WaitTimeout"`
	StopOnIdleEnd bool   `xml:"StopOnIdleEnd"`
	RestartOnIdle bool   `xml:"RestartOnIdle"`
}

type launcherStartupTaskPrincipalXML struct {
	ID        string `xml:"id,attr"`
	UserID    string `xml:"UserId"`
	LogonType string `xml:"LogonType"`
	RunLevel  string `xml:"RunLevel"`
}

type launcherStartupTaskPrincipalsXML struct {
	Principal launcherStartupTaskPrincipalXML `xml:"Principal"`
}

type launcherStartupTaskSettingsXML struct {
	MultipleInstancesPolicy    string                                     `xml:"MultipleInstancesPolicy"`
	DisallowStartIfOnBatteries bool                                       `xml:"DisallowStartIfOnBatteries"`
	StopIfGoingOnBatteries     bool                                       `xml:"StopIfGoingOnBatteries"`
	AllowHardTerminate         bool                                       `xml:"AllowHardTerminate"`
	StartWhenAvailable         bool                                       `xml:"StartWhenAvailable"`
	RunOnlyIfNetworkAvailable  bool                                       `xml:"RunOnlyIfNetworkAvailable"`
	IdleSettings               launcherStartupTaskIdleSettingsXML `xml:"IdleSettings"`
	AllowStartOnDemand         bool                                       `xml:"AllowStartOnDemand"`
	Enabled                    bool                                       `xml:"Enabled"`
	Hidden                     bool                                       `xml:"Hidden"`
	RunOnlyIfIdle              bool                                       `xml:"RunOnlyIfIdle"`
	WakeToRun                  bool                                       `xml:"WakeToRun"`
	ExecutionTimeLimit         string                                     `xml:"ExecutionTimeLimit"`
	Priority                   *int                                       `xml:"Priority,omitempty"`
}

type launcherStartupTaskTriggersXML struct {
	LogonTrigger launcherStartupTaskEnabledXML `xml:"LogonTrigger"`
}

type launcherStartupTaskXML struct {
	XMLName    xml.Name                              `xml:"Task"`
	Version    string                                   `xml:"version,attr"`
	Xmlns      string                                   `xml:"xmlns,attr"`
	Triggers   launcherStartupTaskTriggersXML   `xml:"Triggers"`
	Principals launcherStartupTaskPrincipalsXML `xml:"Principals"`
	Settings   launcherStartupTaskSettingsXML   `xml:"Settings"`
	Actions    launcherStartupTaskActionsXML    `xml:"Actions"`
}

type launcherUpdateHealthHello struct {
	Nonce    string                                `json:"nonce"`
	Identity launcherUpdateProcessIdentity `json:"identity"`
}

type launcherUpdateHelperChallenge struct {
	ParentIdentity launcherUpdateProcessIdentity `json:"parentIdentity"`
	ChildIdentity  launcherUpdateProcessIdentity `json:"childIdentity"`
	ChildNonce     string                                `json:"childNonce"`
	ParentNonce    string                                `json:"parentNonce"`
	Plan           []uint8                               `json:"plan"`
}

type launcherUpdateHelperHello struct {
	Identity   launcherUpdateProcessIdentity `json:"identity"`
	ChildNonce string                                `json:"childNonce"`
}

type launcherUpdateHelperReady struct {
	Proof string `json:"proof"`
}

type launcherUpdatePackageRange struct {
	Start int64
	End   int64
}

type launcherUpdatePlan struct {
	SchemaVersion  int                              `json:"schemaVersion"`
	Nonce          string                           `json:"nonce"`
	HealthNonce    string                           `json:"healthNonce"`
	ParentPID      int                              `json:"parentPid"`
	InstallDir     string                           `json:"installDir"`
	UpdateRoot     string                           `json:"updateRoot"`
	UpdateDir      string                           `json:"updateDir"`
	ExecutableName string                           `json:"executableName"`
	LogFile        string                           `json:"logFile"`
	JournalFile    string                           `json:"journalFile"`
	PackageSize    int64                            `json:"packageSize"`
	PackageSHA256  string                           `json:"packageSha256"`
	PackageRoot    string                           `json:"packageRoot"`
	Files          []launcherUpdatePlanFile `json:"files"`
	Deletes        []string                         `json:"deletes,omitempty"`
}

type launcherUpdatePlanFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	Mode   uint32 `json:"mode"`
}

type launcherUpdateProcessIdentity struct {
	PID           uint32 `json:"pid"`
	ParentPID     uint32 `json:"parentPid"`
	CreatedAt     int64  `json:"createdAt"`
	SessionID     uint32 `json:"sessionId"`
	UserSID       string `json:"userSid"`
	IntegrityRID  uint32 `json:"integrityRid"`
	ImagePath     string `json:"imagePath"`
	ImageSHA256   string `json:"imageSha256"`
	VolumeSerial  uint32 `json:"volumeSerial"`
	FileIndexHigh uint32 `json:"fileIndexHigh"`
	FileIndexLow  uint32 `json:"fileIndexLow"`
}

type launcherUpdateTransaction struct {
	SchemaVersion int                                     `json:"schemaVersion"`
	Nonce         string                                  `json:"nonce"`
	State         string                                  `json:"state"`
	InstallDir    string                                  `json:"installDir"`
	StagingDir    string                                  `json:"stagingDir"`
	BackupDir     string                                  `json:"backupDir"`
	JournalFile   string                                  `json:"journalFile"`
	Executable    string                                  `json:"executable"`
	HealthNonce   string                                  `json:"healthNonce"`
	Files         []launcherUpdateTransactionFile `json:"files"`
	UpdatedAt     string                                  `json:"updatedAt"`
}

type launcherUpdateTransactionFile struct {
	Path         string `json:"path"`
	Delete       bool   `json:"delete,omitempty"`
	Existed      bool   `json:"existed,omitempty"`
	Status       string `json:"status"`
	ExpectedSize int64  `json:"expectedSize,omitempty"`
	ExpectedHash string `json:"expectedSha256,omitempty"`
}

type launcherWindowLayoutSnapshot struct {
	Width     int
	Height    int
	RelativeX int
	RelativeY int
}

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type mouseGestureInput struct {
	Type uint32
	MI   mouseGestureMouseInput
}

type mouseGestureMouseInput struct {
	DX        int32
	DY        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type oledBlackoutHotkeyBinding struct {
	ProfileID string
	Hotkey    string
}

type oledBlackoutHotkeyEvent struct {
	ProfileID string
	Hotkey    string
}

type oledBlackoutHotkeyUpdateResult struct {
	RegisteredProfileIDs []string
	Errors               map[string]string
}

type oledBlackoutIdlePauseTargetScreen struct {
	ID             string
	Bounds         application.Rect
	PhysicalBounds application.Rect
}

type openMeteoAirResponse struct {
	Current struct {
		Time  string  "json:\"time\""
		USAQI int     "json:\"us_aqi\""
		PM25  float64 "json:\"pm2_5\""
		PM10  float64 "json:\"pm10\""
	} `json:"current"`
}

type openMeteoForecastResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Current   struct {
		Time         string  "json:\"time\""
		Temperature  float64 "json:\"temperature_2m\""
		ApparentTemp float64 "json:\"apparent_temperature\""
		Humidity     int     "json:\"relative_humidity_2m\""
		WeatherCode  int     "json:\"weather_code\""
		WindSpeed    float64 "json:\"wind_speed_10m\""
	} `json:"current"`
	Hourly struct {
		Time              []string  "json:\"time\""
		Temperature       []float64 "json:\"temperature_2m\""
		WeatherCode       []int     "json:\"weather_code\""
		PrecipProbability []int     "json:\"precipitation_probability\""
		Precipitation     []float64 "json:\"precipitation\""
		UV                []float64 "json:\"uv_index\""
	} `json:"hourly"`
	Daily struct {
		Time              []string  "json:\"time\""
		WeatherCode       []int     "json:\"weather_code\""
		TemperatureMax    []float64 "json:\"temperature_2m_max\""
		TemperatureMin    []float64 "json:\"temperature_2m_min\""
		PrecipProbability []int     "json:\"precipitation_probability_max\""
		UVMax             []float64 "json:\"uv_index_max\""
		Sunrise           []string  "json:\"sunrise\""
		Sunset            []string  "json:\"sunset\""
	} `json:"daily"`
}

type openMeteoGeocodingItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Admin1      string  `json:"admin1"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

type openMeteoGeocodingResponse struct {
	Results []openMeteoGeocodingItem `json:"results"`
}

type platformDesktopWidgetNotifier struct{}

type pluginRuntimeMeta struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Category    string                 `json:"category"`
	APIVersion  string                 `json:"apiVersion"`
	UI          PluginUIConfig `json:"ui"`
	Permissions []string               `json:"permissions"`
	Messages    map[string]interface{} `json:"messages,omitempty"`
}

type qweatherAirResponse struct {
	Code string `json:"code"`
	Now  struct {
		PubTime  string
		AQI      string
		Category string
		Primary  string
		PM2P5    string
		PM10     string
	} `json:"now"`
}

type qweatherResponse struct {
	Code string `json:"code"`
	Now  struct {
		ObsTime   string
		Temp      string
		FeelsLike string
		Humidity  string
		Icon      string
		WindSpeed string
	} `json:"now"`
	Hourly []struct {
		FxTime string
		Temp   string
		Icon   string
		Pop    string
		Precip string
	} `json:"hourly"`
	Daily []struct {
		FxDate  string
		TempMax string
		TempMin string
		IconDay string
		Precip  string
		UVIndex string
		Sunrise string
		Sunset  string
	} `json:"daily"`
}

type rawIndexEntry struct {
	FRN       uint64
	ParentFRN uint64
	Name      string
	ModTime   uint32
	IsDir     bool
}

type remoteIconifyCollectionRef struct {
	Name string `json:"name"`
}

type remoteIconifySearchResponse struct {
	Icons       []string                                      `json:"icons"`
	Collections map[string]remoteIconifyCollectionRef `json:"collections"`
}

type screenshotAccessibleVtbl struct {
	QueryInterface         uintptr
	AddRef                 uintptr
	Release                uintptr
	GetTypeInfoCount       uintptr
	GetTypeInfo            uintptr
	GetIDsOfNames          uintptr
	Invoke                 uintptr
	GetAccParent           uintptr
	GetAccChildCount       uintptr
	GetAccChild            uintptr
	GetAccName             uintptr
	GetAccValue            uintptr
	GetAccDescription      uintptr
	GetAccRole             uintptr
	GetAccState            uintptr
	GetAccHelp             uintptr
	GetAccHelpTopic        uintptr
	GetAccKeyboardShortcut uintptr
	GetAccFocus            uintptr
	GetAccSelection        uintptr
	GetAccDefaultAction    uintptr
	AccSelect              uintptr
	AccLocation            uintptr
	AccNavigate            uintptr
	AccHitTest             uintptr
	AccDoDefaultAction     uintptr
	PutAccName             uintptr
	PutAccValue            uintptr
}

type screenshotCursorInfo struct {
	Size      uint32
	Flags     uint32
	Cursor    uintptr
	ScreenPos w32.POINT
}

type screenshotDisplayCaptureInfo struct {
	AdapterIndex          int
	OutputIndex           int
	AdapterName           string
	DeviceName            string
	Bounds                image.Rectangle
	Rotation              uint32
	AttachedToDesktop     bool
	BitsPerColor          uint32
	ColorSpace            uint32
	HDR                   bool
	AdvancedColor         bool
	MinLuminance          float32
	MaxLuminance          float32
	MaxFullFrameLuminance float32
}

type screenshotGDIScreenCaptureBackend struct{}

type screenshotHDRToneMapOptions struct {
	SDRWhiteLevel    float32
	SDRWhiteOutput   float32
	MaxInputLevel    float32
	HighlightRolloff float32
}

type screenshotOleVariant struct {
	VT        uint16
	Reserved1 uint16
	Reserved2 uint16
	Reserved3 uint16
	Val       int64
}

type screenshotPinnedWindowSnapshot struct {
	WindowName   string `json:"windowName"`
	ImageData    string `json:"imageData,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	Path         string `json:"path,omitempty"`
	ContentKey   string `json:"contentKey,omitempty"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Scale        string `json:"scale"`
	Opacity      string `json:"opacity"`
	ClickThrough bool   `json:"clickThrough"`
	AutoShow     bool   `json:"autoShow"`
	Order        int64  `json:"order"`
}

type screenshotPreviewWindowUpdate struct {
	Version   int64  `json:"version"`
	ImageURL  string `json:"imageUrl,omitempty"`
	ImageData string `json:"imageData,omitempty"`
	ReadyURL  string `json:"readyURL"`
	OpenURL   string `json:"openURL,omitempty"`
}

type screenshotRouteCapability struct {
	Token      string
	Owner      string
	Generation int64
}

type screenshotSelectionToolbarAction struct {
	SessionID int64  `json:"sessionId"`
	Action    string `json:"action"`
	Tool      string `json:"tool,omitempty"`
	Color     string `json:"color,omitempty"`
	LineWidth int    `json:"lineWidth,omitempty"`
	Preview   bool   `json:"preview,omitempty"`
	Open      bool   `json:"open,omitempty"`
}

type screenshotSelectionToolbarMessages struct {
	LanguageCode   string `json:"languageCode"`
	ResolvedTheme  string `json:"resolvedTheme"`
	UIScalePercent int    `json:"uiScalePercent"`
	ToolbarLabel   string `json:"toolbarLabel"`
	Pointer        string `json:"pointer"`
	Delete         string `json:"delete"`
	Rect           string `json:"rect"`
	Ellipse        string `json:"ellipse"`
	Line           string `json:"line"`
	Arrow          string `json:"arrow"`
	Pen            string `json:"pen"`
	Text           string `json:"text"`
	Number         string `json:"number"`
	Mosaic         string `json:"mosaic"`
	Blur           string `json:"blur"`
	DragToolbar    string `json:"dragToolbar"`
	Color          string `json:"color"`
	CustomColor    string `json:"customColor"`
	Eyedropper     string `json:"eyedropper"`
	ConfirmColor   string `json:"confirmColor"`
	LineWidth      string `json:"lineWidth"`
	CornerRadius   string `json:"cornerRadius"`
	Undo           string `json:"undo"`
	Redo           string `json:"redo"`
	Clear          string `json:"clear"`
	Confirm        string `json:"confirm"`
	Cancel         string `json:"cancel"`
	Processing     string `json:"processing"`
}

type screenshotSelectionToolbarState struct {
	SessionID       int64  `json:"sessionId,string"`
	Tool            string `json:"tool"`
	Color           string `json:"color"`
	LineWidth       int    `json:"lineWidth"`
	CanUndo         bool   `json:"canUndo"`
	CanRedo         bool   `json:"canRedo"`
	CanClear        bool   `json:"canClear"`
	ColorPickerOpen bool   `json:"colorPickerOpen"`
	Processing      bool   `json:"processing"`
}

type screenshotUIAutomationVtbl struct {
	QueryInterface              uintptr
	AddRef                      uintptr
	Release                     uintptr
	CompareElements             uintptr
	CompareRuntimeIds           uintptr
	GetRootElement              uintptr
	ElementFromHandle           uintptr
	ElementFromPoint            uintptr
	GetFocusedElement           uintptr
	GetRootElementBuildCache    uintptr
	ElementFromHandleBuildCache uintptr
	ElementFromPointBuildCache  uintptr
	GetFocusedElementBuildCache uintptr
	CreateTreeWalker            uintptr
	ControlViewWalker           uintptr
	ContentViewWalker           uintptr
	RawViewWalker               uintptr
	RawViewCondition            uintptr
	ControlViewCondition        uintptr
	ContentViewCondition        uintptr
	CreateCacheRequest          uintptr
	CreateTrueCondition         uintptr
	CreateFalseCondition        uintptr
	CreatePropertyCondition     uintptr
}

type shFileInfo struct {
	HIcon         uintptr
	IIcon         int32
	DwAttributes  uint32
	SzDisplayName [260]uint16
	SzTypeName    [80]uint16
}

type startMenuRoot struct {
	Path   string
	Source string
}

type twoFactorParsedImportEntry struct {
	Entry       TwoFactorEntryConfig
	SecretBytes []uint8
}

type twoFactorTimeCache struct {
	HasSuccess   bool
	Offset       int64
	SourceName   string
	SyncedAt     time.Time
	SampleCount  int
	ErrorMessage string
}

type twoFactorTimeSample struct {
	SourceName string
	SourceID   string
	Offset     int64
	RTT        int64
}

type twoFactorTimeTarget struct {
	Name string
	URL  string
}

type volumeIndexCheckpointLayout struct {
	FileSize      int64
	NodeOffset    int64
	NodeBytes     int64
	SortedOffset  int64
	SortedBytes   int64
	NameOffset    int64
	NameBytes     int64
	ExpectedBytes int64
}

type volumeIndexJournalChange struct {
	FRN         uint64
	ParentFRN   uint64
	Reason      uint32
	ModTime     uint32
	Name        string
	IsDirectory bool
}

type volumeIndexMappedSectionPlan struct {
	MapOffset     int64
	MapBytes      int64
	SectionOffset int64
	SectionBytes  int64
}

type volumeIndexOverlayStats struct {
	BaseCount      int
	DeltaCount     int
	TombstoneCount int
	TotalCount     int
	Ratio          float64
}

type volumeIndexPersistenceMeta struct {
	Version                uint32 `json:"version"`
	IndexVersion           uint32 `json:"indexVersion"`
	RootFRN                uint64 `json:"rootFrn"`
	JournalID              uint64 `json:"journalId"`
	LastUSN                int64  `json:"lastUsn"`
	GeneratedAtUnixNano    int64  `json:"generatedAtUnixNano"`
	LastMutationAtUnixNano int64  `json:"lastMutationAtUnixNano"`
	LastSavedAtUnixNano    int64  `json:"lastSavedAtUnixNano"`
}

type volumeIndexPersistencePaths struct {
	CheckpointPath  string
	MetaPath        string
	WALPath         string
	NameTrigramPath string
	BucketPath      string
	PinyinPath      string
}

type volumeIndexSystemInfo struct {
	ProcessorArchitecture     uint16
	Reserved                  uint16
	PageSize                  uint32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       uintptr
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
}

type volumeIndexWALRecord struct {
	Version   uint32                             `json:"version"`
	JournalID uint64                             `json:"journalId"`
	StartUSN  int64                              `json:"startUsn"`
	EndUSN    int64                              `json:"endUsn"`
	Changes   []volumeIndexJournalChange `json:"changes"`
}

type volumeSearchCandidate struct {
	NodeIndex int32
	Name      string
	FRN       uint64
	ModTime   uint32
	IsDir     bool
	IsPrefix  bool
	IsExact   bool
	MatchKind uint8
}

type volumeSearchMatch struct {
	NodeIndex int32
	IsPrefix  bool
	IsExact   bool
	MatchKind uint8
}

type volumeSearchPerfStats struct {
	MatchDuration       int64
	PathResolveDuration int64
	TotalDuration       int64
	ResultCount         int
	TotalMatchCount     int
}

type volumeSearchWithPath struct {
	Name     string
	Path     string
	FRN      uint64
	ModTime  uint32
	IsDir    bool
	IsPrefix bool
	IsExact  bool
}

type volumeWatcher struct {
	Root      string
	Handle    uintptr
	JournalID uint64
	Buffer    []uint8
	Queue     chan fileSearchJournalBatch
	Stop      chan struct{}
	Done      chan struct{}
}

type weatherAPIResponse struct {
	Location struct {
		Name      string
		Region    string
		Country   string
		TZID      string
		Localtime string
		Lat       float64
		Lon       float64
	} `json:"location"`
	Current struct {
		LastUpdated string  "json:\"last_updated\""
		TempC       float64 "json:\"temp_c\""
		FeelsLikeC  float64 "json:\"feelslike_c\""
		Humidity    int     "json:\"humidity\""
		WindKPH     float64 "json:\"wind_kph\""
		UV          float64 "json:\"uv\""
		Condition   struct {
			Code int "json:\"code\""
		} "json:\"condition\""
		AirQuality struct {
			USAQI int     "json:\"us-epa-index\""
			PM25  float64 "json:\"pm2_5\""
			PM10  float64 "json:\"pm10\""
		} "json:\"air_quality\""
	} `json:"current"`
	Forecast struct {
		ForecastDay []struct {
			Date string "json:\"date\""
			Day  struct {
				MaxTempC          float64
				MinTempC          float64
				UV                float64
				DailyChanceOfRain int "json:\"daily_chance_of_rain\""
				Condition         struct {
					Code int "json:\"code\""
				} "json:\"condition\""
			} "json:\"day\""
			Astro struct {
				Sunrise string
				Sunset  string
			} "json:\"astro\""
			Hour []struct {
				Time         string "json:\"time\""
				TempC        float64
				PrecipMM     float64
				UV           float64
				ChanceOfRain int "json:\"chance_of_rain\""
				Condition    struct {
					Code int "json:\"code\""
				} "json:\"condition\""
			} "json:\"hour\""
		} "json:\"forecastday\""
	} `json:"forecast"`
}

type weatherAPISearchItem struct {
	ID      int64 `json:"id"`
	Name    string
	Region  string
	Country string
	Lat     float64
	Lon     float64
	URL     string `json:"url"`
}

type weatherAPITimezoneResponse struct {
	Location struct {
		TZID string "json:\"tz_id\""
	} `json:"location"`
}

type windowManagementMonitorInfo struct {
	Size    uint32
	Monitor windowManagementRECT
	Work    windowManagementRECT
	Flags   uint32
}

type windowManagementOpacitySnapshot struct {
	HWND       uintptr
	ProcessID  uint32
	ExStyle    uintptr
	WasLayered bool
	ColorKey   uint32
	Alpha      uint8
	Flags      uint32
}

type windowManagementPoint struct {
	X int32
	Y int32
}

type windowManagementRECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type workspaceDataManifestEntry struct {
	Size   int64
	SHA256 string
}

type BootstrapService struct {
	workspace                           WorkspaceLayout
	configStore                         *launcherConfigStore
	bootstrapSnapshotRefresh            sync.Mutex
	bootstrapSnapshot                   BootstrapSnapshot
	bootstrapSnapshotReady              bool
	app                                 *application.App
	launcherWindow                      application.Window
	globalHotkey                        launcherGlobalHotkeyManager
	hotkeyCapture                       bool
	hotkeyCaptureLegacyRef              int
	hotkeyCaptureOwners                 map[string]struct{}
	hotkeyCaptureOperation              sync.Mutex
	searchCategoryShortcutActive        bool
	hotkey                              string
	hotkeyEnabled                       bool
	summonOnlyHotkey                    string
	summonOnlyHotkeyEnabled             bool
	screenshotHotkey                    string
	screenshotHotkeyEnabled             bool
	screenshotQRCodeHotkey              string
	screenshotQRCodeHotkeyEnabled       bool
	screenshotAllScreensHotkey          string
	screenshotAllScreensHotkeyEnabled   bool
	screenshotScrollingHotkey           string
	screenshotScrollingHotkeyEnabled    bool
	screenshotActiveWindowHotkey        string
	screenshotActiveWindowHotkeyEnabled bool
	screenshotFeatureEnabled            bool
	screenshotHotkeyActive              bool
	screenshotSaveFormatOverride        string
	screenshotCaptureControlsSetting    bool
	screenshotSelectionConfirmSetting   bool
	screenshotCaptureCursorSetting      bool
	screenshotCaptureSettingsReady      bool
	screenshotAnnotationLineWidth       int
	screenshotCornerRadius              int
	bindings                            launcherHotkeyBindings
	hotkeyRegistrationErrors            map[string]string
	bookmarkSources                     []BookmarkSource
	twoFactor                           *twoFactorService
	memoryRelease                       *memoryReleaseService
	oledBlackout                        *oledBlackoutService
	windowManagement                    *windowManagementService
	mouseGestures                       *mouseGestureService
	assets                              *launcherAssetService
	backgroundAssetLock                 sync.Mutex
	backgroundAssetOwner                *launcherAssetService
	backgroundAssetPath                 string
	backgroundAssetSize                 int64
	backgroundAssetModifiedAt           int64
	backgroundAssetURL                  string
	iconAssetLock                       sync.Mutex
	iconAssetOwner                      *launcherAssetService
	iconAssetURLs                       map[string]launcherConfigIconAssetCacheEntry
	iconAssetSequence                   uint64
	screenshotPin                       *screenshotPinWindowService
	screenshotPreview                   *screenshotPreviewWindowService
	screenshotSelectionToolbar          *screenshotSelectionToolbarWindowService
	pluginWindows                       *pluginWindowService
	inputMonitor                        *inputMonitorService
	fileIndex                           fileSearchRuntimeWarmer
	fileLocator                         *fileLocatorService
	desktopWidgets                      *desktopWidgetService
	resetLauncherSizeOnNextShow         bool
	startupTrayMode                     bool
	pendingLauncherReveal               bool
	allowLauncherWindowClose            bool
	launcherWindowDestroyedToTray       bool
	launcherFrontendReady               bool
	launcherSidebarPinnedOpen           bool
	lastLauncherHotkeyAction            string
	lastLauncherHotkeyAt                time.Time
	launcherUpdateProgress              LauncherUpdateProgressState
	launcherUpdateCancel                func()
	launcherUpdateTaskID                int64
	launcherUpdateDir                   string
	launcherUpdateDone                  chan struct{}
	launcherUpdateShuttingDown          bool
	launcherUpdateNotificationShown     bool
	pendingQRCodeDecodeResult           *QRCodeDecodeResult
	verticalMaximizeSnapshot            *launcherWindowLayoutSnapshot
	pendingFileSearchResidentWarm       bool
	startupTaskSync                     func(bool, int) error
	workspaceTransaction                sync.Mutex
	workspaceDataMaintenance            workspaceDataMaintenanceGate
	lock                                sync.Mutex
}

type FileIndexService struct {
	bootstrap             *BootstrapService
	ctx                   context.Context
	mu                    sync.RWMutex
	snapshotMu            sync.RWMutex
	appliedSignature      string
	lastConfigSignature   string
	initialized           bool
	indexes               map[string]*VolumeIndex
	watchers              map[string]*volumeWatcher
	usnFollowers          map[string]*fileSearchUSNFollower
	volumes               []FileSearchVolume
	volumeMeta            map[string]fileSearchVolumeMeta
	journalRuntime        map[string]fileSearchJournalRuntimeState
	ignoreMatcher         fileSearchIgnoreMatcher
	ready                 bool
	indexedCount          int
	lastIndexedAt         time.Time
	lastFileSearchUsedAt  time.Time
	lastFileSearchQueryAt time.Time
	lastError             string
	runtimeReleased       bool
	runtimeCooling        bool
	recoveryPending       map[string]struct{}
	journalTransitions    map[string]int
	memoryPressureChecks  int
	warmRuntime           fileSearchWarmRuntimeState
	lastReleaseReason     string
	mergeRuntime          map[string]fileSearchCheckpointRuntimeState
	restaticRuntime       fileSearchRestaticRuntimeState
	snapshotCache         FileSearchState
	stopPersistence       chan struct{}
	stopPersistenceOnce   sync.Once
	mergeSignal           chan struct{}
	checkpointSignal      chan struct{}
	restaticSignal        chan struct{}
	mergePending          map[string]struct{}
	checkpointPending     map[string]struct{}
	mergeWorker           func([]string, time.Time)
	checkpointRuntime     map[string]fileSearchCheckpointRuntimeState
	checkpointWorker      func([]string, time.Time)
	activeSearchQueries   int
	configurationMu       sync.Mutex
	maintenanceMu         sync.Mutex
	maintenanceRunning    bool
	maintenancePendingAll bool
	maintenanceTargets    map[string]struct{}
	pinyinGeneration      uint64
	pinyinCancel          func()
	pinyinRunning         bool
	pinyinWorkerDone      chan struct{}
	pinyinTargetVolumes   int
	pinyinReadyVolumes    int
	pinyinLastError       string
	pinyinFailedIdentity  map[string]volumePinyinIdentity
}

type GestureSettings struct {
	GestureButtons           []string `json:"gestureButtons"`
	StartDistancePx          int      `json:"startDistancePx"`
	StartTimeoutMs           int      `json:"startTimeoutMs"`
	StopTimeoutMs            int      `json:"stopTimeoutMs"`
	AllowDiagonal            bool     `json:"allowDiagonal,omitempty"`
	AutoDisableFullscreen    bool     `json:"autoDisableFullscreen,omitempty"`
	TargetWindowUnderPointer bool     `json:"targetWindowUnderPointer,omitempty"`
	ShowTrail                bool     `json:"showTrail,omitempty"`
	ShowGestureName          bool     `json:"showGestureName,omitempty"`
	FadeAfterExecution       bool     `json:"fadeAfterExecution,omitempty"`
	startDistancePxSet       bool
	startTimeoutMsSet        bool
	stopTimeoutMsSet         bool
}

type LauncherNetworkAccess interface {
	Do(*http.Request, int64) (*http.Response, error)
	Get(string, int64) (*http.Response, error)
}

type LauncherNetworkRedirectPolicyAccess interface {
	DoWithRedirectPolicy(*http.Request, int64, func(*http.Request, []*http.Request) error) (*http.Response, error)
}

type PluginHostService struct {
	bootstrap *BootstrapService
}

type VolumeIndex struct {
	mu                      sync.RWMutex
	journalMu               sync.Mutex
	Root                    string
	RootFRN                 uint64
	Nodes                   []IndexNode
	NamePool                []uint8
	SortedIndices           []int32
	JournalID               uint64
	LastUSN                 int64
	GeneratedAt             time.Time
	LastMutationAt          time.Time
	LastSavedAt             time.Time
	activeCount             int
	dirty                   bool
	changeCaught            uint32
	prefixBuckets           namePrefixBucketCounts
	nameTrigramMu           sync.Mutex
	nameTrigram             *volumeNameTrigramIndex
	nameBigramMu            sync.Mutex
	nameBigram              *volumeNameBigramIndex
	readProvider            volumeIndexReadProvider
	pinyinEnabled           bool
	pinyin                  *volumePinyinIndex
	pinyinBaseUSN           int64
	tombstones              map[int32]struct{}
	baseSubtreeActiveCounts []uint32
	deltaNodes              []volumeIndexDeltaNode
	deltaByFRN              map[uint64]int32
	deltaReplaced           map[int32]struct{}
	runtimeVersion          uint64
}

type WindowManagementConfig struct {
	CursorWrapHorizontalEnabled bool                               `json:"cursorWrapHorizontalEnabled,omitempty"`
	CursorWrapVerticalEnabled   bool                               `json:"cursorWrapVerticalEnabled,omitempty"`
	Target                      WindowManagementTarget     `json:"target,omitempty"`
	OpacityPercent              int                                `json:"opacityPercent"`
	Resolution                  WindowManagementSize       `json:"resolution,omitempty"`
	BorderlessSnapshots         []WindowFullscreenSnapshot `json:"borderlessSnapshots,omitempty"`
	FullscreenSnapshots         []WindowFullscreenSnapshot `json:"fullscreenSnapshots,omitempty"`
	opacityPercentSet           bool
}

type audioClient struct {
	enumerator       *audioIMMDeviceEnumerator
	displayNameCache map[string]string
	iconCache        map[string]string
}

type audioIMMDevice struct {
	lpVtbl *audioIMMDeviceVtbl
}

type audioIMMDeviceEnumerator struct {
	lpVtbl *audioIMMDeviceEnumeratorVtbl
}

type audioSessionAccumulator struct {
	session      AudioSession
	volumeSum    int
	sessionCount int
	allMuted     bool
	anyActive    bool
}

type backupEntry struct {
	path    string
	name    string
	date    string
	modTime time.Time
}

type bookmarkFaviconDBCacheEntry struct {
	db         *sql.DB
	cleanup    func()
	modifiedAt time.Time
	size       int64
	users      sync.WaitGroup
	closeOnce  sync.Once
}

type bookmarkFaviconDBLease struct {
	mu       sync.RWMutex
	entry    *bookmarkFaviconDBCacheEntry
	released bool
}

type bookmarkIconDiskCacheEntry struct {
	path       string
	size       int64
	modifiedAt time.Time
}

type bookmarkIconDownloadFailureEntry struct {
	failedAt time.Time
}

type chromiumBookmarkTraversalItem struct {
	node     chromiumBookmarkNode
	ancestry []string
	depth    int
}

type configBackedLauncherNetworkAccess struct {
	configPath string
}

type deliveryEntry struct {
	key string
	at  string
}

type desktopNotePendingDraft struct {
	body             string
	expectedRevision int64
}

type desktopWeatherInflight struct {
	done     chan struct{}
	snapshot DesktopWeatherSnapshot
	err      error
}

type desktopWidgetAudioPlayer interface {
	Play(desktopReminderAudio) error
}

type desktopWidgetDueNotification struct {
	deliveryKey string
	entityID    string
	title       string
	message     string
	audio       *desktopReminderAudio
}

type desktopWidgetNotifier interface {
	Notify(string, string, bool) error
}

type desktopWidgetScheduler struct{
	service any
	wake chan struct {}
	stop chan struct {}
	done chan struct {}
	once sync.Once
	stopOnce sync.Once
}

type desktopWidgetService struct {
	store        *launcherWidgetStore
	notifier     desktopWidgetNotifier
	audio        desktopWidgetAudioPlayer
	weather      *desktopWidgetWeatherService
	scheduler    *desktopWidgetScheduler
	mu           sync.RWMutex
	app          *application.App
	enabled      bool
	status       string
	lastError    error
	shuttingDown bool
	draftMu      sync.Mutex
	drafts       map[string]desktopNotePendingDraft
	draftTimer   *time.Timer
}

type desktopWidgetWeatherService struct {
	store          *launcherWidgetStore
	network        LauncherNetworkAccess
	ctx            context.Context
	cancel         func()
	mu             sync.Mutex
	inflight       map[string]*desktopWeatherInflight
	providerErrors map[string]string
}

type endpointRef struct {
	id     string
	device *audioIMMDevice
}

type fileLocatorBinaryNode struct {
	op    int
	left  fileLocatorBooleanNode
	right fileLocatorBooleanNode
}

type fileLocatorBooleanExpression struct {
	root             fileLocatorBooleanNode
	positiveMatchers []fileLocatorTermMatcher
}

type fileLocatorBooleanNode interface {
	Eval(string) bool
}

type fileLocatorContextReader struct {
	ctx    context.Context
	reader io.Reader
}

type fileLocatorNearNode struct {
	left     fileLocatorBooleanNode
	right    fileLocatorBooleanNode
	distance int
}

type fileLocatorNotNode struct {
	child fileLocatorBooleanNode
}

type fileLocatorPathFilter struct {
	includes []*fileLocatorStringMatcher
	excludes []*fileLocatorStringMatcher
}

type fileLocatorPlainMatcher struct {
	query     string
	queryFold string
	matchCase bool
}

type fileLocatorPreparedSearch struct {
	request         FileLocatorConfig
	roots           []string
	fileNameMatcher *fileLocatorStringMatcher
	contentMatcher  *fileLocatorStringMatcher
	pathFilter      *fileLocatorPathFilter
}

type fileLocatorRegexMatcher struct {
	regex *regexp.Regexp
}

type fileLocatorRuneWindow struct {
	start int
	end   int
}

type fileLocatorService struct {
	lock         sync.RWMutex
	pauseCond    *sync.Cond
	generation   uint64
	state        FileLocatorState
	detailByPath map[string]FileLocatorResultDetail
	cancel       func()
	contentScan  chan struct{}
}

type fileLocatorStringMatcher struct {
	query            string
	queryFold        string
	mode             string
	matchCase        bool
	booleanScope     string
	regex            *regexp.Regexp
	plainWildcard    *regexp.Regexp
	booleanExpr      *fileLocatorBooleanExpression
	wholeWordMatcher fileLocatorTermMatcher
}

type fileLocatorTermMatcher interface {
	Exists(string) bool
	Ranges(string) []FileLocatorTextRange
}

type fileLocatorTermNode struct {
	matcher fileLocatorTermMatcher
}

type fileLocatorToken struct {
	kind     int
	value    string
	distance int
}

type fileLocatorWholeWordMatcher struct {
	queryRunes []int32
	queryFold  []int32
	matchCase  bool
}

type fileSearchCandidateHeap struct {
	items    []scoredFileSearchCandidate
	resolver *fileSearchCandidatePathResolver
}

type fileSearchCandidatePathResolver struct {
	pathCaches          map[*VolumeIndex]*nodePathCache
	pathResolveDuration int64
}

type fileSearchCheckpointCandidate struct {
	reason         string
	root           string
	priority       int
	cost           int
	overlayEntries int
	overlayRatio   float64
	walBytes       int64
	changeCaught   uint32
}

type fileSearchCheckpointCandidateSource struct {
	root string
	idx  *VolumeIndex
}

type fileSearchCheckpointRuntimeState struct {
	running       bool
	retryCount    int
	lastError     string
	lastAttemptAt time.Time
	lastSuccessAt time.Time
	nextRetryAt   time.Time
}

type fileSearchIgnoreMatcher struct {
	nameRules map[string]struct{}
	pathRules []string
}

type fileSearchIndexCandidateSearchResult struct {
	index      *VolumeIndex
	candidates []volumeSearchCandidate
	totalCount int
	perfStats  volumeSearchPerfStats
}

type fileSearchJournalBatch struct {
	records []uint8
	nextUsn int64
	readAt  time.Time
}

type fileSearchJournalChange struct {
	frn       uint64
	parentFRN uint64
	reason    uint32
	modTime   uint32
	name      string
	attrs     uint32
}

type fileSearchJournalRuntimeState struct {
	pendingBatches int
	pendingBytes   int64
	pendingSince   time.Time
	lastReadAt     time.Time
	lastApplyAt    time.Time
	applying       bool
}

type fileSearchPinyinBuildTarget struct {
	root     string
	idx      *VolumeIndex
	path     string
	identity volumePinyinIdentity
}

type fileSearchPinyinFuzzyMatcher struct {
	contains           *regexp.Regexp
	exact              *regexp.Regexp
	prefix             *regexp.Regexp
	requiredSignatures []uint64
}

type fileSearchRestaticRuntimeState struct {
	running       bool
	pending       map[string]time.Time
	retry         map[string]fileSearchCheckpointRuntimeState
	lastRunAt     time.Time
	runningRoot   string
	lastSuccessAt time.Time
}

type fileSearchRuleMatcher struct {
	empty              bool
	simpleByHash       map[uint64][][]uint8
	compoundByLastByte map[uint8][][]uint8
	splitArchiveFast   bool
	glob               *regexp.Regexp
	simpleCount        int
	compoundCount      int
	globCount          int
	signature          string
}

type fileSearchRuntimeWarmer interface {
	scheduleFileSearchMaintenance(map[string]struct{}, bool)
}

type fileSearchSortComparer struct {
	spec     fileSearchSortSpec
	collator *collate.Collator
}

type fileSearchSortedCandidateHeap struct {
	items    []scoredFileSearchCandidate
	resolver *fileSearchCandidatePathResolver
	comparer *fileSearchSortComparer
}

type fileSearchWarmRuntimeState struct {
	running        bool
	targetVolumes  int
	completedRoots map[string]struct{}
	lastTarget     string
}

type firefoxBookmarkGraphFrame struct {
	id      int64
	depth   int
	exiting bool
}

type firefoxBookmarkTraversalItem struct {
	node     firefoxBookmarkNode
	ancestry []string
	depth    int
}

type hashWriter interface {
	Sum([]uint8) []uint8
	Write([]uint8) (int, error)
}

type inlineIcon struct {
	ref     string
	metrics launcherConfigIconMetrics
}

type inputMonitorKeyboardHookEvent struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type inputMonitorMouseHookEvent struct {
	point       inputMonitorPoint
	mouseData   uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type inputMonitorPlatformCommand struct {
	close  bool
	result chan error
}

type inputMonitorPoint struct {
	x int32
	y int32
}

type inputMonitorRawEvent struct {
	generation uint64
	message    uint32
	keyboard   bool
	keyEvent   inputMonitorKeyboardHookEvent
	mouseEvent inputMonitorMouseHookEvent
}

type inputMonitorService struct {
	lifecycleMu                sync.Mutex
	lock                       sync.Mutex
	supported                  bool
	running                    bool
	keyboard                   bool
	mouse                      bool
	startedAt                  time.Time
	lastError                  string
	nextID                     uint64
	events                     []InputMonitorEvent
	eventHead                  int
	eventSize                  int
	droppedEvents              uint64
	rawDropped                 atomic.Uint64
	pressedKeys                map[uint32]bool
	maxEvents                  int
	commandChan                chan inputMonitorPlatformCommand
	ready                      chan error
	done                       chan struct{}
	threadID                   uint32
	keyboardHook               uintptr
	mouseHook                  uintptr
	keyboardHookRef            uintptr
	mouseHookRef               uintptr
	platformGeneration         atomic.Uint64
	platformHealthManaged      bool
	platformThreadAlive        bool
	platformMessageLoopHealthy bool
	platformKeyboardInstalled  bool
	platformMouseInstalled     bool
	closeOnce                  sync.Once
	owners                     map[string]InputMonitorStartRequest
	closed                     bool
	startOverride              func(InputMonitorStartRequest) error
	stopOverride               func() error
	closeOverride              func() error
}

type item struct {
	key   string
	entry *screenshotDXGIOutputCaptureEntry
	used  time.Time
}

type launcherAssetEntry struct {
	id          string
	namespace   string
	version     int64
	contentType string
	data        []uint8
	filePath    string
	size        int64
	createdAt   time.Time
	accessedAt  time.Time
	expiresAt   time.Time
}

type launcherAssetLimits struct {
	maxEntries        int
	maxItemBytes      int64
	maxNamespaceBytes int64
	maxTotalBytes     int64
}

type launcherAssetService struct {
	lock           sync.Mutex
	entries        map[string]launcherAssetEntry
	namespaceBytes map[string]int64
	totalBytes     int64
	next           int64
	now            func() time.Time
	limits         launcherAssetLimits
}

type launcherConfigIconMetrics struct {
	decodedBytes int
	pixels       int
}

type launcherConfigIconStore struct {
	path          string
	readLibrary   func(string) ([]uint8, error)
	writeLibrary  func(string, launcherConfigIconLibrary) error
	loaded        bool
	cachedLoadErr error
	cached        launcherConfigIconLibrary
	cachedBudget  launcherConfigIconStoreBudget
	mu            sync.Mutex
}

type launcherConfigIconStoreBudget struct {
	count        int
	decodedBytes int
	pixels       int
}

type launcherConfigJSONFrame struct {
	kind  int32
	count int
}

type launcherConfigStore struct {
	path                 string
	writeConfig          func(string, LauncherConfig) error
	runtimeReady         bool
	iconFingerprint      [32]uint8
	iconFingerprintReady bool
	mu                   sync.Mutex
}

type launcherGlobalHotkeyCommand struct {
	bindings     launcherHotkeyBindings
	closeManager bool
	result       chan error
}

type launcherGlobalHotkeyManager interface {
	Close() error
	Update(launcherHotkeyBindings) error
}

type launcherGlobalHotkeyRegistration struct {
	action     string
	binding    string
	id         int32
	modifiers  uint32
	virtualKey uint32
}

type launcherUpdateLogger struct {
	file *os.File
}

type launcherUpdateProgressWriter struct {
	ctx     context.Context
	written int64
	write   func(int64)
}

type launcherWidgetStore struct {
	path          string
	mu            sync.Mutex
	loaded        bool
	cachedExists  bool
	cached        DesktopWidgetDocument
	cachedLoadErr error
	writeDocument func(string, DesktopWidgetDocument) error
}

type launcherWindowController interface {
	Close()
	EmitEvent(string, []interface{}) bool
	Focus()
	Hide() application.Window
	IsFocused() bool
	IsMinimised() bool
	IsVisible() bool
	Position() (int, int)
	Restore()
	SetContentProtection(bool) application.Window
	SetPosition(int, int)
	Show() application.Window
}

type launcherWindowLayoutController interface {
	Close()
	EmitEvent(string, []interface{}) bool
	GetScreen() (*application.Screen, error)
	Hide() application.Window
	IsFullscreen() bool
	IsMaximised() bool
	Maximise() application.Window
	Minimise() application.Window
	RelativePosition() (int, int)
	Restore()
	SetMaxSize(int, int) application.Window
	SetMinSize(int, int) application.Window
	SetRelativePosition(int, int) application.Window
	SetSize(int, int) application.Window
	Size() (int, int)
}

type launcherWindowScreenResolver interface {
	GetScreen() (*application.Screen, error)
}

type launcherWindowSizer interface {
	GetScreen() (*application.Screen, error)
	IsFullscreen() bool
	IsMaximised() bool
	Maximise() application.Window
	RelativePosition() (int, int)
	Restore()
	SetMaxSize(int, int) application.Window
	SetMinSize(int, int) application.Window
	SetRelativePosition(int, int) application.Window
	SetSize(int, int) application.Window
	Size() (int, int)
}

type launcherWindowTrayController interface {
	Close()
	EmitEvent(string, []interface{}) bool
	Hide() application.Window
}

type limitedFileSearchCandidateCollector struct {
	limit    int
	items    []scoredFileSearchCandidate
	heap     fileSearchCandidateHeap
	resolver *fileSearchCandidatePathResolver
}

type limitedSortedFileSearchCandidateCollector struct {
	limit    int
	items    []scoredFileSearchCandidate
	heap     fileSearchSortedCandidateHeap
	resolver *fileSearchCandidatePathResolver
	comparer *fileSearchSortComparer
}

type loadJob struct {
	root string
}

type loadJob struct {
	root string
}

type loadResult struct {
	root  string
	idx   *VolumeIndex
	meta  fileSearchVolumeMeta
	err   error
	exist bool
}

type loadResult struct {
	root string
	idx  *VolumeIndex
	meta fileSearchVolumeMeta
	err  error
}

type memoryReleaseService struct {
	lock              sync.Mutex
	config            MemoryReleaseConfig
	moduleEnabled     bool
	state             MemoryReleaseState
	timer             *time.Timer
	scheduleID        uint64
	generation        uint64
	shuttingDown      bool
	runningDone       chan struct{}
	execute           func(string) error
	supported         func() bool
	processIsElevated func() bool
	now               func() time.Time
}

type mouseGestureButtonReplay struct {
	service    *mouseGestureService
	generation uint64
	button     string
	sendClick  func(string) error
}

type mouseGestureEventQueue struct {
	lock  sync.Mutex
	items []mouseGestureHookEvent
	head  int
	size  int
}

type mouseGestureHookEvent struct {
	message    uint32
	x          int
	y          int
	button     string
	modifier   string
	down       bool
	up         bool
	move       bool
	wheel      bool
	injected   bool
	suppressed bool
}

type mouseGestureOverlayLabelSnapshot struct {
	label string
	image *image.RGBA
}

type mouseGestureOverlayPayload struct {
	points      []MouseGesturePoint
	button      string
	label       string
	labelDirty  bool
	bounds      image.Rectangle
	labelBounds image.Rectangle
	alpha       uint8
	scale       float64
}

type mouseGestureOverlayWindow struct {
	ready          chan error
	done           chan struct{}
	lock           sync.Mutex
	closeOnce      sync.Once
	hwnd           uintptr
	threadID       uintptr
	payload        mouseGestureOverlayPayload
	pendingPayload *mouseGestureOverlayPayload
	updatePending  bool
	redrawing      bool
	alpha          uint8
	visible        bool
	labelSnapshot  mouseGestureOverlayLabelSnapshot
}

type mouseGesturePlatformRuntime struct {
	service           *mouseGestureService
	events            *mouseGestureEventQueue
	ready             chan error
	config            MouseGestureConfig
	moduleEnabled     bool
	gestureEnabled    bool
	configReady       bool
	threadID          uint32
	mouseHook         uintptr
	hookRef           uintptr
	generation        uint64
	wake              func() error
	timeoutEpoch      atomic.Uint64
	timeoutPending    atomic.Bool
	queueFailure      atomic.Pointer[mouseGestureHookEvent]
	stopping          atomic.Bool
	targetScope       atomic.Pointer[mouseGestureTargetScope]
	targetHWNDAtPoint func(int, int, bool) uintptr
	buttonReplays     chan mouseGestureButtonReplay
	buttonReplayStop  chan struct{}
	buttonReplayDone  chan struct{}
}

type mouseGestureRuntimeSession struct {
	state             string
	button            string
	start             MouseGesturePoint
	points            []MouseGesturePoint
	startedAt         time.Time
	lastMovedAt       time.Time
	target            MouseGestureTarget
	targetResolved    bool
	suppressed        bool
	replayOnRelease   bool
	currentLabel      string
	currentCorner     string
	triggeredCorner   string
	candidateCorner   string
	candidateSince    time.Time
	lastTriggeredAt   map[string]time.Time
	screens           []MouseGestureScreenBounds
	lastScreenRefresh time.Time
	lastOverlayUpdate time.Time
	lastObservedHWND  uintptr
	overlay           *mouseGestureOverlayWindow
	labelOverlay      *mouseGestureOverlayWindow
	runtimeHandle     *mouseGesturePlatformRuntime
	generation        uint64
	timeoutTimer      *time.Timer
	timeoutDeadline   time.Time
	now               func() time.Time
	sendButtonDown    func(string) error
	sendButtonUp      func(string) error
	sendButtonClick   func(string) error
	targetHWNDAtPoint func(int, int, bool) uintptr
	targetFromHWND    func(uintptr) MouseGestureTarget
}

type mouseGestureService struct {
	lock                   sync.Mutex
	runtimeLifecycle       sync.Mutex
	config                 MouseGestureConfig
	moduleEnabled          bool
	runtimeStop            chan struct{}
	runtimeDone            chan struct{}
	runtimeMarkStopping    func()
	runtimeGeneration      uint64
	runtimeActive          bool
	runtimeActions         sync.WaitGroup
	shuttingDown           bool
	lastError              string
	lastAction             string
	lastGesture            string
	suppressUpMask         atomic.Uint32
	gestureInProgress      atomic.Bool
	capturePaused          bool
	observeGesture         atomic.Bool
	observeHotCorners      atomic.Bool
	observeGestureTargets  atomic.Bool
	launcherActionExecutor func(GestureAction) error
	windowMoveExecutor     func(GestureAction, MouseGestureActionTarget) error
	platformRuntimeSync    func()
}

type mouseGestureTargetScope struct {
	hwnd    uintptr
	allowed bool
}

type namePrefixBucketCounts struct {
	firstByte [256]int32
	twoByte   []int32
}

type nativeFileDragDataObject struct {
	combridge.IUnknownImpl
	paths                     []string
	preferredDropEffectFormat uint16
	formats                   []w32.FORMATETC
}

type nativeFileDragDataObjectCom interface {
	DAdvise(*w32.FORMATETC, uint32, *w32.IAdviseSink, *uint32) uintptr
	DUnadvise(uint32) uintptr
	EnumDAdvise(**w32.IEnumStatData) uintptr
	EnumFormatEtc(uint32, **w32.IEnumFORMATETC) uintptr
	GetCanonicalFormatEtc(*w32.FORMATETC, *w32.FORMATETC) uintptr
	GetData(*w32.FORMATETC, *w32.STGMEDIUM) uintptr
	GetDataHere(*w32.FORMATETC, *w32.STGMEDIUM) uintptr
	QueryGetData(*w32.FORMATETC) uintptr
	SetData(*w32.FORMATETC, *w32.STGMEDIUM, int32) uintptr
}

type nativeFileDragEnumFormatEtcCom interface {
	Clone(**w32.IEnumFORMATETC) uintptr
	Next(uint32, *w32.FORMATETC, *uint32) uintptr
	Reset() uintptr
	Skip(uint32) uintptr
}

type nativeFileDragFormatEnumerator struct {
	combridge.IUnknownImpl
	formats []w32.FORMATETC
	index   int
}

type nodePathCache struct {
	sliceCache []string
	deltaCache map[int32]string
}

type oledBlackoutAudioSessionScanKey struct {
	topologyGeneration uint64
	sessionID          string
	processID          uint32
	systemSounds       bool
}

type oledBlackoutBrowserMediaContinuity struct {
	lock    sync.Mutex
	entries map[oledBlackoutBrowserMediaContinuityKey]string
}

type oledBlackoutBrowserMediaContinuityKey struct {
	windowHandle uintptr
	processID    uint32
	processPath  string
}

type oledBlackoutCursorCommand struct {
	hide     bool
	show     bool
	close    bool
	response chan error
}

type oledBlackoutCursorController interface {
	Close() error
	Hide() error
	Show() error
}

type oledBlackoutHotkeyCommand struct {
	bindings     []oledBlackoutHotkeyBinding
	closeManager bool
	result       chan oledBlackoutHotkeyUpdateResult
}

type oledBlackoutHotkeyManager interface {
	Close() error
	Update([]oledBlackoutHotkeyBinding) oledBlackoutHotkeyUpdateResult
}

type oledBlackoutHotkeyRegistration struct {
	profileID  string
	binding    string
	id         int32
	modifiers  uint32
	virtualKey uint32
}

type oledBlackoutLastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

type oledBlackoutMonitorInfo struct {
	cbSize  uint32
	monitor oledBlackoutRect
	work    oledBlackoutRect
	flags   uint32
	device  [32]uint16
}

type oledBlackoutNativeOverlayWindow struct{
	overlay any
	ready chan error
	done chan struct {}
	lock sync.Mutex
	hwnd uintptr
	threadID uintptr
	bounds application.Rect
	visible bool
	closeOnce sync.Once
}

type oledBlackoutOverlayWindow struct{
	service any
	screenID string
	bounds application.Rect
	window application.Window
	nativeWindow *oledBlackoutNativeOverlayWindow
}

type oledBlackoutRect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type oledBlackoutService struct {
	lock                   sync.Mutex
	hotkeyOperationLock    sync.Mutex
	backgroundActivities   sync.WaitGroup
	browserMediaContinuity oledBlackoutBrowserMediaContinuity
	app                    *application.App
	config                 OLEDBlackoutConfig
	moduleEnabled          bool
	hotkeyCapture          bool
	hotkeyCaptureLegacyRef int
	hotkeyCaptureOwners    map[string]struct{}
	shuttingDown           bool
	lifecycleContext       context.Context
	lifecycleCancel        func()
	lifecycleGeneration    uint64
	shutdownDone           chan struct{}
	activeProfileID        string
	cursorHidden           bool
	cursorController       oledBlackoutCursorController
	lastError              string
	overlayWindows         map[string]*oledBlackoutOverlayWindow
	visibleOverlays        map[string]struct{}
	hotkeyManager          oledBlackoutHotkeyManager
	hotkeyRegistered       map[string]bool
	hotkeyErrors           map[string]string
	shouldSuppressHotkey   func(string) bool
	idleTimer              *time.Timer
	idleScheduleID         uint64
	inputPollTimer         *time.Timer
	inputPollScheduleID    uint64
	inputPollActive        bool
	focusRetryTimer        *time.Timer
	focusRetryScheduleID   uint64
	lastCursorX            int
	lastCursorY            int
	lastCursorValid        bool
	lastPressedKeys        map[uintptr]struct{}
	inputDismissGuardUntil time.Time
	autoActivatedProfileID string
	quickIdleProfileID     string
	quickIdleWakeAt        time.Time
	quickIdleActiveAt      time.Time
}

type oledBlackoutWindowProcessCandidate struct {
	hwnd         uintptr
	processID    uint32
	processPath  string
	processAUMID string
	windowTitle  string
}

type pendingPathRequest struct {
	id        string
	root      string
	nodeIndex int32
	frn       uint64
}

type persistTask struct {
	root string
	idx  *VolumeIndex
}

type platformDesktopWidgetAudioPlayer struct {
	playback sync.Mutex
}

type pluginManagedWindow struct {
	name        string
	logicalName string
	owner       string
	pluginID    string
	entry       string
	mode        string
	url         string
	title       string
	frameless   bool
	fullscreen  bool
	alwaysOnTop bool
	window      application.Window
}

type pluginWindowReservation struct {
	name       string
	owner      string
	pluginID   string
	generation uint64
}

type pluginWindowService struct {
	lock       sync.Mutex
	app        *application.App
	windows    map[string]*pluginManagedWindow
	creating   map[string]*pluginWindowReservation
	generation uint64
	shutting   bool
}

type preparedArchiveEntry struct {
	file       *zip.File
	relative   string
	targetPath string
	directory  bool
}

type qrCodeAnnotationStroke struct {
	id         int
	tool       string
	start      image.Point
	end        image.Point
	points     []image.Point
	color      color.RGBA
	lineWidth  int
	fontSize   int
	fontFamily string
	text       string
	textWidth  int
	number     int
}

type qrCodeCaptureAcceptedNotifier struct {
	callback func()
	once     sync.Once
}

type qrCodeNativeSelectionOptions struct {
	snapToControls        bool
	snapToWindows         bool
	confirmBeforeFinish   bool
	annotationToolbar     *screenshotSelectionToolbarWindowService
	annotationLineWidth   int
	cornerRadius          int
	cornerRadiusLabel     string
	captureCursor         bool
	fastGDIPreview        bool
	prewarmHDRCapture     bool
	freezeInitialFrame    bool
	waitForHDRPreview     bool
	cursorSnapshot        screenshotCursorSnapshot
	onAccepted            func()
	onLineWidthChanged    func(int)
	onCornerRadiusChanged func(int)
}

type qrCodeNativeSelectionResult struct {
	pngData           []uint8
	cancelled         bool
	needsFinalize     bool
	annotated         bool
	sourceProcessName string
	selection         image.Rectangle
	targetWindow      uintptr
	cursorSnapshot    screenshotCursorSnapshot
}

type qrCodeScreenSelectionSession struct {
	snapshot                    qrCodeScreenSnapshot
	originalBitmap              uintptr
	originalDC                  uintptr
	originalPrevious            uintptr
	shadedBitmap                uintptr
	shadedDC                    uintptr
	shadedPrevious              uintptr
	borderBrush                 uintptr
	cancelled                   bool
	resultPNG                   []uint8
	err                         error
	options                     qrCodeNativeSelectionOptions
	sourceProcessName           string
	selectedWindow              uintptr
	hitTestTransparent          bool
	rightButtonDownHandled      bool
	rightButtonSuppressContext  bool
	rightButtonCancelPending    bool
	dragging                    bool
	startPoint                  image.Point
	currentPoint                image.Point
	selection                   image.Rectangle
	selectionResizeHover        uint8
	selectionResizeActive       uint8
	selectionResizeOrigin       image.Rectangle
	selectionResizeRadiusOrigin int
	cornerRadius                int
	cornerRadiusHover           bool
	cornerRadiusDragging        bool
	cornerRadiusDragOrigin      int
	cornerRadiusMask            *image.Alpha
	cornerRadiusMaskRadius      int
	cornerRadiusInnerMask       *image.Alpha
	cornerRadiusInnerMaskRadius int
	controlSelection            image.Rectangle
	controlSelectionActive      bool
	hoverWindowRect             image.Rectangle
	hoverControlRect            image.Rectangle
	hoverControlWindow          uintptr
	hoverControlPoint           image.Point
	hoverControlAt              time.Time
	confirmBeforeFinish         bool
	editing                     bool
	annotationFinishing         bool
	annotationTool              string
	annotationColor             color.RGBA
	annotationColorPickerOpen   bool
	annotationLineWidth         int
	annotationFontSize          int
	annotationFontFamily        string
	annotationStrokes           []qrCodeAnnotationStroke
	annotationRedo              []qrCodeAnnotationStroke
	annotationContentTouched    bool
	annotationDraft             *qrCodeAnnotationStroke
	annotationTextInput         *qrCodeAnnotationStroke
	annotationText              strings.Builder
	annotationTextEditorPoint   image.Point
	annotationTextMoving        bool
	annotationTextSizing        bool
	annotationTextSelecting     bool
	annotationTextCaret         int
	annotationTextAnchor        int
	annotationTextMoveStart     image.Point
	annotationTextMoveSource    image.Point
	annotationTextAnchorStart   image.Point
	selectedStrokeID            int
	movingStrokeID              int
	moveStartPoint              image.Point
	moveSourceStroke            qrCodeAnnotationStroke
	annotationToolbar           *screenshotSelectionToolbarWindowService
	toolbarActions              chan screenshotSelectionToolbarAction
	toolbarTerminalActions      chan screenshotSelectionToolbarAction
	toolbarSessionID            int64
	nextStrokeID                int
	nextNumber                  int
	annotationEditing           atomic.Bool
	accepted                    *qrCodeCaptureAcceptedNotifier
	stateMutex                  sync.Mutex
	hwnd                        uintptr
	timeoutTriggered            bool
	upgradingPreview            bool
	pendingHDRPreview           *qrCodeScreenSnapshot
	finalizeRequest             *qrCodeNativeSelectionResult
}

type qrCodeScreenSnapshot struct {
	virtualBounds image.Rectangle
	original      *image.RGBA
	shaded        *image.RGBA
}

type readResult struct {
	content []uint8
	err     error
}

type remoteIconCacheEntry struct {
	path       string
	size       int64
	modifiedAt time.Time
}

type scoredFileSearchCandidate struct {
	Root           string
	Index          *VolumeIndex
	NodeIndex      int32
	Name           string
	FRN            uint64
	IsDirectory    bool
	ModifiedAtUnix int64
	Score          float64
	Path           string
	pathResolved   bool
}

type screenshotAccessible struct {
	lpVtbl *screenshotAccessibleVtbl
}

type screenshotCOMQueryEvent struct {
	requestID        uint64
	workerGeneration uint64
	started          bool
	result           screenshotCOMQueryResult
}

type screenshotCOMQueryPayload struct {
	point        gpuPreferenceScreenPoint
	targetWindow uintptr
}

type screenshotCOMQueryRequest struct {
	id               uint64
	priority         uint8
	payload          screenshotCOMQueryPayload
	busyError        error
	state            uint8
	workerGeneration uint64
	events           chan screenshotCOMQueryEvent
}

type screenshotCOMQueryResult struct {
	rect        image.Rectangle
	controlType uint32
	err         error
}

type screenshotCOMQueryWorker struct {
	generation uint64
	done       chan struct{}
	poisoned   bool
	request    *screenshotCOMQueryRequest
}

type screenshotCOMQueryWorkerPool struct {
	mu                  sync.Mutex
	name                string
	maxWorkers          int
	maxPendingFinal     int
	timeoutThreshold    int
	cooldown            int64
	pumpInterval        int64
	initializeThread    func() (func(), error)
	pumpMessages        func() bool
	execute             func(screenshotCOMQueryPayload) screenshotCOMQueryResult
	now                 func() time.Time
	wake                chan struct{}
	nextRequestID       uint64
	nextGeneration      uint64
	currentGeneration   uint64
	workers             map[uint64]*screenshotCOMQueryWorker
	pendingFinal        []*screenshotCOMQueryRequest
	pendingHover        *screenshotCOMQueryRequest
	consecutiveTimeouts int
	cooldownUntil       time.Time
	shuttingDown        bool
}

type screenshotCOMQueryWorkerPoolConfig struct {
	name               string
	maxWorkers         int
	maxPendingFinal    int
	timeoutThreshold   int
	cooldown           int64
	pumpInterval       int64
	initializeThread   func() (func(), error)
	pumpThreadMessages func() bool
	execute            func(screenshotCOMQueryPayload) screenshotCOMQueryResult
	now                func() time.Time
}

type screenshotCaptureAcceptedNotifier struct {
	service *BootstrapService
	once    sync.Once
}

type screenshotCursorSnapshot struct {
	info screenshotCursorInfo
	ok   bool
}

type screenshotD3D11Device struct {
	device       unsafe.Pointer
	context      unsafe.Pointer
	featureLevel uint32
}

type screenshotDXGIHDRScreenCaptureBackend struct {
	displays []screenshotDisplayCaptureInfo
	fallback screenshotGDIScreenCaptureBackend
}

type screenshotDXGIOutputCaptureEntry struct {
	mu             sync.Mutex
	key            string
	displayLabel   string
	device         *screenshotD3D11Device
	duplication    unsafe.Pointer
	toneMapOptions screenshotHDRToneMapOptions
	lastFrame      *image.RGBA
	lastFrameFree  func()
	lastUsed       time.Time
	closed         bool
}

type screenshotImageMemoryBudget struct {
	mu    sync.Mutex
	limit int64
	used  int64
}

type screenshotInput struct {
	inputType uint32
	mouse     screenshotMouseInput
}

type screenshotMouseInput struct {
	dx        int32
	dy        int32
	mouseData uint32
	flags     uint32
	time      uint32
	extraInfo uintptr
}

type screenshotNativePinWindow struct{
	service any
	pin any
	pinName string
	pinGeneration uint64
	image *image.RGBA
	imageRelease func()
	ready chan error
	lock sync.Mutex
	hwnd uintptr
	threadID uintptr
	bounds application.Rect
	visible bool
	opacity float64
	clickThrough bool
	hovered bool
	panelVisible bool
	closePressed bool
	opacityDrag bool
	highlightTick int
	closeOnce sync.Once
}

type screenshotNativePreviewWindow struct{
	service any
	ready chan error
	lock sync.Mutex
	hwnd uintptr
	threadID uintptr
	bounds application.Rect
	payload ScreenshotCaptureResult
	version int64
	image *image.RGBA
	opacity float64
	visible bool
	themeDark bool
	closeOnce sync.Once
}

type screenshotPNGIDATChunkWriter struct {
	writer io.Writer
	buffer []uint8
	limit  int
	closed bool
}

type screenshotPNGRowReleaser interface {
	ReleaseRow(int)
}

type screenshotPNGRowSource interface {
	Bounds() image.Rectangle
	Row(int) ([]uint8, error)
}

type screenshotPinWindowService struct {
	lock              sync.Mutex
	app               *application.App
	assets            *launcherAssetService
	root              string
	windows           map[string]*screenshotPinnedWindow
	snapshots         map[string]screenshotPinnedWindowSnapshot
	contentWindows    map[string]string
	lastWindowName    string
	shutting          bool
	skipPersist       bool
	restoreSuppressed bool
	shutdownPersisted bool
	nextOrder         int64
	nextGeneration    uint64
	restoreOnce       sync.Once
}

type screenshotPinnedWindow struct {
	lock         sync.RWMutex
	generation   uint64
	stateVersion uint64
	closed       bool
	contentKey   string
	name         string
	imageData    string
	imageURL     string
	thumbnailURL string
	sourcePath   string
	window       application.Window
	nativeWindow *screenshotNativePinWindow
	imageWidth   int
	imageHeight  int
	scale        float64
	opacity      float64
	clickThrough bool
	autoShow     bool
	order        int64
}

type screenshotPinnedWindowView struct {
	generation   uint64
	stateVersion uint64
	closed       bool
	contentKey   string
	name         string
	imageData    string
	imageURL     string
	thumbnailURL string
	sourcePath   string
	window       application.Window
	nativeWindow *screenshotNativePinWindow
	imageWidth   int
	imageHeight  int
	scale        float64
	opacity      float64
	clickThrough bool
	autoShow     bool
	order        int64
}

type screenshotPreviewWindowService struct {
	createMu     sync.Mutex
	lock         sync.Mutex
	app          *application.App
	window       application.Window
	nativeWindow *screenshotNativePreviewWindow
	payload      ScreenshotCaptureResult
	version      int64
	shutting     bool
	hidden       bool
	bounds       application.Rect
	ready        chan struct{}
	assets       *launcherAssetService
	capability   screenshotRouteCapability
}

type screenshotRGBAImageRowSource struct {
	source *image.RGBA
}

type screenshotScreenCaptureBackend interface {
	captureScreenRect(image.Rectangle, screenshotScreenCaptureOptions) (*image.RGBA, error)
	captureVirtualScreen(screenshotScreenCaptureOptions) (qrCodeScreenSnapshot, error)
	name() string
}

type screenshotScreenCaptureOptions struct {
	captureCursor         bool
	cursorSnapshot        screenshotCursorSnapshot
	requireFreshDXGIFrame bool
}

type screenshotScrollingAppendCandidate struct {
	topSkip int
	overlap int
	score   int64
}

type screenshotScrollingCanvasChunk struct {
	image   *image.RGBA
	release func()
}

type screenshotScrollingCaptureOutput struct {
	pngData          []uint8
	cancelled        bool
	truncated        bool
	truncationReason string
}

type screenshotScrollingChunkedCanvas struct {
	width  int
	height int
	chunks []*screenshotScrollingCanvasChunk
}

type screenshotSelectionOverlayHitTestSession interface {
	setHitTestTransparent(bool)
}

type screenshotSelectionToolbarWindowService struct {
	lifecycle                  sync.Mutex
	lock                       sync.Mutex
	app                        *application.App
	window                     application.Window
	sessionID                  int64
	state                      screenshotSelectionToolbarState
	messages                   screenshotSelectionToolbarMessages
	actions                    chan screenshotSelectionToolbarAction
	terminalActions            chan screenshotSelectionToolbarAction
	notify                     func(bool) bool
	version                    int64
	capability                 screenshotRouteCapability
	removeConfirmEventListener func()
	removeCancelEventListener  func()
	shutting                   bool
	lastKeepVisibleAt          time.Time
}

type screenshotUIAutomation struct {
	lpVtbl *screenshotUIAutomationVtbl
}

type screenshotWindowCaptureAcceptedNotifier struct {
	callback func()
	once     sync.Once
}

type screenshotWindowSelectionSession struct {
	snapshot           qrCodeScreenSnapshot
	shadedBitmap       uintptr
	shadedDC           uintptr
	shadedPrevious     uintptr
	borderBrush        uintptr
	cancelled          bool
	resultPNG          []uint8
	err                error
	sourceProcessName  string
	selectedProcess    WindowProcessPickResult
	captureControls    bool
	processPickOnly    bool
	excludedPID        uint32
	accepted           *screenshotWindowCaptureAcceptedNotifier
	hitTestTransparent bool
	rightCancelPending bool
	pendingCaptureRect image.Rectangle
	hoverRect          image.Rectangle
	hoverControlRect   image.Rectangle
	hoverControlPoint  image.Point
	hoverControlAt     time.Time
	hoverProcess       WindowProcessPickResult
	hoverProcessHwnd   uintptr
	stateMutex         sync.Mutex
	hwnd               uintptr
	timeoutTriggered   bool
}

type searchTermValue struct {
	value          string
	fromPathField  bool
	allowNameMatch bool
}

type shellExecuteInfo struct {
	cbSize       uint32
	fMask        uint32
	hwnd         uintptr
	lpVerb       *uint16
	lpFile       *uint16
	lpParameters *uint16
	lpDirectory  *uint16
	nShow        int32
	hInstApp     uintptr
	lpIDList     uintptr
	lpClass      *uint16
	hkeyClass    uintptr
	dwHotKey     uint32
	hIcon        uintptr
	hProcess     uintptr
}

type sqliteBackuper interface {
	NewBackup(string) (*sqlite.Backup, error)
}

type stagedIcon struct {
	slot        launcherConfigIconSlot
	ref         string
	contentType string
	payload     string
	dataURL     string
}

type stagedWorkspaceData struct {
	sourcePath           string
	targetPath           string
	stagingPath          string
	manifest             map[string]workspaceDataManifestEntry
	sourceInspection     workspacePathInspection
	targetInspection     workspacePathInspection
	targetParentIdentity workspacePathIdentity
	stagingIdentity      workspacePathIdentity
	committedIdentity    workspacePathIdentity
	committed            bool
	targetWasEmpty       bool
}

type twoFactorService struct {
	configPath        string
	lock              sync.Mutex
	writeLock         sync.Mutex
	timeSyncLock      sync.Mutex
	fetchTime         func() (twoFactorTimeCache, error)
	sessionKey        []uint8
	sessionGeneration uint64
	timeCache         twoFactorTimeCache
}

type versionTranslation struct {
	language uint16
	codePage uint16
}

type volumeIndexCheckpointHeader struct {
	rootFRN      uint64
	journalID    uint64
	lastUSN      int64
	generatedAt  time.Time
	entryCount   uint32
	sortedCount  uint32
	namePoolSize uint64
}

type volumeIndexCheckpointSections struct {
	nodes  []uint8
	sorted []uint8
	names  []uint8
}

type volumeIndexCheckpointView struct {
	root          string
	rootFRN       uint64
	journalID     uint64
	lastUSN       int64
	generatedAt   time.Time
	layout        volumeIndexCheckpointLayout
	mappedFile    *volumeIndexMappedFile
	nodes         []IndexNode
	nodeBytes     []uint8
	sortedIndices []int32
	sortedBytes   []uint8
	namePool      []uint8
	namePoolBytes []uint8
	activeCount   int
	prefixBuckets *namePrefixBucketCounts
}

type volumeIndexDeltaNode struct {
	node             IndexNode
	name             []uint8
	bigramSignature  uint64
	trigramSignature uint64
	pinyinFull       []uint8
	pinyinInitials   []uint8
	replaces         int32
}

type volumeIndexHeapReadProvider struct {
	idx *VolumeIndex
}

type volumeIndexJournalPendingEntry struct {
	parentFRN uint64
	name      string
	isDir     bool
	deleted   bool
}

type volumeIndexJournalPendingState struct {
	root          string
	rootFRN       uint64
	ignoreMatcher fileSearchIgnoreMatcher
	lease         volumeIndexReadLease
	view          volumeIndexReadView
	overlay       map[uint64]volumeIndexJournalPendingEntry
}

type volumeIndexMappedFile struct {
	file    *os.File
	mapping uintptr
	views   []uintptr
}

type volumeIndexMappedReadProvider struct {
	mu      sync.Mutex
	view    volumeIndexReadView
	path    string
	close   func() error
	release func()
	active  int
	closing bool
	closed  bool
}

type volumeIndexReadLease struct{
	idx any
	view volumeIndexReadView
	release func()
	released bool
}

type volumeIndexReadProvider interface {
	Acquire() volumeIndexReadLease
}

type volumeIndexReadView struct {
	root          string
	nodes         []IndexNode
	nodeBytes     []uint8
	sortedIndices []int32
	sortedBytes   []uint8
	namePool      []uint8
	activeCount   int
	prefixBuckets *namePrefixBucketCounts
	nameTrigram   *volumeNameTrigramIndex
	nameBigram    *volumeNameBigramIndex
	pinyin        *volumePinyinIndex
	tombstones    map[int32]struct{}
	deltaNodes    []volumeIndexDeltaNode
	deltaByFRN    map[uint64]int32
	deltaReplaced map[int32]struct{}
}

type volumeNameBigramIndex struct {
	version    uint64
	signatures []uint64
}

type volumeNameTrigramIndex struct {
	version        uint64
	signatures     []uint64
	signatureBytes []uint8
	mappedFile     *volumeIndexMappedFile
	closed         bool
}

type volumePinyinIdentity struct {
	rootFRN     uint64
	journalID   uint64
	lastUSN     int64
	generatedAt int64
	nodeCount   uint32
}

type volumePinyinIndex struct {
	identity    volumePinyinIdentity
	records     []uint8
	aliases     []uint8
	mappedFile  *volumeIndexMappedFile
	fileBytes   int64
	mappedBytes int64
	heapBytes   int64
}

type walEntry struct {
	endUSN int64
	raw    []uint8
}

type walkState struct {
	idx      int32
	path     string
	excluded bool
}

type windowManagementConfigAlias struct {
	CursorWrapHorizontalEnabled bool                               `json:"cursorWrapHorizontalEnabled,omitempty"`
	CursorWrapVerticalEnabled   bool                               `json:"cursorWrapVerticalEnabled,omitempty"`
	Target                      WindowManagementTarget     `json:"target,omitempty"`
	OpacityPercent              int                                `json:"opacityPercent"`
	Resolution                  WindowManagementSize       `json:"resolution,omitempty"`
	BorderlessSnapshots         []WindowFullscreenSnapshot `json:"borderlessSnapshots,omitempty"`
	FullscreenSnapshots         []WindowFullscreenSnapshot `json:"fullscreenSnapshots,omitempty"`
	opacityPercentSet           bool
}

type windowManagementService struct {
	operationLock       sync.Mutex
	lock                sync.Mutex
	config              WindowManagementConfig
	moduleEnabled       bool
	shuttingDown        bool
	persistConfig       func(WindowManagementConfig)
	cursorStop          chan struct{}
	cursorDone          chan struct{}
	cursorActive        bool
	cursorGuardPx       int
	opacitySnapshots    map[uintptr]windowManagementOpacitySnapshot
	managedSnapshots    map[uintptr]WindowFullscreenSnapshot
	borderlessSnapshots []WindowFullscreenSnapshot
	fullscreenSnapshots []WindowFullscreenSnapshot
	lastError           string
}

type windowsLauncherGlobalHotkeyManager struct {
	callback             func(string)
	commands             chan launcherGlobalHotkeyCommand
	ready                chan struct{}
	threadID             uint32
	keyboardHook         uintptr
	keyboardHookCallback uintptr
	keyboardHookBindings []launcherGlobalHotkeyRegistration
	printScreenKeyDown   bool
	closeOnce            sync.Once
}

type windowsOLEDBlackoutCursorController struct {
	commands  chan oledBlackoutCursorCommand
	ready     chan struct{}
	done      chan struct{}
	call      func(bool) (int32, error)
	lock      sync.Mutex
	closed    bool
	closeErr  error
	closeOnce sync.Once
}

type windowsOLEDBlackoutHotkeyManager struct {
	callback  func(oledBlackoutHotkeyEvent)
	commands  chan oledBlackoutHotkeyCommand
	ready     chan struct{}
	done      chan struct{}
	threadID  uint32
	closeOnce sync.Once
}

type workspaceDataMaintenanceGate struct {
	mu          sync.Mutex
	maintenance bool
	active      int
	idle        chan struct{}
}

type workspacePathIdentity struct {
	canonical string
	volume    uint64
	file      uint64
	valid     bool
}

type workspacePathInspection struct {
	canonical        string
	exists           bool
	identity         workspacePathIdentity
	ancestorIdentity workspacePathIdentity
}
