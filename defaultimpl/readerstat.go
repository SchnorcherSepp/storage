package impl

import (
	"fmt"
	"github.com/SchnorcherSepp/storage/interfaces"
	"io"
	"log"
	"strings"
	"sync/atomic"
)

// DebugOff deactivates all debug messages. Errors, warnings or information are still printed.
const DebugOff = 0

// DebugLow shows debug messages that happen very rarely during operation (to keep the log files small).
const DebugLow = 1

// DebugHigh shows all debug messages.
const DebugHigh = 2

//--------------------------------------------------------------------------------------------------------------------//

type _ReaderStat struct {
	debugLvl    uint8  // enable debug logging [0, 1, 2] (level: high=2)
	packageName string // text for debug logging

	_CacheHit      uint64
	_CacheMis      uint64
	_CacheSet      uint64
	_RAtNew        uint64
	_RAtClosing    uint64
	_RAtClose      uint64
	_RAtReq        uint64
	_RAtRetErr     uint64
	_RAtSectorSkip uint64
	_RAtSectorRet  uint64
	_RAtBest       uint64
	_RAtAdd        uint64
	_RAtAddErr     uint64
}

func (s *_ReaderStat) Stat() map[string]uint64 {
	ret := map[string]uint64{
		"CacheHit":      atomic.LoadUint64(&s._CacheHit),
		"CacheMis":      atomic.LoadUint64(&s._CacheMis),
		"CacheSet":      atomic.LoadUint64(&s._CacheSet),
		"RAtNew":        atomic.LoadUint64(&s._RAtNew),
		"RAtClosing":    atomic.LoadUint64(&s._RAtClosing),
		"RAtClose":      atomic.LoadUint64(&s._RAtClose),
		"RAtReq":        atomic.LoadUint64(&s._RAtReq),
		"RAtRetErr":     atomic.LoadUint64(&s._RAtRetErr),
		"RAtSectorSkip": atomic.LoadUint64(&s._RAtSectorSkip),
		"RAtSectorRet":  atomic.LoadUint64(&s._RAtSectorRet),
		"RAtBest":       atomic.LoadUint64(&s._RAtBest),
		"RAtAdd":        atomic.LoadUint64(&s._RAtAdd),
		"RAtAddErr":     atomic.LoadUint64(&s._RAtAddErr),
	}

	// ignore zero values
	for k, v := range ret {
		if v == 0 {
			delete(ret, k)
		}
	}
	return ret
}

func (s *_ReaderStat) PrintStatAfterClose(fileId string) {
	// final call in .Close()

	first := true
	var sb strings.Builder
	for k, v := range s.Stat() {
		if !first {
			sb.WriteString(", ")
		} else {
			first = false
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%d", v))
	}

	if s.debugLvl >= DebugLow { // Debug level: low=1
		log.Printf("DEBUG: %s/stat.PrintStatAfterClose: fileId=%s: %s", s.packageName, fileId, sb.String())
	}
}

// ------------------------------------------------------------------------------------------------------------------ //

func (s *_ReaderStat) CacheGet(fileId string, sector uint64, reqLen, retLen int, err error) {
	if err == nil {
		atomic.AddUint64(&s._CacheHit, 1)
	} else {
		atomic.AddUint64(&s._CacheMis, 1)
	}
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.CacheGet: id=%s, sector=%d, req=%d/%d, ret=%d/%d, err=%v", s.packageName, fileId, sector, reqLen, interf.SectorSize, retLen, interf.SectorSize, err)
	}
}

func (s *_ReaderStat) CacheSet(fileId string, sector uint64, data int, err error) {
	atomic.AddUint64(&s._CacheSet, 1)
	if s.debugLvl >= DebugHigh || err != nil {
		pre := "DEBUG" // Debug level: high=2
		if err != nil {
			pre = "ERROR" // Debug level: error=0
		}
		log.Printf("%s: %s/stat.CacheSet: id=%s, sector=%d, data=%d/%d, expire=%d, err=%v", pre, s.packageName, fileId, sector, data, interf.SectorSize, interf.CacheExpireSeconds, err)
	}
}

func (s *_ReaderStat) RAtNew(fileId string, cache bool) {
	atomic.AddUint64(&s._RAtNew, 1)
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtNew: id=%s, _Cache=%v", s.packageName, fileId, cache)
	}
}

func (s *_ReaderStat) RAtClosing(fileId string) {
	atomic.AddUint64(&s._RAtClosing, 1)
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtClosing: id=%s", s.packageName, fileId)
	}
}

func (s *_ReaderStat) RAtClose(fileId string, slot int, active bool) {
	atomic.AddUint64(&s._RAtClose, 1)
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtClose: id=%s, slot=%d, active=%v", s.packageName, fileId, slot, active)
	}
}

func (s *_ReaderStat) RAtReq(fileId string, off int64, req int, sector uint64, innerOff int) {
	atomic.AddUint64(&s._RAtReq, 1)
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtReq: id=%s, off=%d, req=%d, startSector=%d, innerOff=%d", s.packageName, fileId, off, req, sector, innerOff)
	}
}

func (s *_ReaderStat) RAtRet(fileId string, off int64, req int, ret int, err error) {
	if err != nil && err != io.EOF {
		atomic.AddUint64(&s._RAtRetErr, 1)
	}
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtRet: id=%s, off=%d, req=%d, ret=%d, err=%v", s.packageName, fileId, off, req, ret, err)
	}
}

func (s *_ReaderStat) RAtSectorSkip(fileId string, skip uint64, n int, err error) {
	atomic.AddUint64(&s._RAtSectorSkip, 1)
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtSectorSkip: id=%s, skipSector=%d, n=%d/%d, err=%v", s.packageName, fileId, skip, n, interf.SectorSize, err)
	}
}

func (s *_ReaderStat) RAtSectorRet(fileId string, sector uint64, n int, err error) {
	atomic.AddUint64(&s._RAtSectorRet, 1)
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtSectorRet: id=%s, sector=%d, n=%d/%d, err=%v", s.packageName, fileId, sector, n, interf.SectorSize, err)
	}
}

func (s *_ReaderStat) RAtBest(fileId string, index int, current uint64) {
	if index >= 0 {
		atomic.AddUint64(&s._RAtBest, 1)
	}
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtBest: id=%s, index=%d, current=%d", s.packageName, fileId, index, current)
	}
}

func (s *_ReaderStat) RAtAdd(fileId string, sector uint64, err error) {
	atomic.AddUint64(&s._RAtAdd, 1)
	if err != nil && err != io.EOF {
		atomic.AddUint64(&s._RAtAddErr, 1)
	}
	if s.debugLvl >= DebugHigh { // Debug level: high=2
		log.Printf("DEBUG: %s/stat.RAtAdd: id=%s, startSector=%d, err=%v", s.packageName, fileId, sector, err)
	}
}
