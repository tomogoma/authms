// Code generated by protoc-gen-go.
// source: github.com/tomogoma/authms/proto/authms/authms.proto
// DO NOT EDIT!

/*
Package authms is a generated protocol buffer package.

It is generated from these files:
	github.com/tomogoma/authms/proto/authms/authms.proto

It has these top-level messages:
	History
	OAuth
	Value
	User
	BasicAuthRequest
	RegisterRequest
	UpdateRequest
	OAuthRequest
	SMSVerificationRequest
	SMSVerificationCodeRequest
	SMSVerificationStatus
	Response
*/
package authms

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
	context "golang.org/x/net/context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type History struct {
	ID            int64  `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	UserID        int64  `protobuf:"varint,2,opt,name=userID" json:"userID,omitempty"`
	IpAddress     string `protobuf:"bytes,3,opt,name=ipAddress" json:"ipAddress,omitempty"`
	Date          string `protobuf:"bytes,4,opt,name=date" json:"date,omitempty"`
	AccessType    string `protobuf:"bytes,5,opt,name=accessType" json:"accessType,omitempty"`
	SuccessStatus bool   `protobuf:"varint,6,opt,name=successStatus" json:"successStatus,omitempty"`
	DevID         string `protobuf:"bytes,7,opt,name=devID" json:"devID,omitempty"`
}

func (m *History) Reset()                    { *m = History{} }
func (m *History) String() string            { return proto.CompactTextString(m) }
func (*History) ProtoMessage()               {}
func (*History) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *History) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *History) GetUserID() int64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *History) GetIpAddress() string {
	if m != nil {
		return m.IpAddress
	}
	return ""
}

func (m *History) GetDate() string {
	if m != nil {
		return m.Date
	}
	return ""
}

func (m *History) GetAccessType() string {
	if m != nil {
		return m.AccessType
	}
	return ""
}

func (m *History) GetSuccessStatus() bool {
	if m != nil {
		return m.SuccessStatus
	}
	return false
}

func (m *History) GetDevID() string {
	if m != nil {
		return m.DevID
	}
	return ""
}

type OAuth struct {
	AppName   string `protobuf:"bytes,1,opt,name=appName" json:"appName,omitempty"`
	AppUserID string `protobuf:"bytes,2,opt,name=appUserID" json:"appUserID,omitempty"`
	AppToken  string `protobuf:"bytes,3,opt,name=appToken" json:"appToken,omitempty"`
	Verified  bool   `protobuf:"varint,4,opt,name=verified" json:"verified,omitempty"`
}

func (m *OAuth) Reset()                    { *m = OAuth{} }
func (m *OAuth) String() string            { return proto.CompactTextString(m) }
func (*OAuth) ProtoMessage()               {}
func (*OAuth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *OAuth) GetAppName() string {
	if m != nil {
		return m.AppName
	}
	return ""
}

func (m *OAuth) GetAppUserID() string {
	if m != nil {
		return m.AppUserID
	}
	return ""
}

func (m *OAuth) GetAppToken() string {
	if m != nil {
		return m.AppToken
	}
	return ""
}

func (m *OAuth) GetVerified() bool {
	if m != nil {
		return m.Verified
	}
	return false
}

type Value struct {
	Value    string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
	Verified bool   `protobuf:"varint,2,opt,name=verified" json:"verified,omitempty"`
}

func (m *Value) Reset()                    { *m = Value{} }
func (m *Value) String() string            { return proto.CompactTextString(m) }
func (*Value) ProtoMessage()               {}
func (*Value) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Value) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *Value) GetVerified() bool {
	if m != nil {
		return m.Verified
	}
	return false
}

type User struct {
	ID           int64             `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	Token        string            `protobuf:"bytes,2,opt,name=token" json:"token,omitempty"`
	Password     string            `protobuf:"bytes,3,opt,name=password" json:"password,omitempty"`
	UserName     string            `protobuf:"bytes,4,opt,name=userName" json:"userName,omitempty"`
	LoginHistory []*History        `protobuf:"bytes,5,rep,name=loginHistory" json:"loginHistory,omitempty"`
	Phone        *Value            `protobuf:"bytes,6,opt,name=phone" json:"phone,omitempty"`
	Email        *Value            `protobuf:"bytes,7,opt,name=email" json:"email,omitempty"`
	OAuths       map[string]*OAuth `protobuf:"bytes,8,rep,name=oAuths" json:"oAuths,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *User) Reset()                    { *m = User{} }
func (m *User) String() string            { return proto.CompactTextString(m) }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *User) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *User) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *User) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *User) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *User) GetLoginHistory() []*History {
	if m != nil {
		return m.LoginHistory
	}
	return nil
}

func (m *User) GetPhone() *Value {
	if m != nil {
		return m.Phone
	}
	return nil
}

func (m *User) GetEmail() *Value {
	if m != nil {
		return m.Email
	}
	return nil
}

func (m *User) GetOAuths() map[string]*OAuth {
	if m != nil {
		return m.OAuths
	}
	return nil
}

type BasicAuthRequest struct {
	DeviceID string `protobuf:"bytes,1,opt,name=deviceID" json:"deviceID,omitempty"`
	BasicID  string `protobuf:"bytes,2,opt,name=basicID" json:"basicID,omitempty"`
	Password string `protobuf:"bytes,3,opt,name=password" json:"password,omitempty"`
}

func (m *BasicAuthRequest) Reset()                    { *m = BasicAuthRequest{} }
func (m *BasicAuthRequest) String() string            { return proto.CompactTextString(m) }
func (*BasicAuthRequest) ProtoMessage()               {}
func (*BasicAuthRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *BasicAuthRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *BasicAuthRequest) GetBasicID() string {
	if m != nil {
		return m.BasicID
	}
	return ""
}

func (m *BasicAuthRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type RegisterRequest struct {
	DeviceID string `protobuf:"bytes,1,opt,name=deviceID" json:"deviceID,omitempty"`
	User     *User  `protobuf:"bytes,2,opt,name=user" json:"user,omitempty"`
}

func (m *RegisterRequest) Reset()                    { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string            { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()               {}
func (*RegisterRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *RegisterRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *RegisterRequest) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

type UpdateRequest struct {
	Token    string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
	DeviceID string `protobuf:"bytes,2,opt,name=deviceID" json:"deviceID,omitempty"`
	User     *User  `protobuf:"bytes,3,opt,name=user" json:"user,omitempty"`
	AppName  string `protobuf:"bytes,4,opt,name=appName" json:"appName,omitempty"`
}

func (m *UpdateRequest) Reset()                    { *m = UpdateRequest{} }
func (m *UpdateRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateRequest) ProtoMessage()               {}
func (*UpdateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *UpdateRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *UpdateRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *UpdateRequest) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *UpdateRequest) GetAppName() string {
	if m != nil {
		return m.AppName
	}
	return ""
}

type OAuthRequest struct {
	DeviceID string `protobuf:"bytes,1,opt,name=deviceID" json:"deviceID,omitempty"`
	OAuth    *OAuth `protobuf:"bytes,2,opt,name=OAuth" json:"OAuth,omitempty"`
}

func (m *OAuthRequest) Reset()                    { *m = OAuthRequest{} }
func (m *OAuthRequest) String() string            { return proto.CompactTextString(m) }
func (*OAuthRequest) ProtoMessage()               {}
func (*OAuthRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *OAuthRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *OAuthRequest) GetOAuth() *OAuth {
	if m != nil {
		return m.OAuth
	}
	return nil
}

type SMSVerificationRequest struct {
	DeviceID string `protobuf:"bytes,1,opt,name=deviceID" json:"deviceID,omitempty"`
	Phone    string `protobuf:"bytes,2,opt,name=phone" json:"phone,omitempty"`
}

func (m *SMSVerificationRequest) Reset()                    { *m = SMSVerificationRequest{} }
func (m *SMSVerificationRequest) String() string            { return proto.CompactTextString(m) }
func (*SMSVerificationRequest) ProtoMessage()               {}
func (*SMSVerificationRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *SMSVerificationRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *SMSVerificationRequest) GetPhone() string {
	if m != nil {
		return m.Phone
	}
	return ""
}

type SMSVerificationCodeRequest struct {
	DeviceID string `protobuf:"bytes,1,opt,name=deviceID" json:"deviceID,omitempty"`
	Token    string `protobuf:"bytes,2,opt,name=token" json:"token,omitempty"`
	Code     string `protobuf:"bytes,3,opt,name=code" json:"code,omitempty"`
}

func (m *SMSVerificationCodeRequest) Reset()                    { *m = SMSVerificationCodeRequest{} }
func (m *SMSVerificationCodeRequest) String() string            { return proto.CompactTextString(m) }
func (*SMSVerificationCodeRequest) ProtoMessage()               {}
func (*SMSVerificationCodeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *SMSVerificationCodeRequest) GetDeviceID() string {
	if m != nil {
		return m.DeviceID
	}
	return ""
}

func (m *SMSVerificationCodeRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *SMSVerificationCodeRequest) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

type SMSVerificationStatus struct {
	Token        string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
	Phone        string `protobuf:"bytes,2,opt,name=phone" json:"phone,omitempty"`
	Retries      int32  `protobuf:"varint,3,opt,name=retries" json:"retries,omitempty"`
	ExpiresAt    string `protobuf:"bytes,4,opt,name=expiresAt" json:"expiresAt,omitempty"`
	IsBlocked    bool   `protobuf:"varint,5,opt,name=isBlocked" json:"isBlocked,omitempty"`
	BlockedUntil string `protobuf:"bytes,6,opt,name=blockedUntil" json:"blockedUntil,omitempty"`
	Verified     bool   `protobuf:"varint,8,opt,name=verified" json:"verified,omitempty"`
}

func (m *SMSVerificationStatus) Reset()                    { *m = SMSVerificationStatus{} }
func (m *SMSVerificationStatus) String() string            { return proto.CompactTextString(m) }
func (*SMSVerificationStatus) ProtoMessage()               {}
func (*SMSVerificationStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *SMSVerificationStatus) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *SMSVerificationStatus) GetPhone() string {
	if m != nil {
		return m.Phone
	}
	return ""
}

func (m *SMSVerificationStatus) GetRetries() int32 {
	if m != nil {
		return m.Retries
	}
	return 0
}

func (m *SMSVerificationStatus) GetExpiresAt() string {
	if m != nil {
		return m.ExpiresAt
	}
	return ""
}

func (m *SMSVerificationStatus) GetIsBlocked() bool {
	if m != nil {
		return m.IsBlocked
	}
	return false
}

func (m *SMSVerificationStatus) GetBlockedUntil() string {
	if m != nil {
		return m.BlockedUntil
	}
	return ""
}

func (m *SMSVerificationStatus) GetVerified() bool {
	if m != nil {
		return m.Verified
	}
	return false
}

type Response struct {
	Id     string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Code   int32  `protobuf:"varint,2,opt,name=code" json:"code,omitempty"`
	User   *User  `protobuf:"bytes,3,opt,name=user" json:"user,omitempty"`
	Detail string `protobuf:"bytes,4,opt,name=detail" json:"detail,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *Response) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Response) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *Response) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *Response) GetDetail() string {
	if m != nil {
		return m.Detail
	}
	return ""
}

