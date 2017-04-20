package ioservice

/*
 * Written by kimi kan, 2016-10
 * This file is used for manager the request. and combine the request to same device to a channel.
 * it finally make a pair (every dev <=> per go routine)
 */

import (
	"Kharites/network"
	"Kharites/util"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"

	"github.com/coocood/freecache"
)

type ioContext struct {
	writer *network.NetWriter
	msg    interface{}
}

//IOService ...
type IOService struct {
	c chan *ioContext

	//buf []byte //cache for
}

//Manager ...
type Manager struct {
	ios   map[ /*DiskID*/ uint32]*IOService
	mutex sync.Mutex
	wg    sync.WaitGroup

	cc    chan struct{}
	cache *freecache.Cache
}

//ReadBlock ...
func ReadBlock(diskID uint32, blockid int64) []byte {
	path, ok := util.DiskID2Path(diskID)
	if !ok {
		return nil
	}

	var count int64 = 64 * 1024
	buf := make([]byte, count)

	len := ReadBytesFromFile(path, buf, int64(blockid*count))
	if len == int(count) && len > 0 {
		return buf
	}
	return nil
}

var sum, got int

//ReadData ...
func (m *Manager) ReadData(diskID uint32, sectorCount uint16, sectorOffset uint64) []byte {
	blockid := sectorOffset / 128
	offsetInBlock := sectorOffset % 128
	if offsetInBlock+uint64(sectorCount) > 128 {
		fmt.Println("imput parameter overflow 64k")
		return nil
	}

	key := uint64(diskID) | (blockid)<<32
	const BytesInInt64 = 8
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  BytesInInt64, Cap: BytesInInt64,
	}

	hash := *(*[]byte)(unsafe.Pointer(&hdr))
	value, err := m.cache.Get(hash)
	sum++
	if err != nil {
		value = ReadBlock(diskID, int64(blockid))
		got++
		if value == nil {
			return nil
		}
		m.cache.Set(hash, value, 0)
	}

	return value[offsetInBlock*512 : (offsetInBlock+uint64(sectorCount))*512]
}

//NewFIOManager ...
func NewFIOManager() *Manager {
	m := &Manager{
		ios:   make(map[uint32]*IOService),
		cc:    make(chan struct{}),
		cache: freecache.NewCache(util.GetConfig().CacheSize),
	}

	go func() {
		for {
			time.Sleep(time.Second * 20)
			fmt.Println(sum, got)
			fmt.Printf("[Pre:] hit rate is %v, evacuates %v, entries %v, average time %v, expire count %v\n",
				m.cache.HitRate(), m.cache.EvacuateCount(),
				m.cache.EntryCount(), m.cache.AverageAccessTime(),
				m.cache.ExpiredCount())
			m.cache.ResetStatistics()
			fmt.Printf("[Post:] hit rate is %v, evacuates %v, entries %v, average time %v, expire count %v\n",
				m.cache.HitRate(), m.cache.EvacuateCount(),
				m.cache.EntryCount(), m.cache.AverageAccessTime(),
				m.cache.ExpiredCount())
		}

	}()

	debug.SetGCPercent(10)
	return m
}

//AddIORequest ...
func (m *Manager) AddIORequest(diskID uint32, msg2 *network.ReadDiskMsg, w *network.NetWriter) {
	ctx := &ioContext{
		writer: w,
		msg:    msg2,
	}
	srv := m.GetIOService(diskID)
	if srv == nil {
		srv = m.AddIOService(diskID)
	}
	srv.c <- ctx
}

//AddIOService ...
func (m *Manager) AddIOService(diskID uint32) *IOService {
	v := &IOService{
		c: make(chan *ioContext),
	}
	m.mutex.Lock()
	m.ios[diskID] = v
	m.mutex.Unlock()

	go func() {
		m.wg.Add(1)
	label:
		for {
			select {
			case <-m.cc:
				fmt.Println("Exit")
				break label
			case i, ok := <-v.c:
				if ok {
					v, ok := i.msg.(*network.ReadDiskMsg)
					if ok {
						v.Header.Type = network.PacketTypeReadDiskAck
						buf := m.ReadData(v.DiskID, v.SectorCount, v.SectorOffset)
						//fmt.Println(r, "buf read: ", len(buf))
						v.Header.Len = uint32(len(buf)) + 38
						v.Data = buf

						if !i.writer.WriteMsg(v) {
							log.Println("Read data ack failed!")
						}
					}
				}
				break
			}
		}
		m.wg.Done()
	}()

	return v
}

//GetIOService ...
func (m *Manager) GetIOService(diskID uint32) *IOService {
	m.mutex.Lock()
	v, ok := m.ios[diskID]
	m.mutex.Unlock()
	if !ok {
		return nil
	}
	return v
}

//Close ...
func (m *Manager) Close() {
	close(m.cc)
	m.wg.Wait()
}
