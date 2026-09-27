// AUTO-RECONSTRUCTED TYPES — DOMAIN: gpu
// 研究用途
package main

type GPUPreferenceAdapterInfo struct {
	PowerSavingName     string `json:"powerSavingName"`
	HighPerformanceName string `json:"highPerformanceName"`
}

type GPUPreferenceEntry struct {
	Path            string                 `json:"path"`
	Name            string                 `json:"name"`
	Icon            string                 `json:"icon"`
	IconData        string                 `json:"iconData"`
	PreferenceMode  string                 `json:"preferenceMode"`
	PreferenceValue int                    `json:"preferenceValue"`
	Settings        []GPUPreferenceSetting `json:"settings"`
	RawValue        string                 `json:"rawValue"`
}

type GPUPreferenceSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GPUPreferenceState struct {
	Entries     []GPUPreferenceEntry     `json:"entries"`
	AdapterInfo GPUPreferenceAdapterInfo `json:"adapterInfo"`
}

type gpuPreferenceScreenPoint struct {
	X int32
	Y int32
}
