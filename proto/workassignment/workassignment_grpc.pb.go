// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.31.1
// source: proto/workassignment/workassignment.proto

package workassignment

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
	WorkAssignmentService_GetWorksForAssistant_FullMethodName = "/workassignment.WorkAssignmentService/GetWorksForAssistant"
	WorkAssignmentService_GetWorkDetails_FullMethodName       = "/workassignment.WorkAssignmentService/GetWorkDetails"
	WorkAssignmentService_SubmitWork_FullMethodName           = "/workassignment.WorkAssignmentService/SubmitWork"
	WorkAssignmentService_GetTaskDetails_FullMethodName       = "/workassignment.WorkAssignmentService/GetTaskDetails"
	WorkAssignmentService_GenerateDownloadURL_FullMethodName  = "/workassignment.WorkAssignmentService/GenerateDownloadURL"
	WorkAssignmentService_CreateWork_FullMethodName           = "/workassignment.WorkAssignmentService/CreateWork"
	WorkAssignmentService_CheckExistingWork_FullMethodName    = "/workassignment.WorkAssignmentService/CheckExistingWork"
	WorkAssignmentService_GenerateUploadURL_FullMethodName    = "/workassignment.WorkAssignmentService/GenerateUploadURL"
)

// WorkAssignmentServiceClient is the client API for WorkAssignmentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkAssignmentServiceClient interface {
	GetWorksForAssistant(ctx context.Context, in *GetWorksForAssistantRequest, opts ...grpc.CallOption) (*GetWorksForAssistantResponse, error)
	GetWorkDetails(ctx context.Context, in *GetWorkDetailsRequest, opts ...grpc.CallOption) (*GetWorkDetailsResponse, error)
	SubmitWork(ctx context.Context, in *SubmitWorkRequest, opts ...grpc.CallOption) (*SubmitWorkResponse, error)
	GetTaskDetails(ctx context.Context, in *GetTaskDetailsRequest, opts ...grpc.CallOption) (*GetTaskDetailsResponse, error)
	GenerateDownloadURL(ctx context.Context, in *GenerateDownloadURLRequest, opts ...grpc.CallOption) (*GenerateDownloadURLResponse, error)
	CreateWork(ctx context.Context, in *CreateWorkRequest, opts ...grpc.CallOption) (*CreateWorkResponse, error)
	CheckExistingWork(ctx context.Context, in *CheckExistingWorkRequest, opts ...grpc.CallOption) (*CheckExistingWorkResponse, error)
	GenerateUploadURL(ctx context.Context, in *GenerateUploadURLRequest, opts ...grpc.CallOption) (*GenerateUploadURLResponse, error)
}

type workAssignmentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkAssignmentServiceClient(cc grpc.ClientConnInterface) WorkAssignmentServiceClient {
	return &workAssignmentServiceClient{cc}
}

