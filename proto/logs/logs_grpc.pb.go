// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package logs

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LogsServiceV1Client is the client API for LogsServiceV1 service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogsServiceV1Client interface {
	LogV1(ctx context.Context, in *LogRequestV1, opts ...grpc.CallOption) (*LogResponseV1, error)
}

type logsServiceV1Client struct {
	cc grpc.ClientConnInterface
}

func NewLogsServiceV1Client(cc grpc.ClientConnInterface) LogsServiceV1Client {
	return &logsServiceV1Client{cc}
}

func (c *logsServiceV1Client) LogV1(ctx context.Context, in *LogRequestV1, opts ...grpc.CallOption) (*LogResponseV1, error) {
	out := new(LogResponseV1)
	err := c.cc.Invoke(ctx, "/logs.LogsServiceV1/LogV1", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogsServiceV1Server is the server API for LogsServiceV1 service.
// All implementations must embed UnimplementedLogsServiceV1Server
// for forward compatibility
type LogsServiceV1Server interface {
	LogV1(context.Context, *LogRequestV1) (*LogResponseV1, error)
	mustEmbedUnimplementedLogsServiceV1Server()
}

// UnimplementedLogsServiceV1Server must be embedded to have forward compatible implementations.
type UnimplementedLogsServiceV1Server struct {
}

func (UnimplementedLogsServiceV1Server) LogV1(context.Context, *LogRequestV1) (*LogResponseV1, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogV1 not implemented")
}
func (UnimplementedLogsServiceV1Server) mustEmbedUnimplementedLogsServiceV1Server() {}

// UnsafeLogsServiceV1Server may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogsServiceV1Server will
// result in compilation errors.
type UnsafeLogsServiceV1Server interface {
	mustEmbedUnimplementedLogsServiceV1Server()
}

func RegisterLogsServiceV1Server(s grpc.ServiceRegistrar, srv LogsServiceV1Server) {
	s.RegisterService(&LogsServiceV1_ServiceDesc, srv)
}

func _LogsServiceV1_LogV1_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogRequestV1)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogsServiceV1Server).LogV1(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logs.LogsServiceV1/LogV1",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogsServiceV1Server).LogV1(ctx, req.(*LogRequestV1))
	}
	return interceptor(ctx, in, info, handler)
}

// LogsServiceV1_ServiceDesc is the grpc.ServiceDesc for LogsServiceV1 service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LogsServiceV1_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logs.LogsServiceV1",
	HandlerType: (*LogsServiceV1Server)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LogV1",
			Handler:    _LogsServiceV1_LogV1_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "logs.proto",
}