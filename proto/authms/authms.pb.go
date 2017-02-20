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
	SMSVerificationResponse
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
	Token    string `protobuf:"bytes,2,opt,name=token" json:"token,omitempty"`
	UserID   int64  `protobuf:"varint,3,opt,name=userID" json:"userID,omitempty"`
	Phone    string `protobuf:"bytes,4,opt,name=phone" json:"phone,omitempty"`
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

func (m *SMSVerificationRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *SMSVerificationRequest) GetUserID() int64 {
	if m != nil {
		return m.UserID
	}
	return 0
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
	UserID   int64  `protobuf:"varint,3,opt,name=userID" json:"userID,omitempty"`
	SmsToken string `protobuf:"bytes,4,opt,name=smsToken" json:"smsToken,omitempty"`
	Code     string `protobuf:"bytes,5,opt,name=code" json:"code,omitempty"`
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

func (m *SMSVerificationCodeRequest) GetUserID() int64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *SMSVerificationCodeRequest) GetSmsToken() string {
	if m != nil {
		return m.SmsToken
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
	Token     string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
	Phone     string `protobuf:"bytes,2,opt,name=phone" json:"phone,omitempty"`
	ExpiresAt string `protobuf:"bytes,3,opt,name=expiresAt" json:"expiresAt,omitempty"`
	Verified  bool   `protobuf:"varint,4,opt,name=verified" json:"verified,omitempty"`
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

func (m *SMSVerificationStatus) GetExpiresAt() string {
	if m != nil {
		return m.ExpiresAt
	}
	return ""
}

func (m *SMSVerificationStatus) GetVerified() bool {
	if m != nil {
		return m.Verified
	}
	return false
}

type SMSVerificationResponse struct {
	Id     string                 `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Code   int32                  `protobuf:"varint,2,opt,name=code" json:"code,omitempty"`
	Status *SMSVerificationStatus `protobuf:"bytes,3,opt,name=status" json:"status,omitempty"`
	Detail string                 `protobuf:"bytes,4,opt,name=detail" json:"detail,omitempty"`
}

func (m *SMSVerificationResponse) Reset()                    { *m = SMSVerificationResponse{} }
func (m *SMSVerificationResponse) String() string            { return proto.CompactTextString(m) }
func (*SMSVerificationResponse) ProtoMessage()               {}
func (*SMSVerificationResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *SMSVerificationResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SMSVerificationResponse) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *SMSVerificationResponse) GetStatus() *SMSVerificationStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *SMSVerificationResponse) GetDetail() string {
	if m != nil {
		return m.Detail
	}
	return ""
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
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

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
	proto.RegisterType((*SMSVerificationResponse)(nil), "SMSVerificationResponse")
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
	VerifyPhone(ctx context.Context, in *SMSVerificationRequest, opts ...client.CallOption) (*SMSVerificationResponse, error)
	VerifyPhoneCode(ctx context.Context, in *SMSVerificationCodeRequest, opts ...client.CallOption) (*SMSVerificationResponse, error)
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

func (c *authMSClient) VerifyPhone(ctx context.Context, in *SMSVerificationRequest, opts ...client.CallOption) (*SMSVerificationResponse, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.VerifyPhone", in)
	out := new(SMSVerificationResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authMSClient) VerifyPhoneCode(ctx context.Context, in *SMSVerificationCodeRequest, opts ...client.CallOption) (*SMSVerificationResponse, error) {
	req := c.c.NewRequest(c.serviceName, "AuthMS.VerifyPhoneCode", in)
	out := new(SMSVerificationResponse)
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
	VerifyPhone(context.Context, *SMSVerificationRequest, *SMSVerificationResponse) error
	VerifyPhoneCode(context.Context, *SMSVerificationCodeRequest, *SMSVerificationResponse) error
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

func (h *AuthMS) VerifyPhone(ctx context.Context, in *SMSVerificationRequest, out *SMSVerificationResponse) error {
	return h.AuthMSHandler.VerifyPhone(ctx, in, out)
}

func (h *AuthMS) VerifyPhoneCode(ctx context.Context, in *SMSVerificationCodeRequest, out *SMSVerificationResponse) error {
	return h.AuthMSHandler.VerifyPhoneCode(ctx, in, out)
}

func init() {
	proto.RegisterFile("github.com/tomogoma/authms/proto/authms/authms.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 817 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x56, 0xdd, 0x6e, 0xda, 0x48,
	0x14, 0x8e, 0x0d, 0x26, 0x70, 0x08, 0xf9, 0x19, 0x65, 0x13, 0x2f, 0x1b, 0xad, 0x90, 0xb5, 0x17,
	0x64, 0x15, 0x0d, 0x52, 0x76, 0x2f, 0x76, 0xf7, 0x8e, 0x2c, 0x91, 0x42, 0xd5, 0x94, 0xca, 0x24,
	0xb9, 0x9f, 0xe0, 0x29, 0x58, 0x09, 0x8c, 0xeb, 0x19, 0x93, 0x70, 0xd1, 0xcb, 0xbe, 0x42, 0x9f,
	0xa6, 0x52, 0xdf, 0xa1, 0x4f, 0x54, 0xcd, 0x8f, 0xcd, 0x4f, 0x80, 0x52, 0xa9, 0x57, 0xf8, 0x3b,
	0x73, 0x7c, 0x7e, 0xbe, 0x73, 0xe6, 0x33, 0xf0, 0x77, 0x3f, 0x14, 0x83, 0xe4, 0x1e, 0xf7, 0xd8,
	0xb0, 0x21, 0xd8, 0x90, 0xf5, 0xd9, 0x90, 0x34, 0x48, 0x22, 0x06, 0x43, 0xde, 0x88, 0x62, 0x26,
	0x58, 0x0a, 0xf4, 0x0f, 0x56, 0x36, 0xef, 0x8b, 0x05, 0xdb, 0x57, 0x21, 0x17, 0x2c, 0x9e, 0xa0,
	0x5d, 0xb0, 0xdb, 0x2d, 0xd7, 0xaa, 0x59, 0xf5, 0x9c, 0x6f, 0xb7, 0x5b, 0xe8, 0x08, 0x0a, 0x09,
	0xa7, 0x71, 0xbb, 0xe5, 0xda, 0xca, 0x66, 0x10, 0x3a, 0x81, 0x52, 0x18, 0x35, 0x83, 0x20, 0xa6,
	0x9c, 0xbb, 0xb9, 0x9a, 0x55, 0x2f, 0xf9, 0x53, 0x03, 0x42, 0x90, 0x0f, 0x88, 0xa0, 0x6e, 0x5e,
	0x1d, 0xa8, 0x67, 0xf4, 0x3b, 0x00, 0xe9, 0xf5, 0x28, 0xe7, 0x37, 0x93, 0x88, 0xba, 0x8e, 0x3a,
	0x99, 0xb1, 0xa0, 0x3f, 0xa0, 0xc2, 0x13, 0x05, 0xbb, 0x82, 0x88, 0x84, 0xbb, 0x85, 0x9a, 0x55,
	0x2f, 0xfa, 0xf3, 0x46, 0x74, 0x08, 0x4e, 0x40, 0xc7, 0xed, 0x96, 0xbb, 0xad, 0x02, 0x68, 0xe0,
	0x3d, 0x81, 0xd3, 0x69, 0x26, 0x62, 0x80, 0x5c, 0xd8, 0x26, 0x51, 0xf4, 0x86, 0x0c, 0xa9, 0xea,
	0xa1, 0xe4, 0xa7, 0x50, 0x16, 0x4c, 0xa2, 0xe8, 0x76, 0xda, 0x4b, 0xc9, 0x9f, 0x1a, 0x50, 0x15,
	0x8a, 0x24, 0x8a, 0x6e, 0xd8, 0x03, 0x1d, 0x99, 0x6e, 0x32, 0x2c, 0xcf, 0xc6, 0x34, 0x0e, 0xdf,
	0x85, 0x34, 0x50, 0x0d, 0x15, 0xfd, 0x0c, 0x7b, 0xff, 0x82, 0x73, 0x47, 0x1e, 0x13, 0x2a, 0xeb,
	0x1a, 0xcb, 0x07, 0x93, 0x56, 0x83, 0xb9, 0x57, 0xed, 0x85, 0x57, 0x3f, 0xdb, 0x90, 0x97, 0xd9,
	0x5f, 0x50, 0x7e, 0x08, 0x8e, 0x50, 0x85, 0xe8, 0x2a, 0x35, 0x90, 0xa1, 0x22, 0xc2, 0xf9, 0x13,
	0x8b, 0x83, 0xb4, 0xc2, 0x14, 0xcb, 0x33, 0x39, 0x16, 0xd5, 0xb6, 0xa6, 0x3c, 0xc3, 0xe8, 0x0c,
	0x76, 0x1e, 0x59, 0x3f, 0x1c, 0x99, 0x01, 0xbb, 0x4e, 0x2d, 0x57, 0x2f, 0x9f, 0x17, 0xb1, 0xc1,
	0xfe, 0xdc, 0x29, 0x3a, 0x01, 0x27, 0x1a, 0xb0, 0x11, 0x55, 0xe4, 0x97, 0xcf, 0x0b, 0x58, 0x75,
	0xe7, 0x6b, 0xa3, 0x3c, 0xa5, 0x43, 0x12, 0x3e, 0x2a, 0xf2, 0x67, 0x4e, 0x95, 0x11, 0x9d, 0x42,
	0x81, 0xc9, 0x21, 0x70, 0xb7, 0xa8, 0x72, 0x1c, 0x60, 0xd9, 0x1e, 0x56, 0x83, 0xe1, 0x97, 0x23,
	0x11, 0x4f, 0x7c, 0xe3, 0x50, 0x6d, 0x42, 0x79, 0xc6, 0x8c, 0xf6, 0x21, 0xf7, 0x40, 0x27, 0x86,
	0x3a, 0xf9, 0x28, 0x33, 0x69, 0x3a, 0x6d, 0x93, 0x49, 0xb9, 0x1b, 0x5a, 0xff, 0xb3, 0xff, 0xb1,
	0xbc, 0x00, 0xf6, 0x2f, 0x08, 0x0f, 0x7b, 0xca, 0x4e, 0xdf, 0x27, 0x94, 0x0b, 0xc9, 0x43, 0x40,
	0xc7, 0x61, 0x8f, 0x1a, 0x3e, 0x4b, 0x7e, 0x86, 0xe5, 0x66, 0xdc, 0x4b, 0xff, 0x6c, 0xfa, 0x29,
	0x5c, 0xc7, 0xac, 0x77, 0x05, 0x7b, 0x3e, 0xed, 0x87, 0x5c, 0xd0, 0x78, 0x93, 0x24, 0xbf, 0x42,
	0x5e, 0x12, 0x6f, 0xaa, 0x76, 0x14, 0x01, 0xbe, 0x32, 0x79, 0x63, 0xa8, 0xdc, 0x46, 0xf2, 0x22,
	0xa4, 0x71, 0xb2, 0x31, 0x5b, 0x0b, 0x63, 0xce, 0xa2, 0xdb, 0x2b, 0xa2, 0xe7, 0x5e, 0x44, 0x9f,
	0xdd, 0xfb, 0xfc, 0xdc, 0xde, 0x7b, 0x57, 0xb0, 0xd3, 0xd9, 0x94, 0xa3, 0x13, 0x73, 0x8d, 0x16,
	0x59, 0x57, 0x3f, 0xde, 0x33, 0x1c, 0x75, 0xaf, 0xbb, 0x77, 0x6a, 0x7f, 0x7b, 0x44, 0x84, 0x6c,
	0xb4, 0x49, 0xcc, 0xe5, 0xdb, 0x3c, 0x95, 0x95, 0xdc, 0x9c, 0xac, 0x1c, 0xa6, 0xfb, 0xa7, 0xbb,
	0xd0, 0xc0, 0xfb, 0x64, 0x41, 0x75, 0x21, 0xf5, 0xff, 0x2c, 0xa0, 0x3f, 0x3f, 0x7d, 0x15, 0x8a,
	0x7c, 0xc8, 0xb5, 0x0c, 0x98, 0x8b, 0x94, 0x62, 0xa9, 0x69, 0x3d, 0x16, 0xa4, 0xca, 0xa5, 0x9e,
	0xbd, 0x0f, 0xf0, 0xcb, 0x42, 0x5d, 0x53, 0x99, 0x5a, 0x32, 0xdc, 0xac, 0x3b, 0x7b, 0xa6, 0x3b,
	0xa9, 0x4c, 0xf4, 0x39, 0x0a, 0x63, 0xca, 0x9b, 0x22, 0x95, 0xd2, 0xcc, 0xb0, 0x56, 0x7d, 0x3e,
	0x5a, 0x70, 0xfc, 0x62, 0x24, 0x3c, 0x62, 0x23, 0x4e, 0xa5, 0xaa, 0x84, 0x81, 0x49, 0x6f, 0x87,
	0x41, 0x56, 0xbe, 0x4c, 0xed, 0xe8, 0xf2, 0x11, 0x86, 0x02, 0xd7, 0x5a, 0xab, 0x57, 0xea, 0x08,
	0x2f, 0xed, 0xc6, 0x37, 0x5e, 0x92, 0xb6, 0x80, 0x0a, 0x29, 0x00, 0x9a, 0x1c, 0x83, 0x3c, 0x02,
	0xc5, 0x1f, 0xca, 0xbb, 0x66, 0x91, 0x57, 0xa4, 0x38, 0xff, 0x9a, 0x83, 0x82, 0xdc, 0xc2, 0xeb,
	0x2e, 0x3a, 0x95, 0xd9, 0xf4, 0x9d, 0x44, 0xfb, 0x78, 0xe1, 0x7a, 0x56, 0x4b, 0x38, 0x2d, 0xc5,
	0xdb, 0x42, 0x0d, 0xa8, 0xbc, 0x96, 0xf2, 0x76, 0x9b, 0xaa, 0xe1, 0x01, 0x5e, 0x14, 0x8d, 0xf9,
	0x17, 0xce, 0x00, 0xd4, 0x0b, 0x97, 0x4a, 0xd1, 0x36, 0xf5, 0x7e, 0xab, 0xe6, 0xf8, 0x3d, 0xef,
	0xba, 0xf1, 0xd6, 0x5f, 0xaa, 0x0a, 0xee, 0xac, 0xf4, 0xfc, 0x13, 0xca, 0x5a, 0x2b, 0x74, 0xe0,
	0x5d, 0x3c, 0xa7, 0x1c, 0x2b, 0x7c, 0x3b, 0xf2, 0x9b, 0xbe, 0xde, 0xf7, 0x02, 0xca, 0x6a, 0xba,
	0x13, 0x1d, 0xf7, 0x18, 0x2f, 0xbf, 0xcf, 0x55, 0x17, 0xaf, 0xd8, 0x2a, 0x6f, 0x0b, 0xbd, 0x82,
	0xbd, 0x99, 0x18, 0xf2, 0x1a, 0xa2, 0xdf, 0xf0, 0xea, 0xcb, 0xb9, 0x2e, 0xd6, 0x7d, 0x41, 0xfd,
	0xff, 0xf8, 0xeb, 0x5b, 0x00, 0x00, 0x00, 0xff, 0xff, 0x25, 0xd1, 0x2b, 0x22, 0xb7, 0x08, 0x00,
	0x00,
}