func init() {
	proto.RegisterType((*History)(nil), "History")
	proto.RegisterType((*OAuth)(nil), "OAuth")
	proto.RegisterType((*Value)(nil), "Value")
	proto.RegisterType((*User)(nil), "User")
	proto.RegisterType((*BasicAuthRequest)(nil), "BasicAuthRequest")
	proto.RegisterType((*RegisterRequest)(nil), "RegisterRequest")
	proto.RegisterType((*UpdateRequest)(nil), "UpdateRequest")
	proto.RegisterType((*OAuthRequest)(nil), "OAuthRequest")
	proto.RegisterType((*SMSVerificationRequest)(nil), "SMSVerificationRequest")
	proto.RegisterType((*SMSVerificationCodeRequest)(nil), "SMSVerificationCodeRequest")
	proto.RegisterType((*SMSVerificationStatus)(nil), "SMSVerificationStatus")
	proto.RegisterType((*Response)(nil), "Response")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Publisher API

type Publisher interface {
	Publish(ctx context.Context, msg interface{}, opts ...client.PublishOption) error
}

type publisher struct {
	c     client.Client
	topic string
}

func (p *publisher) Publish(ctx context.Context, msg interface{}, opts ...client.PublishOption) error {
	return p.c.Publish(ctx, p.c.NewPublication(p.topic, msg), opts...)
}

func NewPublisher(topic string, c client.Client) Publisher {
	if c == nil {
		c = client.NewClient()
	}
	return &publisher{c, topic}
}

// Subscriber API

func RegisterSubscriber(topic string, s server.Server, h interface{}, opts ...server.SubscriberOption) error {
	return s.Subscribe(s.NewSubscriber(topic, h, opts...))
}

// Client API for AuthMS service

type AuthMSClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...client.CallOption) (*Response, error)
	LoginUserName(ctx context.Context, in *BasicAuthRequest, opts ...client.CallOption) (*Response, error)
	LoginEmail(ctx context.Context, in *BasicAuthRequest, opts ...client.CallOption) (*Response, error)
	LoginPhone(ctx context.Context, in *BasicAuthRequest, opts ...client.CallOption) (*Response, error)
	LoginOAuth(ctx context.Context, in *OAuthRequest, opts ...client.CallOption) (*Response, error)
	UpdatePhone(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*Response, error)
	UpdateOauth(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*Response, error)
}

