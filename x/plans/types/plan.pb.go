// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lavanet/lava/plans/plan.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/lavanet/lava/v2/x/spec/types"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// The geolocation values are encoded as bits in a bitmask, with two special values:
// GLS is set to 0 so it will be restrictive with the AND operator.
// GL is set to -1 so it will be permissive with the AND operator.
type Geolocation int32

const (
	Geolocation_GLS Geolocation = 0
	Geolocation_USC Geolocation = 1
	Geolocation_EU  Geolocation = 2
	Geolocation_USE Geolocation = 4
	Geolocation_USW Geolocation = 8
	Geolocation_AF  Geolocation = 16
	Geolocation_AS  Geolocation = 32
	Geolocation_AU  Geolocation = 64
	Geolocation_GL  Geolocation = 65535
)

var Geolocation_name = map[int32]string{
	0:     "GLS",
	1:     "USC",
	2:     "EU",
	4:     "USE",
	8:     "USW",
	16:    "AF",
	32:    "AS",
	64:    "AU",
	65535: "GL",
}

var Geolocation_value = map[string]int32{
	"GLS": 0,
	"USC": 1,
	"EU":  2,
	"USE": 4,
	"USW": 8,
	"AF":  16,
	"AS":  32,
	"AU":  64,
	"GL":  65535,
}

func (x Geolocation) String() string {
	return proto.EnumName(Geolocation_name, int32(x))
}

func (Geolocation) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_64c3707a3b09a2e5, []int{0}
}

type Plan struct {
	Index                    string     `protobuf:"bytes,1,opt,name=index,proto3" json:"index"`
	Block                    uint64     `protobuf:"varint,3,opt,name=block,proto3" json:"block"`
	Price                    types.Coin `protobuf:"bytes,4,opt,name=price,proto3" json:"price"`
	AllowOveruse             bool       `protobuf:"varint,8,opt,name=allow_overuse,json=allowOveruse,proto3" json:"allow_overuse"`
	OveruseRate              uint64     `protobuf:"varint,9,opt,name=overuse_rate,json=overuseRate,proto3" json:"overuse_rate"`
	Description              string     `protobuf:"bytes,11,opt,name=description,proto3" json:"description"`
	Type                     string     `protobuf:"bytes,12,opt,name=type,proto3" json:"type"`
	AnnualDiscountPercentage uint64     `protobuf:"varint,13,opt,name=annual_discount_percentage,json=annualDiscountPercentage,proto3" json:"annual_discount_percentage"`
	PlanPolicy               Policy     `protobuf:"bytes,14,opt,name=plan_policy,json=planPolicy,proto3" json:"plan_policy"`
	ProjectsLimit            uint64     `protobuf:"varint,15,opt,name=projects_limit,json=projectsLimit,proto3" json:"projects_limit"`
	AllowedBuyers            []string   `protobuf:"bytes,16,rep,name=allowed_buyers,json=allowedBuyers,proto3" json:"allowed_buyers"`
}

func (m *Plan) Reset()         { *m = Plan{} }
func (m *Plan) String() string { return proto.CompactTextString(m) }
func (*Plan) ProtoMessage()    {}
func (*Plan) Descriptor() ([]byte, []int) {
	return fileDescriptor_64c3707a3b09a2e5, []int{0}
}
func (m *Plan) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Plan) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Plan.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Plan) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Plan.Merge(m, src)
}
func (m *Plan) XXX_Size() int {
	return m.Size()
}
func (m *Plan) XXX_DiscardUnknown() {
	xxx_messageInfo_Plan.DiscardUnknown(m)
}

var xxx_messageInfo_Plan proto.InternalMessageInfo

func (m *Plan) GetIndex() string {
	if m != nil {
		return m.Index
	}
	return ""
}

func (m *Plan) GetBlock() uint64 {
	if m != nil {
		return m.Block
	}
	return 0
}

func (m *Plan) GetPrice() types.Coin {
	if m != nil {
		return m.Price
	}
	return types.Coin{}
}

func (m *Plan) GetAllowOveruse() bool {
	if m != nil {
		return m.AllowOveruse
	}
	return false
}

func (m *Plan) GetOveruseRate() uint64 {
	if m != nil {
		return m.OveruseRate
	}
	return 0
}

func (m *Plan) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Plan) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Plan) GetAnnualDiscountPercentage() uint64 {
	if m != nil {
		return m.AnnualDiscountPercentage
	}
	return 0
}

func (m *Plan) GetPlanPolicy() Policy {
	if m != nil {
		return m.PlanPolicy
	}
	return Policy{}
}

func (m *Plan) GetProjectsLimit() uint64 {
	if m != nil {
		return m.ProjectsLimit
	}
	return 0
}

