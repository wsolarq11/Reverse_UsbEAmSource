// AUTO-RECONSTRUCTED TYPES — DOMAIN: twofactor
// 研究用途
package main

import (
	"sync"
	"time"
)

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
	ImportedCount  int             `json:"importedCount"`
	DuplicateCount int             `json:"duplicateCount"`
	SkippedCount   int             `json:"skippedCount"`
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
	PasswordConfigured bool                  `json:"passwordConfigured"`
	PasswordBlank      bool                  `json:"passwordBlank"`
	Unlocked           bool                  `json:"unlocked"`
	EntryCount         int                   `json:"entryCount"`
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
