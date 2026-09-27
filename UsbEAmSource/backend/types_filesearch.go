// AUTO-RECONSTRUCTED TYPES — DOMAIN: filesearch
// 研究用途
package main

import (
	"context"
	"golang.org/x/text/collate"
	"os"
	"regexp"
	"sync"
	"time"
)

type FileSearchConfig struct {
	Enabled             bool                   `json:"enabled"`
	PinyinSearchEnabled bool                   `json:"pinyinSearchEnabled,omitempty"`
	RecentItemsEnabled  *bool                  `json:"recentItemsEnabled,omitempty"`
	RecentItemsLimit    int                    `json:"recentItemsLimit,omitempty"`
	RecentItems         []FileSearchRecentItem `json:"recentItems,omitempty"`
	Volumes             []string               `json:"volumes"`
	MaxResults          int                    `json:"maxResults"`
	IgnoredDirectories  []string               `json:"ignoredDirectories,omitempty"`
	ResourceMode        string                 `json:"resourceMode,omitempty"`
	FileTypeFilters     []FileSearchTypeFilter `json:"fileTypeFilters"`
	LegacyMode          string                 `json:"mode,omitempty"`
	LegacyRoots         []string               `json:"roots,omitempty"`
	LegacyRules         []string               `json:"rules,omitempty"`
}

// fileSearchConfigNormalized 是 normalizeFileSearchConfig 的归一化输出（10 字段，无 Legacy*）。
// 汇编 0x140879e00 实证输出结构体偏移：Enabled 0x00 / PinyinSearchEnabled 0x01 /
// RecentItemsEnabled 0x08 / RecentItemsLimit 0x10 / RecentItems 0x18 / Volumes 0x30 /
// MaxResults 0x48 / IgnoredDirectories 0x50 / ResourceMode 0x68 / FileTypeFilters 0x78。
type fileSearchConfigNormalized struct {
	Enabled             bool
	PinyinSearchEnabled bool
	RecentItemsEnabled  *bool
	RecentItemsLimit    int
	RecentItems         []FileSearchRecentItem
	Volumes             []string
	MaxResults          int
	IgnoredDirectories  []string
	ResourceMode        string
	FileTypeFilters     []FileSearchTypeFilter
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
	TotalMatchCount      int             `json:"totalMatchCount"`
	SnapshotStale        bool            `json:"snapshotStale"`
	Syncing              bool            `json:"syncing"`
	RuntimeWarming       bool            `json:"runtimeWarming"`
	PinyinRefreshPending bool            `json:"pinyinRefreshPending"`
}