func (m *Plan) GetAllowedBuyers() []string {
	if m != nil {
		return m.AllowedBuyers
	}
	return nil
}

func init() {
	proto.RegisterEnum("lavanet.lava.plans.Geolocation", Geolocation_name, Geolocation_value)
	proto.RegisterType((*Plan)(nil), "lavanet.lava.plans.Plan")
}

func init() { proto.RegisterFile("lavanet/lava/plans/plan.proto", fileDescriptor_64c3707a3b09a2e5) }

var fileDescriptor_64c3707a3b09a2e5 = []byte{
	// 607 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0x4f, 0x6f, 0xd3, 0x3e,
	0x18, 0x6e, 0xda, 0xb4, 0x4b, 0x9d, 0x75, 0xf3, 0xcf, 0x3f, 0x0e, 0xa1, 0x82, 0xa4, 0xe2, 0x80,
	0x2a, 0x0e, 0x89, 0xb6, 0x49, 0x48, 0x5c, 0x10, 0x74, 0x6c, 0x93, 0xaa, 0x49, 0x54, 0x99, 0x26,
	0x24, 0x40, 0x8a, 0x5c, 0xd7, 0x2a, 0x06, 0x2f, 0x8e, 0x12, 0xb7, 0x6c, 0x5f, 0x81, 0x13, 0x1f,
	0x83, 0x8f, 0xb2, 0xe3, 0x8e, 0x9c, 0x22, 0xd4, 0xdd, 0xf2, 0x25, 0x86, 0x6c, 0x67, 0x68, 0x05,
	0xc4, 0xc5, 0x8f, 0x9f, 0xf7, 0x79, 0xde, 0xf6, 0xfd, 0x13, 0x83, 0x87, 0x1c, 0x2f, 0x71, 0x4a,
	0x65, 0xa4, 0x30, 0xca, 0x38, 0x4e, 0x0b, 0x7d, 0x86, 0x59, 0x2e, 0xa4, 0x40, 0xa8, 0x96, 0x43,
	0x85, 0xa1, 0x96, 0xfb, 0xf7, 0xe6, 0x62, 0x2e, 0xb4, 0x1c, 0xa9, 0x9b, 0x71, 0xf6, 0x7d, 0x22,
	0x8a, 0x33, 0x51, 0x44, 0x53, 0x5c, 0xd0, 0x68, 0xb9, 0x33, 0xa5, 0x12, 0xef, 0x44, 0x44, 0xb0,
	0xfa, 0x97, 0xfa, 0x8f, 0xd7, 0xfe, 0xa8, 0xc8, 0x28, 0x89, 0x70, 0xc6, 0x12, 0x22, 0x38, 0xa7,
	0x44, 0x32, 0x71, 0xeb, 0x0b, 0xfe, 0x56, 0x90, 0xe0, 0x8c, 0x5c, 0x18, 0xc3, 0xa3, 0x2f, 0x6d,
	0x60, 0x4f, 0x38, 0x4e, 0x51, 0x00, 0xda, 0x2c, 0x9d, 0xd1, 0x73, 0xcf, 0x1a, 0x58, 0xc3, 0xee,
	0xa8, 0x5b, 0x95, 0x81, 0x09, 0xc4, 0x06, 0x94, 0x61, 0xca, 0x05, 0xf9, 0xe4, 0xb5, 0x06, 0xd6,
	0xd0, 0x36, 0x06, 0x1d, 0x88, 0x0d, 0xa0, 0xe7, 0xa0, 0x9d, 0xe5, 0x8c, 0x50, 0xcf, 0x1e, 0x58,
	0x43, 0x77, 0xf7, 0x7e, 0x68, 0x7a, 0x08, 0x55, 0x0f, 0x61, 0xdd, 0x43, 0xb8, 0x2f, 0x58, 0x3a,
	0xea, 0x5d, 0x96, 0x41, 0x43, 0xe5, 0x6b, 0x7f, 0x6c, 0x00, 0x3d, 0x05, 0x3d, 0xcc, 0xb9, 0xf8,
	0x9c, 0x88, 0x25, 0xcd, 0x17, 0x05, 0xf5, 0x9c, 0x81, 0x35, 0x74, 0x46, 0xff, 0x55, 0x65, 0xb0,
	0x2e, 0xc4, 0x9b, 0x9a, 0xbe, 0x36, 0x0c, 0xed, 0x81, 0xcd, 0x5a, 0x48, 0x72, 0x2c, 0xa9, 0xd7,
	0xd5, 0xf5, 0xc1, 0xaa, 0x0c, 0xd6, 0xe2, 0xb1, 0x7b, 0x9b, 0x8e, 0x25, 0x45, 0x3b, 0xc0, 0x9d,
	0xd1, 0x82, 0xe4, 0x2c, 0x53, 0xd3, 0xf2, 0x5c, 0xdd, 0xf4, 0x76, 0x55, 0x06, 0x77, 0xc3, 0xf1,
	0x5d, 0x82, 0x1e, 0x00, 0x5b, 0x5e, 0x64, 0xd4, 0xdb, 0xd4, 0x5e, 0xa7, 0x2a, 0x03, 0xcd, 0x63,
	0x7d, 0xa2, 0xf7, 0xa0, 0x8f, 0xd3, 0x74, 0x81, 0x79, 0x32, 0x63, 0x05, 0x11, 0x8b, 0x54, 0x26,
	0x19, 0xcd, 0x09, 0x4d, 0x25, 0x9e, 0x53, 0xaf, 0xa7, 0x6b, 0xf2, 0xab, 0x32, 0xf8, 0x87, 0x2b,
	0xf6, 0x8c, 0xf6, 0xaa, 0x96, 0x26, 0xbf, 0x14, 0x34, 0x01, 0xae, 0x5a, 0x5e, 0x62, 0x76, 0xe7,
	0x6d, 0xe9, 0x09, 0xf7, 0xc3, 0x3f, 0xbf, 0xa7, 0x70, 0xa2, 0x1d, 0xa3, 0xff, 0xeb, 0x11, 0xdf,
	0x4d, 0x8b, 0x81, 0x22, 0xc6, 0x80, 0x9e, 0x81, 0xad, 0x2c, 0x17, 0x1f, 0x29, 0x91, 0x45, 0xc2,
	0xd9, 0x19, 0x93, 0xde, 0xb6, 0xae, 0x11, 0x55, 0x65, 0xf0, 0x9b, 0x12, 0xf7, 0x6e, 0xf9, 0xb1,
	0xa2, 0x2a, 0x55, 0x2f, 0x80, 0xce, 0x92, 0xe9, 0xe2, 0x82, 0xe6, 0x85, 0x07, 0x07, 0xad, 0x61,
	0xd7, 0xa4, 0xae, 0x2b, 0x71, 0xaf, 0xe6, 0x23, 0x4d, 0xc7, 0xb6, 0x03, 0xa0, 0x3b, 0xb6, 0x9d,
	0x26, 0x6c, 0x8d, 0x6d, 0xa7, 0x0d, 0x3b, 0x63, 0xdb, 0xe9, 0xc0, 0x8d, 0xb1, 0xed, 0x6c, 0x40,
	0xe7, 0xc9, 0x3b, 0xe0, 0x1e, 0x51, 0xc1, 0x05, 0xc1, 0x7a, 0xe0, 0x1b, 0xa0, 0x75, 0x74, 0x7c,
	0x02, 0x1b, 0xea, 0x72, 0x7a, 0xb2, 0x0f, 0x2d, 0xd4, 0x01, 0xcd, 0x83, 0x53, 0xd8, 0x34, 0x81,
	0x03, 0x68, 0x9b, 0xcb, 0x1b, 0xe8, 0x28, 0xe5, 0xe5, 0x21, 0x84, 0x1a, 0x4f, 0xe0, 0x40, 0xe3,
	0x29, 0x7c, 0x81, 0x1c, 0xd0, 0x3c, 0x3a, 0x86, 0x37, 0x37, 0xad, 0xd1, 0xe1, 0xb7, 0x95, 0x6f,
	0x5d, 0xae, 0x7c, 0xeb, 0x6a, 0xe5, 0x5b, 0x3f, 0x56, 0xbe, 0xf5, 0xf5, 0xda, 0x6f, 0x5c, 0x5d,
	0xfb, 0x8d, 0xef, 0xd7, 0x7e, 0xe3, 0xed, 0x70, 0xce, 0xe4, 0x87, 0xc5, 0x34, 0x24, 0xe2, 0x2c,
	0x5a, 0x7b, 0x33, 0xcb, 0xdd, 0xe8, 0xbc, 0x7e, 0x38, 0x6a, 0xcf, 0xc5, 0xb4, 0xa3, 0x1f, 0xce,
	0xde, 0xcf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xdb, 0xde, 0x75, 0xec, 0xec, 0x03, 0x00, 0x00,
}

