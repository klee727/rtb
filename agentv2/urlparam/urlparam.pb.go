// Code generated by protoc-gen-go.
// source: urlparam.proto
// DO NOT EDIT!

/*
Package urlparam is a generated protocol buffer package.

It is generated from these files:
	urlparam.proto

It has these top-level messages:
	TdSegmentInfo
	UrlParam
*/
package urlparam

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TdSegmentInfo struct {
	DeviceId         *string `protobuf:"bytes,1,opt,name=deviceId" json:"deviceId,omitempty"`
	DeviceIdType     *string `protobuf:"bytes,2,opt,name=deviceIdType" json:"deviceIdType,omitempty"`
	DeviceIdEncode   *string `protobuf:"bytes,3,opt,name=deviceIdEncode" json:"deviceIdEncode,omitempty"`
	SegmentId        *string `protobuf:"bytes,4,opt,name=segmentId" json:"segmentId,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TdSegmentInfo) Reset()         { *m = TdSegmentInfo{} }
func (m *TdSegmentInfo) String() string { return proto.CompactTextString(m) }
func (*TdSegmentInfo) ProtoMessage()    {}

func (m *TdSegmentInfo) GetDeviceId() string {
	if m != nil && m.DeviceId != nil {
		return *m.DeviceId
	}
	return ""
}

func (m *TdSegmentInfo) GetDeviceIdType() string {
	if m != nil && m.DeviceIdType != nil {
		return *m.DeviceIdType
	}
	return ""
}

func (m *TdSegmentInfo) GetDeviceIdEncode() string {
	if m != nil && m.DeviceIdEncode != nil {
		return *m.DeviceIdEncode
	}
	return ""
}

func (m *TdSegmentInfo) GetSegmentId() string {
	if m != nil && m.SegmentId != nil {
		return *m.SegmentId
	}
	return ""
}

type UrlParam struct {
	AuctionId        *string        `protobuf:"bytes,1,req,name=auctionId" json:"auctionId,omitempty"`
	BiddingTimestamp *float64       `protobuf:"fixed64,2,req,name=biddingTimestamp" json:"biddingTimestamp,omitempty"`
	ImpressionId     *string        `protobuf:"bytes,3,req,name=impressionId" json:"impressionId,omitempty"`
	UserIds          *string        `protobuf:"bytes,4,req,name=userIds" json:"userIds,omitempty"`
	Account          *string        `protobuf:"bytes,5,req,name=account" json:"account,omitempty"`
	CreativeId       *string        `protobuf:"bytes,6,req,name=creativeId" json:"creativeId,omitempty"`
	BigtreeAppId     *int32         `protobuf:"varint,7,req,name=bigtreeAppId" json:"bigtreeAppId,omitempty"`
	Ip               *string        `protobuf:"bytes,8,req,name=ip" json:"ip,omitempty"`
	Os               *string        `protobuf:"bytes,9,req,name=os" json:"os,omitempty"`
	Exchange         *string        `protobuf:"bytes,10,req,name=exchange" json:"exchange,omitempty"`
	Provider         *string        `protobuf:"bytes,11,req,name=provider" json:"provider,omitempty"`
	DeviceId         *string        `protobuf:"bytes,12,req,name=deviceId" json:"deviceId,omitempty"`
	ChargeType       *string        `protobuf:"bytes,13,opt,name=chargeType" json:"chargeType,omitempty"`
	ChargePrice      *string        `protobuf:"bytes,14,opt,name=chargePrice" json:"chargePrice,omitempty"`
	TagsHash         []int32        `protobuf:"varint,15,rep,name=tagsHash" json:"tagsHash,omitempty"`
	IpHash           []int32        `protobuf:"varint,16,rep,name=ipHash" json:"ipHash,omitempty"`
	CreativeHash     *string        `protobuf:"bytes,17,opt,name=creativeHash" json:"creativeHash,omitempty"`
	TdSegmentInfo    *TdSegmentInfo `protobuf:"bytes,18,opt,name=tdSegmentInfo" json:"tdSegmentInfo,omitempty"`
	Mac              *string        `protobuf:"bytes,19,opt,name=mac" json:"mac,omitempty"`
	BidPrice         *int64         `protobuf:"varint,20,opt,name=bidPrice" json:"bidPrice,omitempty"`
	OsVersion        *string        `protobuf:"bytes,21,opt,name=osVersion" json:"osVersion,omitempty"`
	DeviceModel      *string        `protobuf:"bytes,22,opt,name=deviceModel" json:"deviceModel,omitempty"`
	AdvId            *string        `protobuf:"bytes,24,opt,name=advId" json:"advId,omitempty"`
	Agent            *string        `protobuf:"bytes,23,opt,name=agent" json:"agent,omitempty"`
	BiddingKey       *string        `protobuf:"bytes,25,opt,name=biddingKey" json:"biddingKey,omitempty"`
	AdunitId         *string        `protobuf:"bytes,26,opt,name=adunitId" json:"adunitId,omitempty"`
	DeviceType       *int32         `protobuf:"varint,27,opt,name=deviceType" json:"deviceType,omitempty"`
	ConnectionType   *int32         `protobuf:"varint,28,opt,name=connectionType" json:"connectionType,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *UrlParam) Reset()         { *m = UrlParam{} }
func (m *UrlParam) String() string { return proto.CompactTextString(m) }
func (*UrlParam) ProtoMessage()    {}

func (m *UrlParam) GetAuctionId() string {
	if m != nil && m.AuctionId != nil {
		return *m.AuctionId
	}
	return ""
}

func (m *UrlParam) GetBiddingTimestamp() float64 {
	if m != nil && m.BiddingTimestamp != nil {
		return *m.BiddingTimestamp
	}
	return 0
}

func (m *UrlParam) GetImpressionId() string {
	if m != nil && m.ImpressionId != nil {
		return *m.ImpressionId
	}
	return ""
}

func (m *UrlParam) GetUserIds() string {
	if m != nil && m.UserIds != nil {
		return *m.UserIds
	}
	return ""
}

func (m *UrlParam) GetAccount() string {
	if m != nil && m.Account != nil {
		return *m.Account
	}
	return ""
}

func (m *UrlParam) GetCreativeId() string {
	if m != nil && m.CreativeId != nil {
		return *m.CreativeId
	}
	return ""
}

func (m *UrlParam) GetBigtreeAppId() int32 {
	if m != nil && m.BigtreeAppId != nil {
		return *m.BigtreeAppId
	}
	return 0
}

func (m *UrlParam) GetIp() string {
	if m != nil && m.Ip != nil {
		return *m.Ip
	}
	return ""
}

func (m *UrlParam) GetOs() string {
	if m != nil && m.Os != nil {
		return *m.Os
	}
	return ""
}

func (m *UrlParam) GetExchange() string {
	if m != nil && m.Exchange != nil {
		return *m.Exchange
	}
	return ""
}

func (m *UrlParam) GetProvider() string {
	if m != nil && m.Provider != nil {
		return *m.Provider
	}
	return ""
}

func (m *UrlParam) GetDeviceId() string {
	if m != nil && m.DeviceId != nil {
		return *m.DeviceId
	}
	return ""
}

func (m *UrlParam) GetChargeType() string {
	if m != nil && m.ChargeType != nil {
		return *m.ChargeType
	}
	return ""
}

func (m *UrlParam) GetChargePrice() string {
	if m != nil && m.ChargePrice != nil {
		return *m.ChargePrice
	}
	return ""
}

func (m *UrlParam) GetTagsHash() []int32 {
	if m != nil {
		return m.TagsHash
	}
	return nil
}

func (m *UrlParam) GetIpHash() []int32 {
	if m != nil {
		return m.IpHash
	}
	return nil
}

func (m *UrlParam) GetCreativeHash() string {
	if m != nil && m.CreativeHash != nil {
		return *m.CreativeHash
	}
	return ""
}

func (m *UrlParam) GetTdSegmentInfo() *TdSegmentInfo {
	if m != nil {
		return m.TdSegmentInfo
	}
	return nil
}

func (m *UrlParam) GetMac() string {
	if m != nil && m.Mac != nil {
		return *m.Mac
	}
	return ""
}

func (m *UrlParam) GetBidPrice() int64 {
	if m != nil && m.BidPrice != nil {
		return *m.BidPrice
	}
	return 0
}

func (m *UrlParam) GetOsVersion() string {
	if m != nil && m.OsVersion != nil {
		return *m.OsVersion
	}
	return ""
}

func (m *UrlParam) GetDeviceModel() string {
	if m != nil && m.DeviceModel != nil {
		return *m.DeviceModel
	}
	return ""
}

func (m *UrlParam) GetAdvId() string {
	if m != nil && m.AdvId != nil {
		return *m.AdvId
	}
	return ""
}

func (m *UrlParam) GetAgent() string {
	if m != nil && m.Agent != nil {
		return *m.Agent
	}
	return ""
}

func (m *UrlParam) GetBiddingKey() string {
	if m != nil && m.BiddingKey != nil {
		return *m.BiddingKey
	}
	return ""
}

func (m *UrlParam) GetAdunitId() string {
	if m != nil && m.AdunitId != nil {
		return *m.AdunitId
	}
	return ""
}

func (m *UrlParam) GetDeviceType() int32 {
	if m != nil && m.DeviceType != nil {
		return *m.DeviceType
	}
	return 0
}

func (m *UrlParam) GetConnectionType() int32 {
	if m != nil && m.ConnectionType != nil {
		return *m.ConnectionType
	}
	return 0
}