type FileSearchState struct {
	Config                 FileSearchConfig   `json:"config"`
	Volumes                []FileSearchVolume `json:"volumes"`
	Provider               string             `json:"provider"`
	Ready                  bool               `json:"ready"`
	Watching               bool               `json:"watching"`
	RuntimeIndexCount      int                `json:"runtimeIndexCount"`
	RuntimeIndexEntries    int                `json:"runtimeIndexEntries"`
	WatcherCount           int                `json:"watcherCount"`
	USNFollowerCount       int                `json:"usnFollowerCount"`
	USNFollowerBufferBytes int64              `json:"usnFollowerBufferBytes"`
	GoHeapAllocBytes       uint64             `json:"goHeapAllocBytes"`
	GoHeapSysBytes         uint64             `json:"goHeapSysBytes"`
	GoHeapReleasedBytes    uint64             `json:"goHeapReleasedBytes"`
	GoHeapRetainedBytes    uint64             `json:"goHeapRetainedBytes"`
	ProcessWorkingSetBytes uint64             `json:"processWorkingSetBytes"`
	ProcessPrivateBytes    uint64             `json:"processPrivateBytes"`
	IndexedCount           int                `json:"indexedCount"`
	LastIndexedAt          string             `json:"lastIndexedAt"`
	LastJournalReadAt      string             `json:"lastJournalReadAt"`
	LastJournalApplyAt     string             `json:"lastJournalApplyAt"`
	JournalNextUSN         int64              `json:"journalNextUsn"`
	JournalLastUSN         int64              `json:"journalLastUsn"`
	JournalLag             int64              `json:"journalLag"`
	WatcherRestarting      bool               `json:"watcherRestarting"`
	LastUsedAt             string             `json:"lastUsedAt"`
	LastError              string             `json:"lastError"`
	IndexPath              string             `json:"indexPath"`
	RuntimeState           string             `json:"runtimeState"`
	IdleUnloadAfter        int                `json:"idleUnloadAfter"`
	Syncing                bool               `json:"syncing"`
	Lagging                bool               `json:"lagging"`
	PendingBatches         int                `json:"pendingBatches"`
	PendingBytes           int64              `json:"pendingBytes"`
	PendingAgeSeconds      int                `json:"pendingAgeSeconds"`
	StaticMmap             bool               `json:"staticMmap"`
	StaticMmapActive       bool               `json:"staticMmapActive"`
	OverlayEntries         int                `json:"overlayEntries"`
	OverlayRatio           float64            `json:"overlayRatio"`
	MergePending           int                `json:"mergePending"`
	MergeRunning           bool               `json:"mergeRunning"`
	MergeRetries           int                `json:"mergeRetries"`
	MergeError             string             `json:"mergeError"`
	CheckpointPending      int                `json:"checkpointPending"`
	CheckpointRunning      bool               `json:"checkpointRunning"`
	CheckpointRetries      int                `json:"checkpointRetries"`
	CheckpointError        string             `json:"checkpointError"`
	WarmingVolumes         int                `json:"warmingVolumes"`
	WarmedVolumes          int                `json:"warmedVolumes"`
	WarmingTarget          string             `json:"warmingTarget"`
	LastReleaseReason      string             `json:"lastReleaseReason"`
	RestaticPending        int                `json:"restaticPending"`
	RestaticRunning        bool               `json:"restaticRunning"`
	RestaticRetries        int                `json:"restaticRetries"`
	RestaticError          string             `json:"restaticError"`
	PinyinIndexWarming     bool               `json:"pinyinIndexWarming"`
	PinyinReadyVolumes     int                `json:"pinyinReadyVolumes"`
	PinyinTargetVolumes    int                `json:"pinyinTargetVolumes"`
	PinyinRecordCount      int                `json:"pinyinRecordCount"`
	PinyinIndexFileBytes   int64              `json:"pinyinIndexFileBytes"`
	PinyinMappedBytes      int64              `json:"pinyinMappedBytes"`
	PinyinHeapBytes        int64              `json:"pinyinHeapBytes"`
	PinyinIndexError       string             `json:"pinyinIndexError"`
	RealtimeEnabled        bool               `json:"realtimeEnabled"`
	RequiresAdmin          bool               `json:"requiresAdmin"`
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
	TotalMatchCount      int             `json:"totalMatchCount"`
	Offset               int             `json:"offset"`
	Limit                int             `json:"limit"`
	NextOffset           int             `json:"nextOffset"`
	HasMore              bool            `json:"hasMore"`
	SnapshotStale        bool            `json:"snapshotStale"`
	Syncing              bool            `json:"syncing"`
	RuntimeWarming       bool            `json:"runtimeWarming"`
	PinyinRefreshPending bool            `json:"pinyinRefreshPending"`
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

type rawIndexEntry struct {
	FRN       uint64
	ParentFRN uint64
	Name      string
	ModTime   uint32
	IsDir     bool
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
	Version   uint32                     `json:"version"`
	JournalID uint64                     `json:"journalId"`
	StartUSN  int64                      `json:"startUsn"`
	EndUSN    int64                      `json:"endUsn"`
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

type namePrefixBucketCounts struct {
	firstByte [256]int32
	twoByte   []int32
}

type nodePathCache struct {
	sliceCache []string
	deltaCache map[int32]string
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

type volumeIndexReadLease struct {
	idx      any
	view     volumeIndexReadView
	release  func()
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

// volumeIndexTombstoneLookup 墓碑后代判定上下文（0xe8 字节）。
// 布局经 asm 0x1407f1a60 实证：嵌入 volumeIndexReadView（0x00..0xe0），
// +0xe0 memo map[int32]bool 记忆化缓存。
type volumeIndexTombstoneLookup struct {
	volumeIndexReadView
	memo map[int32]bool
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