func (this *Plan) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Plan)
	if !ok {
		that2, ok := that.(Plan)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Index != that1.Index {
		return false
	}
	if this.Block != that1.Block {
		return false
	}
	if !this.Price.Equal(&that1.Price) {
		return false
	}
	if this.AllowOveruse != that1.AllowOveruse {
		return false
	}
	if this.OveruseRate != that1.OveruseRate {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if this.Type != that1.Type {
		return false
	}
	if this.AnnualDiscountPercentage != that1.AnnualDiscountPercentage {
		return false
	}
	if !this.PlanPolicy.Equal(&that1.PlanPolicy) {
		return false
	}
	if this.ProjectsLimit != that1.ProjectsLimit {
		return false
	}
	if len(this.AllowedBuyers) != len(that1.AllowedBuyers) {
		return false
	}
	for i := range this.AllowedBuyers {
		if this.AllowedBuyers[i] != that1.AllowedBuyers[i] {
			return false
		}
	}
	return true
}
func (m *Plan) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Plan) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Plan) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AllowedBuyers) > 0 {
		for iNdEx := len(m.AllowedBuyers) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllowedBuyers[iNdEx])
			copy(dAtA[i:], m.AllowedBuyers[iNdEx])
			i = encodeVarintPlan(dAtA, i, uint64(len(m.AllowedBuyers[iNdEx])))
			i--
			dAtA[i] = 0x1
			i--
			dAtA[i] = 0x82
		}
	}
	if m.ProjectsLimit != 0 {
		i = encodeVarintPlan(dAtA, i, uint64(m.ProjectsLimit))
		i--
		dAtA[i] = 0x78
	}
	{
		size, err := m.PlanPolicy.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPlan(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x72
	if m.AnnualDiscountPercentage != 0 {
		i = encodeVarintPlan(dAtA, i, uint64(m.AnnualDiscountPercentage))
		i--
		dAtA[i] = 0x68
	}
	if len(m.Type) > 0 {
		i -= len(m.Type)
		copy(dAtA[i:], m.Type)
		i = encodeVarintPlan(dAtA, i, uint64(len(m.Type)))
		i--
		dAtA[i] = 0x62
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintPlan(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x5a
	}
	if m.OveruseRate != 0 {
		i = encodeVarintPlan(dAtA, i, uint64(m.OveruseRate))
		i--
		dAtA[i] = 0x48
	}
	if m.AllowOveruse {
		i--
		if m.AllowOveruse {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x40
	}
	{
		size, err := m.Price.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintPlan(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if m.Block != 0 {
		i = encodeVarintPlan(dAtA, i, uint64(m.Block))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Index) > 0 {
		i -= len(m.Index)
		copy(dAtA[i:], m.Index)
		i = encodeVarintPlan(dAtA, i, uint64(len(m.Index)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintPlan(dAtA []byte, offset int, v uint64) int {
	offset -= sovPlan(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Plan) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Index)
	if l > 0 {
		n += 1 + l + sovPlan(uint64(l))
	}
	if m.Block != 0 {
		n += 1 + sovPlan(uint64(m.Block))
	}
	l = m.Price.Size()
	n += 1 + l + sovPlan(uint64(l))
	if m.AllowOveruse {
		n += 2
	}
	if m.OveruseRate != 0 {
		n += 1 + sovPlan(uint64(m.OveruseRate))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovPlan(uint64(l))
	}
	l = len(m.Type)
	if l > 0 {
		n += 1 + l + sovPlan(uint64(l))
	}
	if m.AnnualDiscountPercentage != 0 {
		n += 1 + sovPlan(uint64(m.AnnualDiscountPercentage))
	}
	l = m.PlanPolicy.Size()
	n += 1 + l + sovPlan(uint64(l))
	if m.ProjectsLimit != 0 {
		n += 1 + sovPlan(uint64(m.ProjectsLimit))
	}
	if len(m.AllowedBuyers) > 0 {
		for _, s := range m.AllowedBuyers {
			l = len(s)
			n += 2 + l + sovPlan(uint64(l))
		}
	}
	return n
}

func sovPlan(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPlan(x uint64) (n int) {
	return sovPlan(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Plan) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPlan
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Plan: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Plan: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPlan
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPlan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Index = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Block", wireType)
			}
			m.Block = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Block |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Price", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPlan
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPlan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Price.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowOveruse", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.AllowOveruse = bool(v != 0)
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OveruseRate", wireType)
			}
			m.OveruseRate = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.OveruseRate |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPlan
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPlan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPlan
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPlan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 13:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnnualDiscountPercentage", wireType)
			}
			m.AnnualDiscountPercentage = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AnnualDiscountPercentage |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 14:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PlanPolicy", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPlan
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPlan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.PlanPolicy.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 15:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProjectsLimit", wireType)
			}
			m.ProjectsLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProjectsLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 16:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedBuyers", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPlan
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPlan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllowedBuyers = append(m.AllowedBuyers, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPlan(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPlan
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipPlan(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPlan
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPlan
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthPlan
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPlan
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPlan
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPlan        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPlan          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPlan = fmt.Errorf("proto: unexpected end of group")
)
