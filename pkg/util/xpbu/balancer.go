package pbu

import (
	"fmt"
	"strconv"

	"github.com/douyu/jupiter/pkg/constant"
	"github.com/douyu/jupiter/pkg/server"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

const (
	Name = "pbu"
)

type pbuPickerBuilder struct {
	userCap int64
}

// newBuilder creates a new balance builder.
func newBuilder(userCap int64) balancer.Builder {
	return base.NewBalancerBuilderV2(Name, &pbuPickerBuilder{userCap}, base.Config{HealthCheck: true})
}

func RegisterBalancer(userCap int64) {
	balancer.Register(newBuilder(userCap))
}

// attributeKey is the type used as the key to store AddrInfo in the Attributes
// field of resolver.Address.
type attributeKey struct{}

// AddrInfo will be stored inside Address metadata in order to use weighted balancer.
type AddrInfo struct {
	UserGroupID int64
}

// SetAddrInfo returns a copy of addr in which the Attributes field is updated
// with addrInfo.
func SetAddrInfo(addr resolver.Address, addrInfo AddrInfo) resolver.Address {
	addr.Attributes = attributes.New()
	addr.Attributes = addr.Attributes.WithValues(attributeKey{}, addrInfo)
	return addr
}

// GetAddrInfo returns the AddrInfo stored in the Attributes fields of addr.
func GetAddrInfo(addr resolver.Address) AddrInfo {
	v := addr.Attributes.Value(attributeKey{})
	ai, _ := v.(AddrInfo)
	return ai
}

type pbuPicker struct {
	pbu Pbu
}

func (c *pbuPickerBuilder) Build(info base.PickerBuildInfo) balancer.V2Picker {
	grpclog.Infof("pbuPicker: newPicker called with info: %v", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPickerV2(balancer.ErrNoSubConnAvailable)
	}

	var p = New(c.userCap)

	for subConn, info := range info.ReadySCs {
		if info.Address.Attributes != nil {
			if serviceInfo, ok := info.Address.Attributes.Value(constant.KeyServiceInfo).(server.ServiceInfo); ok {
				var weight = int64(serviceInfo.Weight)
				p.Add(weight, subConn)
			}
		}
	}

	return &pbuPicker{
		pbu: p,
	}
}

func (p *pbuPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var result balancer.PickResult
	var md, ok = metadata.FromOutgoingContext(info.Ctx)
	if !ok {
		return result, fmt.Errorf("Metadata not found in context")
	}

	var uids = md.Get("X-AEX-UID")
	if len(uids) == 0 {
		return result, fmt.Errorf("Key 'X-AEX-UID' nof found")
	}

	var uid, err = strconv.ParseInt(uids[0], 0, 64)
	if err != nil {
		return result, fmt.Errorf("Invalid uid: %v", err)
	}

	if uid <= 0 {
		return result, fmt.Errorf("Invalid uid: %d", uid)
	}

	var item, _ = p.pbu.Get(uid)
	if item == nil {
		return balancer.PickResult{}, fmt.Errorf("Uid %d not supported by balancer", uid)
	}

	return balancer.PickResult{SubConn: item.(balancer.SubConn)}, nil
}
