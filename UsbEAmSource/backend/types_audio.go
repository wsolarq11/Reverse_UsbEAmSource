// AUTO-RECONSTRUCTED TYPES — DOMAIN: audio
// 研究用途
package main

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
	DefaultOutputDeviceID            string         `json:"defaultOutputDeviceId"`
	DefaultInputDeviceID             string         `json:"defaultInputDeviceId"`
	DefaultOutputDeviceVolumePercent int            `json:"defaultOutputDeviceVolumePercent"`
	DefaultInputDeviceVolumePercent  int            `json:"defaultInputDeviceVolumePercent"`
	DefaultOutputDeviceMuted         bool           `json:"defaultOutputDeviceMuted"`
	DefaultInputDeviceMuted          bool           `json:"defaultInputDeviceMuted"`
	Sessions                         []AudioSession `json:"sessions"`
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
