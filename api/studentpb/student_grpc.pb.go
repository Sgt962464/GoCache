// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.5
// source: student.proto

package __

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	StudentService_StudentCreate_FullMethodName = "/studentpb.StudentService/StudentCreate"
	StudentService_StudentDelete_FullMethodName = "/studentpb.StudentService/StudentDelete"
	StudentService_StudentUpdate_FullMethodName = "/studentpb.StudentService/StudentUpdate"
	StudentService_StudentShow_FullMethodName   = "/studentpb.StudentService/StudentShow"
)

// StudentServiceClient is the client API for StudentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StudentServiceClient interface {
	StudentCreate(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentCommonResponse, error)
	StudentDelete(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentCommonResponse, error)
	StudentUpdate(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentCommonResponse, error)
	StudentShow(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentDetailResponse, error)
}

type studentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewStudentServiceClient(cc grpc.ClientConnInterface) StudentServiceClient {
	return &studentServiceClient{cc}
}

func (c *studentServiceClient) StudentCreate(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentCommonResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StudentCommonResponse)
	err := c.cc.Invoke(ctx, StudentService_StudentCreate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentServiceClient) StudentDelete(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentCommonResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StudentCommonResponse)
	err := c.cc.Invoke(ctx, StudentService_StudentDelete_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentServiceClient) StudentUpdate(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentCommonResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StudentCommonResponse)
	err := c.cc.Invoke(ctx, StudentService_StudentUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentServiceClient) StudentShow(ctx context.Context, in *StudentRequest, opts ...grpc.CallOption) (*StudentDetailResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StudentDetailResponse)
	err := c.cc.Invoke(ctx, StudentService_StudentShow_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StudentServiceServer is the server API for StudentService service.
// All implementations must embed UnimplementedStudentServiceServer
// for forward compatibility.
type StudentServiceServer interface {
	StudentCreate(context.Context, *StudentRequest) (*StudentCommonResponse, error)
	StudentDelete(context.Context, *StudentRequest) (*StudentCommonResponse, error)
	StudentUpdate(context.Context, *StudentRequest) (*StudentCommonResponse, error)
	StudentShow(context.Context, *StudentRequest) (*StudentDetailResponse, error)
	mustEmbedUnimplementedStudentServiceServer()
}

// UnimplementedStudentServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStudentServiceServer struct{}

func (UnimplementedStudentServiceServer) StudentCreate(context.Context, *StudentRequest) (*StudentCommonResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StudentCreate not implemented")
}
func (UnimplementedStudentServiceServer) StudentDelete(context.Context, *StudentRequest) (*StudentCommonResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StudentDelete not implemented")
}
func (UnimplementedStudentServiceServer) StudentUpdate(context.Context, *StudentRequest) (*StudentCommonResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StudentUpdate not implemented")
}
func (UnimplementedStudentServiceServer) StudentShow(context.Context, *StudentRequest) (*StudentDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StudentShow not implemented")
}
func (UnimplementedStudentServiceServer) mustEmbedUnimplementedStudentServiceServer() {}
func (UnimplementedStudentServiceServer) testEmbeddedByValue()                        {}

// UnsafeStudentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StudentServiceServer will
// result in compilation errors.
type UnsafeStudentServiceServer interface {
	mustEmbedUnimplementedStudentServiceServer()
}

func RegisterStudentServiceServer(s grpc.ServiceRegistrar, srv StudentServiceServer) {
	// If the following call pancis, it indicates UnimplementedStudentServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&StudentService_ServiceDesc, srv)
}

func _StudentService_StudentCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StudentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServiceServer).StudentCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StudentService_StudentCreate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServiceServer).StudentCreate(ctx, req.(*StudentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudentService_StudentDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StudentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServiceServer).StudentDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StudentService_StudentDelete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServiceServer).StudentDelete(ctx, req.(*StudentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudentService_StudentUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StudentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServiceServer).StudentUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StudentService_StudentUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServiceServer).StudentUpdate(ctx, req.(*StudentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudentService_StudentShow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StudentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServiceServer).StudentShow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StudentService_StudentShow_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServiceServer).StudentShow(ctx, req.(*StudentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// StudentService_ServiceDesc is the grpc.ServiceDesc for StudentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StudentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "studentpb.StudentService",
	HandlerType: (*StudentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StudentCreate",
			Handler:    _StudentService_StudentCreate_Handler,
		},
		{
			MethodName: "StudentDelete",
			Handler:    _StudentService_StudentDelete_Handler,
		},
		{
			MethodName: "StudentUpdate",
			Handler:    _StudentService_StudentUpdate_Handler,
		},
		{
			MethodName: "StudentShow",
			Handler:    _StudentService_StudentShow_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "student.proto",
}