type authMSClient struct {
	c           client.Client
	serviceName string
}

func NewAuthMSClient(serviceName string, c client.Client) AuthMSClient {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "authms"
	}
	return &authMSClient{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *authMSClient) Register(ctx context.Context, in *RegisterRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.Register", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) LoginUserName(ctx context.Context, in *BasicAuthRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.LoginUserName", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) LoginEmail(ctx context.Context, in *BasicAuthRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.LoginEmail", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) LoginPhone(ctx context.Context, in *BasicAuthRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.LoginPhone", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) LoginOAuth(ctx context.Context, in *OAuthRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.LoginOAuth", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) UpdatePhone(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.UpdatePhone", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) UpdateOauth(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.UpdateOauth", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for AuthMS service

type AuthMSHandler interface {
	Register(context.Context, *RegisterRequest, *Response) error
	LoginUserName(context.Context, *BasicAuthRequest, *Response) error
	LoginEmail(context.Context, *BasicAuthRequest, *Response) error
	LoginPhone(context.Context, *BasicAuthRequest, *Response) error
	LoginOAuth(context.Context, *OAuthRequest, *Response) error
	UpdatePhone(context.Context, *UpdateRequest, *Response) error
	UpdateOauth(context.Context, *UpdateRequest, *Response) error
}

func RegisterAuthMSHandler(s server.Server, hdlr AuthMSHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&AuthMS{hdlr}, opts...))
}

