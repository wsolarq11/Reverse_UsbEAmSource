// AUTO-RECONSTRUCTED TYPES — DOMAIN: desktopwidget
// 研究用途
package main

import (
	"context"
	"github.com/wailsapp/wails/v3/pkg/application"
	"sync"
	"time"
)

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
	Year      int                  `json:"year"`
	Month     int                  `json:"month"`
	ServerNow string               `json:"serverNow"`
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
	WidgetID    string                               `json:"widgetId,omitempty"`
	Location    DesktopWeatherLocation               `json:"location"`
	Current     DesktopWeatherCurrent                `json:"current"`
	Hourly      []DesktopWeatherHour                 `json:"hourly"`
	Daily       []DesktopWeatherDay                  `json:"daily"`
	AirQuality  DesktopAirQuality                    `json:"airQuality"`
	Attribution []string                             `json:"attribution"`
	Domains     map[string]DesktopWeatherDomainState `json:"domains"`
	FetchedAt   string                               `json:"fetchedAt"`
	ExpiresAt   string                               `json:"expiresAt"`
	StaleUntil  string                               `json:"staleUntil"`
}

type DesktopWidget struct {
	ID             string              `json:"id"`
	Type           string              `json:"type"`
	Title          string              `json:"title"`
	Config         DesktopWidgetConfig `json:"config"`
	Revision       int64               `json:"revision"`
	LifecycleState string              `json:"lifecycleState,omitempty"`
	CreatedAt      string              `json:"createdAt"`
	UpdatedAt      string              `json:"updatedAt"`
}

type DesktopWidgetConfig struct {
	WeatherLocation      *DesktopWeatherLocation    `json:"weatherLocation,omitempty"`
	WeatherProviderOrder []string                   `json:"weatherProviderOrder,omitempty"`
	Timezones            []DesktopWorldClockZone    `json:"timezones,omitempty"`
	DurationMs           int64                      `json:"durationMs,omitempty"`
	Reminder             *DesktopReminderDefinition `json:"reminder,omitempty"`
	HideTitleOnHome      bool                       `json:"hideTitleOnHome,omitempty"`
	HourFormat           string                     `json:"hourFormat,omitempty"`
}

type DesktopWidgetCreateInput struct {
	Type   string              `json:"type"`
	Title  string              `json:"title"`
	Config DesktopWidgetConfig `json:"config"`
}

type DesktopWidgetDocument struct {
	Version                int                                  `json:"version"`
	Revision               int64                                `json:"revision"`
	UpdatedAt              string                               `json:"updatedAt"`
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
	Status               string                               `json:"status"`
	ErrorCode            string                               `json:"errorCode,omitempty"`
	ErrorMessage         string                               `json:"errorMessage,omitempty"`
	Revision             int64                                `json:"revision"`
	ServerNow            string                               `json:"serverNow"`
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
	ID               string              `json:"id"`
	Title            string              `json:"title"`
	Config           DesktopWidgetConfig `json:"config"`
	ExpectedRevision int64               `json:"expectedRevision"`
}

type DesktopWorldClockSnapshot struct {
	ServerNow string                  `json:"serverNow"`
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

type desktopWidgetScheduler struct {
	service  any
	wake     chan struct{}
	stop     chan struct{}
	done     chan struct{}
	once     sync.Once
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
