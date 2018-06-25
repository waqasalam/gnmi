package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aristanetworks/glog"
	"github.com/aristanetworks/goarista/gnmi"
	"github.com/golang/protobuf/proto"
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
)

type gnmiInfo struct {
	ctx context.Context
	clt gnmipb.GNMIClient
}

func newGNMIInfo(cfg *gnmi.Config) (*gnmiInfo, error) {
	var err error
	gi := &gnmiInfo{}
	gi.ctx = gnmi.NewContext(context.Background(), cfg)
	gi.clt, err = gnmi.Dial(cfg)
	if err != nil {
		glog.Fatal(err)
		return nil, err
	}
	return gi, nil
}

// handleUpdate parses a protobuf message received from the target. This implementation handles only
// gNMI SubscribeResponse messages.
func (s *gnmiInfo) handleResponse(msg proto.Message) error {
	resp, ok := msg.(*gnmipb.SubscribeResponse)
	if !ok {
		return fmt.Errorf("failed to type assert message %#v", msg)
	}
	switch v := resp.Response.(type) {
	case *gnmipb.SubscribeResponse_Update:

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