type AuthMS struct {
	AuthMSHandler
}

func (h *AuthMS) Register(ctx context.Context, in *RegisterRequest, out *Response) error {
	return h.AuthMSHandler.Register(ctx, in, out)
}

func (h *AuthMS) LoginUserName(ctx context.Context, in *BasicAuthRequest, out *Response) error {
	return h.AuthMSHandler.LoginUserName(ctx, in, out)
}

func (h *AuthMS) LoginEmail(ctx context.Context, in *BasicAuthRequest, out *Response) error {
	return h.AuthMSHandler.LoginEmail(ctx, in, out)
}

func (h *AuthMS) LoginPhone(ctx context.Context, in *BasicAuthRequest, out *Response) error {
	return h.AuthMSHandler.LoginPhone(ctx, in, out)
}

func (h *AuthMS) LoginOAuth(ctx context.Context, in *OAuthRequest, out *Response) error {
	return h.AuthMSHandler.LoginOAuth(ctx, in, out)
}

func (h *AuthMS) UpdatePhone(ctx context.Context, in *UpdateRequest, out *Response) error {
	return h.AuthMSHandler.UpdatePhone(ctx, in, out)
}

func (h *AuthMS) UpdateOauth(ctx context.Context, in *UpdateRequest, out *Response) error {
	return h.AuthMSHandler.UpdateOauth(ctx, in, out)
}