func (c *workAssignmentServiceClient) GetWorksForAssistant(ctx context.Context, in *GetWorksForAssistantRequest, opts ...grpc.CallOption) (*GetWorksForAssistantResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetWorksForAssistantResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_GetWorksForAssistant_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) GetWorkDetails(ctx context.Context, in *GetWorkDetailsRequest, opts ...grpc.CallOption) (*GetWorkDetailsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetWorkDetailsResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_GetWorkDetails_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) SubmitWork(ctx context.Context, in *SubmitWorkRequest, opts ...grpc.CallOption) (*SubmitWorkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SubmitWorkResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_SubmitWork_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) GetTaskDetails(ctx context.Context, in *GetTaskDetailsRequest, opts ...grpc.CallOption) (*GetTaskDetailsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTaskDetailsResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_GetTaskDetails_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) GenerateDownloadURL(ctx context.Context, in *GenerateDownloadURLRequest, opts ...grpc.CallOption) (*GenerateDownloadURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateDownloadURLResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_GenerateDownloadURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) CreateWork(ctx context.Context, in *CreateWorkRequest, opts ...grpc.CallOption) (*CreateWorkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateWorkResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_CreateWork_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) CheckExistingWork(ctx context.Context, in *CheckExistingWorkRequest, opts ...grpc.CallOption) (*CheckExistingWorkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckExistingWorkResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_CheckExistingWork_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workAssignmentServiceClient) GenerateUploadURL(ctx context.Context, in *GenerateUploadURLRequest, opts ...grpc.CallOption) (*GenerateUploadURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenerateUploadURLResponse)
	err := c.cc.Invoke(ctx, WorkAssignmentService_GenerateUploadURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkAssignmentServiceServer is the server API for WorkAssignmentService service.
// All implementations must embed UnimplementedWorkAssignmentServiceServer
// for forward compatibility.
type WorkAssignmentServiceServer interface {
	GetWorksForAssistant(context.Context, *GetWorksForAssistantRequest) (*GetWorksForAssistantResponse, error)
	GetWorkDetails(context.Context, *GetWorkDetailsRequest) (*GetWorkDetailsResponse, error)
	SubmitWork(context.Context, *SubmitWorkRequest) (*SubmitWorkResponse, error)
	GetTaskDetails(context.Context, *GetTaskDetailsRequest) (*GetTaskDetailsResponse, error)
	GenerateDownloadURL(context.Context, *GenerateDownloadURLRequest) (*GenerateDownloadURLResponse, error)
	CreateWork(context.Context, *CreateWorkRequest) (*CreateWorkResponse, error)
	CheckExistingWork(context.Context, *CheckExistingWorkRequest) (*CheckExistingWorkResponse, error)
	GenerateUploadURL(context.Context, *GenerateUploadURLRequest) (*GenerateUploadURLResponse, error)
	mustEmbedUnimplementedWorkAssignmentServiceServer()
}

// UnimplementedWorkAssignmentServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWorkAssignmentServiceServer struct{}

func (UnimplementedWorkAssignmentServiceServer) GetWorksForAssistant(context.Context, *GetWorksForAssistantRequest) (*GetWorksForAssistantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorksForAssistant not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) GetWorkDetails(context.Context, *GetWorkDetailsRequest) (*GetWorkDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkDetails not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) SubmitWork(context.Context, *SubmitWorkRequest) (*SubmitWorkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitWork not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) GetTaskDetails(context.Context, *GetTaskDetailsRequest) (*GetTaskDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTaskDetails not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) GenerateDownloadURL(context.Context, *GenerateDownloadURLRequest) (*GenerateDownloadURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateDownloadURL not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) CreateWork(context.Context, *CreateWorkRequest) (*CreateWorkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWork not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) CheckExistingWork(context.Context, *CheckExistingWorkRequest) (*CheckExistingWorkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckExistingWork not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) GenerateUploadURL(context.Context, *GenerateUploadURLRequest) (*GenerateUploadURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateUploadURL not implemented")
}
func (UnimplementedWorkAssignmentServiceServer) mustEmbedUnimplementedWorkAssignmentServiceServer() {}
func (UnimplementedWorkAssignmentServiceServer) testEmbeddedByValue()                               {}

// UnsafeWorkAssignmentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkAssignmentServiceServer will
// result in compilation errors.
type UnsafeWorkAssignmentServiceServer interface {
	mustEmbedUnimplementedWorkAssignmentServiceServer()
}

func RegisterWorkAssignmentServiceServer(s grpc.ServiceRegistrar, srv WorkAssignmentServiceServer) {
	// If the following call pancis, it indicates UnimplementedWorkAssignmentServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&WorkAssignmentService_ServiceDesc, srv)
}

func _WorkAssignmentService_GetWorksForAssistant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorksForAssistantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).GetWorksForAssistant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_GetWorksForAssistant_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).GetWorksForAssistant(ctx, req.(*GetWorksForAssistantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_GetWorkDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkDetailsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).GetWorkDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_GetWorkDetails_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).GetWorkDetails(ctx, req.(*GetWorkDetailsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_SubmitWork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitWorkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).SubmitWork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_SubmitWork_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).SubmitWork(ctx, req.(*SubmitWorkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_GetTaskDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTaskDetailsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).GetTaskDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_GetTaskDetails_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).GetTaskDetails(ctx, req.(*GetTaskDetailsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_GenerateDownloadURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateDownloadURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).GenerateDownloadURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_GenerateDownloadURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).GenerateDownloadURL(ctx, req.(*GenerateDownloadURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_CreateWork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWorkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).CreateWork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_CreateWork_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).CreateWork(ctx, req.(*CreateWorkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_CheckExistingWork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckExistingWorkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).CheckExistingWork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_CheckExistingWork_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).CheckExistingWork(ctx, req.(*CheckExistingWorkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkAssignmentService_GenerateUploadURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateUploadURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkAssignmentServiceServer).GenerateUploadURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkAssignmentService_GenerateUploadURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkAssignmentServiceServer).GenerateUploadURL(ctx, req.(*GenerateUploadURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WorkAssignmentService_ServiceDesc is the grpc.ServiceDesc for WorkAssignmentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkAssignmentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "workassignment.WorkAssignmentService",
	HandlerType: (*WorkAssignmentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetWorksForAssistant",
			Handler:    _WorkAssignmentService_GetWorksForAssistant_Handler,
		},
		{
			MethodName: "GetWorkDetails",
			Handler:    _WorkAssignmentService_GetWorkDetails_Handler,
		},
		{
			MethodName: "SubmitWork",
			Handler:    _WorkAssignmentService_SubmitWork_Handler,
		},
		{
			MethodName: "GetTaskDetails",
			Handler:    _WorkAssignmentService_GetTaskDetails_Handler,
		},
		{
			MethodName: "GenerateDownloadURL",
			Handler:    _WorkAssignmentService_GenerateDownloadURL_Handler,
		},
		{
			MethodName: "CreateWork",
			Handler:    _WorkAssignmentService_CreateWork_Handler,
		},
		{
			MethodName: "CheckExistingWork",
			Handler:    _WorkAssignmentService_CheckExistingWork_Handler,
		},
		{
			MethodName: "GenerateUploadURL",
			Handler:    _WorkAssignmentService_GenerateUploadURL_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/workassignment/workassignment.proto",
}
