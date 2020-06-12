package interf

// SectorSize is the size of a sector. A sector is a part of a file.
// It is comparable to sectors of a block device.
// The SectorSize is also the buffer size for the download.
const SectorSize = 16384 // 16 kiB

// MaxSectorJump determines how far you can jump backwards in an open reader.
// An open reader for google drive does not allow random read access.
// To reach a more distant sector, you either have to read up to this point or open a new reader.
// Opening a new reader often takes longer than reading unnecessary data.
const MaxSectorJump = (50 * 1024 * 1024) / SectorSize // 3200 sectors (=50 MiB, ~1sec with 400 MBit/s)

// MaxReadersPerFile determines how many open readers can be kept for later use. This should reduce reader openings.
const MaxReadersPerFile = 6

// CacheExpireSeconds is the default value n. The cache stores data for max. n seconds.
const CacheExpireSeconds = 2 * 24 * 60 * 60 // 2 days

// MaxFileSize defines the maximum size in byte of the supported files.
const MaxFileSize = 100 * 1024 * 1024 * 1024 // 100 GiB