func init() {
	proto.RegisterFile("github.com/tomogoma/authms/proto/authms/authms.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 774 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x55, 0xdd, 0x6e, 0xe2, 0x46,
	0x14, 0xae, 0x0d, 0x26, 0xe6, 0x10, 0xb6, 0xd9, 0x51, 0x1a, 0xb9, 0x28, 0xaa, 0x90, 0xd5, 0x0b,
	0xb6, 0x5a, 0x19, 0x29, 0xed, 0x45, 0xdb, 0x3b, 0xb6, 0xac, 0x14, 0xaa, 0x6e, 0xa9, 0x86, 0x90,
	0x7b, 0x63, 0x4f, 0x61, 0x14, 0x60, 0x5c, 0xcf, 0x98, 0x94, 0x77, 0xab, 0x54, 0xa9, 0x6f, 0xd1,
	0xb7, 0xa9, 0xe6, 0x78, 0x6c, 0xb0, 0xf3, 0x53, 0xae, 0x3c, 0xdf, 0x99, 0x33, 0xe7, 0xe7, 0x3b,
	0x3f, 0x86, 0xef, 0x96, 0x5c, 0xad, 0xb2, 0x45, 0x10, 0x89, 0xcd, 0x50, 0x89, 0x8d, 0x58, 0x8a,
	0x4d, 0x38, 0x0c, 0x33, 0xb5, 0xda, 0xc8, 0x61, 0x92, 0x0a, 0x25, 0x0a, 0x90, 0x7f, 0x02, 0x94,
	0xf9, 0x7f, 0x5b, 0x70, 0x76, 0xcb, 0xa5, 0x12, 0xe9, 0x9e, 0xbc, 0x01, 0x7b, 0x32, 0xf6, 0xac,
	0xbe, 0x35, 0x68, 0x50, 0x7b, 0x32, 0x26, 0x57, 0xd0, 0xca, 0x24, 0x4b, 0x27, 0x63, 0xcf, 0x46,
	0x99, 0x41, 0xe4, 0x1a, 0xda, 0x3c, 0x19, 0xc5, 0x71, 0xca, 0xa4, 0xf4, 0x1a, 0x7d, 0x6b, 0xd0,
	0xa6, 0x07, 0x01, 0x21, 0xd0, 0x8c, 0x43, 0xc5, 0xbc, 0x26, 0x5e, 0xe0, 0x99, 0x7c, 0x05, 0x10,
	0x46, 0x11, 0x93, 0xf2, 0x6e, 0x9f, 0x30, 0xcf, 0xc1, 0x9b, 0x23, 0x09, 0xf9, 0x1a, 0xba, 0x32,
	0x43, 0x38, 0x53, 0xa1, 0xca, 0xa4, 0xd7, 0xea, 0x5b, 0x03, 0x97, 0x56, 0x85, 0xe4, 0x12, 0x9c,
	0x98, 0xed, 0x26, 0x63, 0xef, 0x0c, 0x0d, 0xe4, 0xc0, 0x7f, 0x04, 0x67, 0x3a, 0xca, 0xd4, 0x8a,
	0x78, 0x70, 0x16, 0x26, 0xc9, 0xaf, 0xe1, 0x86, 0x61, 0x0e, 0x6d, 0x5a, 0x40, 0x1d, 0x70, 0x98,
	0x24, 0xf3, 0x43, 0x2e, 0x6d, 0x7a, 0x10, 0x90, 0x1e, 0xb8, 0x61, 0x92, 0xdc, 0x89, 0x07, 0xb6,
	0x35, 0xd9, 0x94, 0x58, 0xdf, 0xed, 0x58, 0xca, 0x7f, 0xe7, 0x2c, 0xc6, 0x84, 0x5c, 0x5a, 0x62,
	0xff, 0x07, 0x70, 0xee, 0xc3, 0x75, 0xc6, 0x74, 0x5c, 0x3b, 0x7d, 0x30, 0x6e, 0x73, 0x50, 0x79,
	0x6a, 0xd7, 0x9e, 0xfe, 0x65, 0x43, 0x53, 0x7b, 0x7f, 0x42, 0xf9, 0x25, 0x38, 0x0a, 0x03, 0xc9,
	0xa3, 0xcc, 0x81, 0x36, 0x95, 0x84, 0x52, 0x3e, 0x8a, 0x34, 0x2e, 0x22, 0x2c, 0xb0, 0xbe, 0xd3,
	0x65, 0xc1, 0xb4, 0x73, 0xca, 0x4b, 0x4c, 0xde, 0xc3, 0xf9, 0x5a, 0x2c, 0xf9, 0xd6, 0x14, 0xd8,
	0x73, 0xfa, 0x8d, 0x41, 0xe7, 0xc6, 0x0d, 0x0c, 0xa6, 0x95, 0x5b, 0x72, 0x0d, 0x4e, 0xb2, 0x12,
	0x5b, 0x86, 0xe4, 0x77, 0x6e, 0x5a, 0x01, 0x66, 0x47, 0x73, 0xa1, 0xbe, 0x65, 0x9b, 0x90, 0xaf,
	0x91, 0xfc, 0xa3, 0x5b, 0x14, 0x92, 0x77, 0xd0, 0x12, 0xba, 0x08, 0xd2, 0x73, 0xd1, 0xc7, 0xdb,
	0x40, 0xa7, 0x17, 0x60, 0x61, 0xe4, 0xc7, 0xad, 0x4a, 0xf7, 0xd4, 0x28, 0xf4, 0x46, 0xd0, 0x39,
	0x12, 0x93, 0x0b, 0x68, 0x3c, 0xb0, 0xbd, 0xa1, 0x4e, 0x1f, 0xb5, 0xa7, 0x9c, 0x4e, 0xdb, 0x78,
	0x42, 0x75, 0x43, 0xeb, 0x8f, 0xf6, 0xf7, 0x96, 0x1f, 0xc3, 0xc5, 0x87, 0x50, 0xf2, 0x08, 0xe5,
	0xec, 0x8f, 0x8c, 0x49, 0xa5, 0x79, 0x88, 0xd9, 0x8e, 0x47, 0xcc, 0xf0, 0xd9, 0xa6, 0x25, 0xd6,
	0x9d, 0xb1, 0xd0, 0xfa, 0x65, 0xf5, 0x0b, 0xf8, 0x1a, 0xb3, 0xfe, 0x2d, 0x7c, 0x4e, 0xd9, 0x92,
	0x4b, 0xc5, 0xd2, 0x53, 0x9c, 0x7c, 0x09, 0x4d, 0x4d, 0xbc, 0x89, 0xda, 0x41, 0x02, 0x28, 0x8a,
	0xfc, 0x1d, 0x74, 0xe7, 0x89, 0x1e, 0x84, 0xc2, 0x4e, 0x59, 0x66, 0xab, 0x56, 0xe6, 0xd2, 0xba,
	0xfd, 0x82, 0xf5, 0xc6, 0x13, 0xeb, 0xc7, 0x7d, 0xdf, 0xac, 0xf4, 0xbd, 0x7f, 0x0b, 0xe7, 0xd3,
	0x53, 0x39, 0xba, 0x36, 0x63, 0x54, 0x67, 0x1d, 0x3f, 0xfe, 0xcf, 0x70, 0x35, 0xfb, 0x34, 0xbb,
	0xc7, 0xfe, 0x8d, 0x42, 0xc5, 0xc5, 0xf6, 0x14, 0x9b, 0x97, 0x45, 0x47, 0x99, 0x6e, 0x46, 0xe0,
	0x2f, 0xa0, 0x57, 0xb3, 0xf5, 0x93, 0x88, 0xd9, 0x89, 0xf6, 0x9e, 0x99, 0x0e, 0x02, 0xcd, 0x48,
	0xc4, 0xcc, 0xd4, 0x0f, 0xcf, 0xfe, 0xbf, 0x16, 0x7c, 0x51, 0x73, 0x72, 0x58, 0x22, 0xcf, 0x50,
	0xff, 0x6c, 0xa4, 0x9a, 0xd9, 0x94, 0xa9, 0x94, 0xb3, 0x7c, 0xcd, 0x39, 0xb4, 0x80, 0x7a, 0xa3,
	0xb0, 0x3f, 0x13, 0x9e, 0x32, 0x39, 0x52, 0x86, 0xf5, 0x83, 0x00, 0x17, 0xa4, 0xfc, 0xb0, 0x16,
	0xd1, 0x03, 0x8b, 0x71, 0xdb, 0xb9, 0xf4, 0x20, 0x20, 0x3e, 0x9c, 0x2f, 0xf2, 0xe3, 0x7c, 0xab,
	0xf8, 0x1a, 0xc7, 0xad, 0x4d, 0x2b, 0xb2, 0xca, 0xf2, 0x70, 0x6b, 0xcb, 0x23, 0x04, 0x97, 0x32,
	0x99, 0x88, 0xad, 0x64, 0x7a, 0x7f, 0xf0, 0xd8, 0xa4, 0x62, 0xf3, 0xb8, 0xe4, 0xc2, 0xc6, 0x70,
	0xf1, 0xfc, 0x5a, 0xeb, 0x5c, 0x41, 0x2b, 0x66, 0x4a, 0x4f, 0x75, 0x9e, 0x83, 0x41, 0x37, 0xff,
	0xd8, 0xd0, 0xd2, 0x75, 0xff, 0x34, 0x23, 0xef, 0xb4, 0xb7, 0x7c, 0x0a, 0xc8, 0x45, 0x50, 0x1b,
	0x88, 0x5e, 0x3b, 0x28, 0x42, 0xf1, 0x3f, 0x23, 0x43, 0xe8, 0xfe, 0xa2, 0x17, 0xca, 0xbc, 0xd8,
	0x3f, 0x6f, 0x83, 0xfa, 0x98, 0x56, 0x1f, 0xbc, 0x07, 0xc0, 0x07, 0x1f, 0x71, 0x87, 0x9c, 0xaa,
	0xfd, 0x1b, 0xd6, 0xe6, 0xff, 0xb4, 0x07, 0x46, 0x3b, 0xff, 0x37, 0x74, 0x83, 0xe9, 0x8b, 0x9a,
	0xdf, 0x40, 0x27, 0x9f, 0xce, 0xdc, 0xf0, 0x9b, 0xa0, 0x32, 0xab, 0x2f, 0xe8, 0x4e, 0xf5, 0x5f,
	0xf4, 0x55, 0xdd, 0x45, 0x0b, 0xff, 0xb0, 0xdf, 0xfe, 0x17, 0x00, 0x00, 0xff, 0xff, 0x5d, 0xa7,
	0x98, 0x04, 0x99, 0x07, 0x00, 0x00,
}
