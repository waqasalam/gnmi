package main

import (
	"github.com/aristanetworks/glog"
	"reflect"
)

type intfStats struct {
	InBroadcastPkts  uint64
	InDiscards       uint32
	InErrors         uint32
	InMulticastPkts  uint64
	InOctets         uint64
	InUnicastPkts    uint64
	IutBroadcastPkts uint64
	OutDiscards      uint32
	OutErrors        uint32
	OutMulticastPkts uint64
	OutOctetsPkts    uint64
	OutUnicastPkts   uint64
}

type intf struct {
	stats intfStats
}

type intfDB struct {
	db map[string]*intf
}

func newIntfDB() *intfDB {
	return &intfDB{db: make(map[string]*intf)}
}

func (db *intfDB) addIntf(name string) *intf {
	db.db[name] = &intf{}
	return db.db[name]
}

func (db *intfDB) lookupIntf(name string) *intf {
	if intf, ok := db.db[name]; ok {
		return intf
	}
	glog.Info("Lookup failed", name)
	return nil
}

func (db *intfDB) setField(in *intfStats, field string, value interface{}) {
	r := reflect.ValueOf(in)
	f := r.Elem().FieldByName(field)

	if f.IsValid() {
		if f.CanSet() {
			glog.Info("Kind", f.Kind())
			switch f.Kind() {
			case reflect.Uint64:
				v := value.(uint64)
				f.SetUint(v)

			case reflect.Uint32:
				v := value.(uint64)
				f.SetUint(v)
			default:
				glog.Info("Unknow type")
			}

		}
	}

}
