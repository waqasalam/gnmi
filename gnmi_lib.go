package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aristanetworks/glog"
	"github.com/aristanetworks/goarista/gnmi"
	"github.com/golang/protobuf/proto"
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
	"strings"
	_ "time"
)

type gnmiInfo struct {
	ctx context.Context
	clt gnmipb.GNMIClient
	idb *intfDB
}

type gnmiOper int

const (
	gnmiUnknown gnmiOper = 0
	gnmiState   gnmiOper = 1
	gnmiConfig  gnmiOper = 2
)

func newGNMIInfo(cfg *gnmi.Config, idb *intfDB) (*gnmiInfo, error) {
	var err error
	gi := &gnmiInfo{idb: idb}
	gi.ctx = gnmi.NewContext(context.Background(), cfg)
	gi.clt, err = gnmi.Dial(cfg)
	if err != nil {
		glog.Fatal(err)
		return nil, err
	}
	return gi, nil
}

func (gi *gnmiInfo) validateInterface(pelem []*gnmipb.PathElem) (*intf, string, error) {
	state := gnmiUnknown
	counters := false
	intfname, field := "", ""
	var in *intf
	for _, elm := range pelem {
		if counters && state == gnmiState && intfname != "" {

			in = gi.idb.lookupIntf(intfname)
			if in == nil {
				in = gi.idb.addIntf(intfname)
			}
			field = elm.Name
		}
		switch {
		case elm.Name == "interface":
			for _, v := range elm.Key {
				intfname = v
				break
			}
		case elm.Name == "state":
			state = gnmiState

		case elm.Name == "counters":
			counters = true

		}
	}

	if state == gnmiUnknown || !counters || intfname == "" || field == "" {
		return nil, "", fmt.Errorf("Interface encoding wrong")
	}
	return in, field, nil
}
func (gi *gnmiInfo) decodeElem(pelem []*gnmipb.PathElem) (*intf, string, error) {

	elm := pelem[0]
	glog.Info(elm.Name)
	switch {
	case elm.Name == "interfaces":
		return gi.validateInterface(pelem[1:])

	}
	return nil, "", fmt.Errorf("Unknown Operation")
}

func (gi *gnmiInfo) decodePath(path *gnmipb.Path) (*intf, string, error) {
	switch {
	case path == nil:
		return nil, "", fmt.Errorf("Invalid Encoding")
	case len(path.Elem) != 0:

		return gi.decodeElem(path.Elem)
	case len(path.Element) != 0:
		return nil, "", fmt.Errorf("Deprecated encoding")
	}
	return nil, "", fmt.Errorf("Wrong Encoding")
}
func setval(val *gnmipb.TypedValue) string {
	switch v := val.GetValue().(type) {
	case *gnmipb.TypedValue_StringVal:
		return "string"
	case *gnmipb.TypedValue_JsonIetfVal:
		return "json"
	case *gnmipb.TypedValue_JsonVal:
		return "jsonval"
	case *gnmipb.TypedValue_IntVal:
		return "int"
	case *gnmipb.TypedValue_UintVal:
		return "uint"
	case *gnmipb.TypedValue_BoolVal:
		return "boolval"
	case *gnmipb.TypedValue_BytesVal:
		return "byteps"
	case *gnmipb.TypedValue_DecimalVal:
		return "decimal"
	case *gnmipb.TypedValue_FloatVal:
		return "float"
	case *gnmipb.TypedValue_LeaflistVal:
		return "leaf"
	case *gnmipb.TypedValue_AsciiVal:
		return "ascii"
	case *gnmipb.TypedValue_AnyVal:
		return "anyval"
	default:
		panic(v)
	}
	return ""
}

func kebabCaseToCamelCase(kebab string) (camelCase string) {
	isToUpper := true
	for _, runeValue := range kebab {
		if isToUpper {
			camelCase += strings.ToUpper(string(runeValue))
			isToUpper = false
		} else {
			if runeValue == '-' {
				isToUpper = true
			} else {
				camelCase += string(runeValue)
			}
		}
	}
	return
}
func (gi *gnmiInfo) handleUpdate(n *gnmipb.Notification) {
	gi.decodePath(n.Prefix)
	count := 0
	for _, update := range n.Update {
		intf, field, err := gi.decodePath(update.Path)
		if err != nil {
			glog.Info("Unable to decode gnmi update", err)
			return
		}
		val, err := value.ToScalar(update.Val)
		if err != nil {
			glog.Info("Unable to decode gnmi update", err)
			return
		}

		gi.idb.setField(&intf.stats, kebabCaseToCamelCase(field), val)
		count++

	}
}

// handleUpdate parses a protobuf message received from the target. This implementation handles only
// gNMI SubscribeResponse messages.
func (gi *gnmiInfo) handleResponse(msg proto.Message) error {
	resp, ok := msg.(*gnmipb.SubscribeResponse)
	if !ok {
		return fmt.Errorf("failed to type assert message %#v", msg)
	}
	switch v := resp.Response.(type) {
	case *gnmipb.SubscribeResponse_Update:
		glog.Info("response received")
		gi.handleUpdate(v.Update)
	case *gnmipb.SubscribeResponse_SyncResponse:

	case *gnmipb.SubscribeResponse_Error:
		return errors.New(v.Error.Message)
	default:
		return fmt.Errorf("unknown response %T: %s", v, v)
	}
	return nil
}

func (gi *gnmiInfo) handleOpSubscribe(args []string) {
	respChan := make(chan *gnmipb.SubscribeResponse)
	errChan := make(chan error)
	defer close(respChan)
	defer close(errChan)
	go gnmi.Subscribe(gi.ctx, gi.clt, gnmi.SplitPaths(args), respChan, errChan)
	for {
		select {
		case resp := <-respChan:
			if err := gi.handleResponse(resp); err != nil {
				glog.Fatal(err)
			}
		case err := <-errChan:
			glog.Fatal(err)
		}
	}
}

func (gi *gnmiInfo) handleOpGet(args []string) {
	glog.Info("args ", args)
	err := gnmi.Get(gi.ctx, gi.clt, gnmi.SplitPaths(args))
	if err != nil {
		glog.Fatal(err)
	}
}

func (gi *gnmiInfo) handleOpCapabilities(args []string) {
	err := gnmi.Capabilities(gi.ctx, gi.clt)
	if err != nil {
		glog.Fatal(err)
	}
}
