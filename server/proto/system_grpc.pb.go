// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: server/proto/system.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SystemClient is the client API for System service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SystemClient interface {
	// GetInfo returns info about the client
	GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ServerStatus, error)
	// PeersAdd adds a new peer
	PeersAdd(ctx context.Context, in *PeersAddRequest, opts ...grpc.CallOption) (*PeersAddResponse, error)
	// PeersList returns the list of peers
	PeersList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*PeersListResponse, error)
	// PeersInfo returns the info of a peer
	PeersStatus(ctx context.Context, in *PeersStatusRequest, opts ...grpc.CallOption) (*Peer, error)
	// Subscribe subscribes to blockchain events
	Subscribe(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (System_SubscribeClient, error)
	// Export returns blockchain data
	BlockByNumber(ctx context.Context, in *BlockByNumberRequest, opts ...grpc.CallOption) (*BlockResponse, error)
	// Export returns blockchain data
	Export(ctx context.Context, in *ExportRequest, opts ...grpc.CallOption) (System_ExportClient, error)
	// WhitelistAdd adds some contracts to ddos white list
	WhitelistAddList(ctx context.Context, in *WhitelistAddListRequest, opts ...grpc.CallOption) (*WhitelistAddListResponse, error)
	// whitelistDelete deletes some contracts from ddos white list
	WhitelistDeleteList(ctx context.Context, in *WhitelistDeleteListRequest, opts ...grpc.CallOption) (*WhitelistDeleteListResponse, error)
	// query ddos contract list
	DDOSContractList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*DDOSContractListResponse, error)
}

type systemClient struct {
	cc grpc.ClientConnInterface
}

func NewSystemClient(cc grpc.ClientConnInterface) SystemClient {
	return &systemClient{cc}
}

func (c *systemClient) GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ServerStatus, error) {
	out := new(ServerStatus)
	err := c.cc.Invoke(ctx, "/v1.System/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) PeersAdd(ctx context.Context, in *PeersAddRequest, opts ...grpc.CallOption) (*PeersAddResponse, error) {
	out := new(PeersAddResponse)
	err := c.cc.Invoke(ctx, "/v1.System/PeersAdd", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) PeersList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*PeersListResponse, error) {
	out := new(PeersListResponse)
	err := c.cc.Invoke(ctx, "/v1.System/PeersList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) PeersStatus(ctx context.Context, in *PeersStatusRequest, opts ...grpc.CallOption) (*Peer, error) {
	out := new(Peer)
	err := c.cc.Invoke(ctx, "/v1.System/PeersStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) Subscribe(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (System_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &System_ServiceDesc.Streams[0], "/v1.System/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &systemSubscribeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type System_SubscribeClient interface {
	Recv() (*BlockchainEvent, error)
	grpc.ClientStream
}

type systemSubscribeClient struct {
	grpc.ClientStream
}

func (x *systemSubscribeClient) Recv() (*BlockchainEvent, error) {
	m := new(BlockchainEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *systemClient) BlockByNumber(ctx context.Context, in *BlockByNumberRequest, opts ...grpc.CallOption) (*BlockResponse, error) {
	out := new(BlockResponse)
	err := c.cc.Invoke(ctx, "/v1.System/BlockByNumber", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) Export(ctx context.Context, in *ExportRequest, opts ...grpc.CallOption) (System_ExportClient, error) {
	stream, err := c.cc.NewStream(ctx, &System_ServiceDesc.Streams[1], "/v1.System/Export", opts...)
	if err != nil {
		return nil, err
	}
	x := &systemExportClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type System_ExportClient interface {
	Recv() (*ExportEvent, error)
	grpc.ClientStream
}

type systemExportClient struct {
	grpc.ClientStream
}

func (x *systemExportClient) Recv() (*ExportEvent, error) {
	m := new(ExportEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *systemClient) WhitelistAddList(ctx context.Context, in *WhitelistAddListRequest, opts ...grpc.CallOption) (*WhitelistAddListResponse, error) {
	out := new(WhitelistAddListResponse)
	err := c.cc.Invoke(ctx, "/v1.System/WhitelistAddList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) WhitelistDeleteList(ctx context.Context, in *WhitelistDeleteListRequest, opts ...grpc.CallOption) (*WhitelistDeleteListResponse, error) {
	out := new(WhitelistDeleteListResponse)
	err := c.cc.Invoke(ctx, "/v1.System/WhitelistDeleteList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemClient) DDOSContractList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*DDOSContractListResponse, error) {
	out := new(DDOSContractListResponse)
	err := c.cc.Invoke(ctx, "/v1.System/DDOSContractList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SystemServer is the server API for System service.
// All implementations must embed UnimplementedSystemServer
// for forward compatibility
type SystemServer interface {
	// GetInfo returns info about the client
	GetStatus(context.Context, *emptypb.Empty) (*ServerStatus, error)
	// PeersAdd adds a new peer
	PeersAdd(context.Context, *PeersAddRequest) (*PeersAddResponse, error)
	// PeersList returns the list of peers
	PeersList(context.Context, *emptypb.Empty) (*PeersListResponse, error)
	// PeersInfo returns the info of a peer
	PeersStatus(context.Context, *PeersStatusRequest) (*Peer, error)
	// Subscribe subscribes to blockchain events
	Subscribe(*emptypb.Empty, System_SubscribeServer) error
	// Export returns blockchain data
	BlockByNumber(context.Context, *BlockByNumberRequest) (*BlockResponse, error)
	// Export returns blockchain data
	Export(*ExportRequest, System_ExportServer) error
	// WhitelistAdd adds some contracts to ddos white list
	WhitelistAddList(context.Context, *WhitelistAddListRequest) (*WhitelistAddListResponse, error)
	// whitelistDelete deletes some contracts from ddos white list
	WhitelistDeleteList(context.Context, *WhitelistDeleteListRequest) (*WhitelistDeleteListResponse, error)
	// query ddos contract list
	DDOSContractList(context.Context, *emptypb.Empty) (*DDOSContractListResponse, error)
	mustEmbedUnimplementedSystemServer()
}

// UnimplementedSystemServer must be embedded to have forward compatible implementations.
type UnimplementedSystemServer struct {
}

func (UnimplementedSystemServer) GetStatus(context.Context, *emptypb.Empty) (*ServerStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedSystemServer) PeersAdd(context.Context, *PeersAddRequest) (*PeersAddResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeersAdd not implemented")
}
func (UnimplementedSystemServer) PeersList(context.Context, *emptypb.Empty) (*PeersListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeersList not implemented")
}
func (UnimplementedSystemServer) PeersStatus(context.Context, *PeersStatusRequest) (*Peer, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeersStatus not implemented")
}
func (UnimplementedSystemServer) Subscribe(*emptypb.Empty, System_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedSystemServer) BlockByNumber(context.Context, *BlockByNumberRequest) (*BlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BlockByNumber not implemented")
}
func (UnimplementedSystemServer) Export(*ExportRequest, System_ExportServer) error {
	return status.Errorf(codes.Unimplemented, "method Export not implemented")
}
func (UnimplementedSystemServer) WhitelistAddList(context.Context, *WhitelistAddListRequest) (*WhitelistAddListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WhitelistAddList not implemented")
}
func (UnimplementedSystemServer) WhitelistDeleteList(context.Context, *WhitelistDeleteListRequest) (*WhitelistDeleteListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WhitelistDeleteList not implemented")
}
func (UnimplementedSystemServer) DDOSContractList(context.Context, *emptypb.Empty) (*DDOSContractListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DDOSContractList not implemented")
}
func (UnimplementedSystemServer) mustEmbedUnimplementedSystemServer() {}

// UnsafeSystemServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SystemServer will
// result in compilation errors.
type UnsafeSystemServer interface {
	mustEmbedUnimplementedSystemServer()
}

func RegisterSystemServer(s grpc.ServiceRegistrar, srv SystemServer) {
	s.RegisterService(&System_ServiceDesc, srv)
}

func _System_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/GetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).GetStatus(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_PeersAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeersAddRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).PeersAdd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/PeersAdd",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).PeersAdd(ctx, req.(*PeersAddRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_PeersList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).PeersList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/PeersList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).PeersList(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_PeersStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeersStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).PeersStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/PeersStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).PeersStatus(ctx, req.(*PeersStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SystemServer).Subscribe(m, &systemSubscribeServer{stream})
}

type System_SubscribeServer interface {
	Send(*BlockchainEvent) error
	grpc.ServerStream
}

type systemSubscribeServer struct {
	grpc.ServerStream
}

func (x *systemSubscribeServer) Send(m *BlockchainEvent) error {
	return x.ServerStream.SendMsg(m)
}

func _System_BlockByNumber_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockByNumberRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).BlockByNumber(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/BlockByNumber",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).BlockByNumber(ctx, req.(*BlockByNumberRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_Export_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ExportRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SystemServer).Export(m, &systemExportServer{stream})
}

type System_ExportServer interface {
	Send(*ExportEvent) error
	grpc.ServerStream
}

type systemExportServer struct {
	grpc.ServerStream
}

func (x *systemExportServer) Send(m *ExportEvent) error {
	return x.ServerStream.SendMsg(m)
}

func _System_WhitelistAddList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WhitelistAddListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).WhitelistAddList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/WhitelistAddList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).WhitelistAddList(ctx, req.(*WhitelistAddListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_WhitelistDeleteList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WhitelistDeleteListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).WhitelistDeleteList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/WhitelistDeleteList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).WhitelistDeleteList(ctx, req.(*WhitelistDeleteListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _System_DDOSContractList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemServer).DDOSContractList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.System/DDOSContractList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemServer).DDOSContractList(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// System_ServiceDesc is the grpc.ServiceDesc for System service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var System_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v1.System",
	HandlerType: (*SystemServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStatus",
			Handler:    _System_GetStatus_Handler,
		},
		{
			MethodName: "PeersAdd",
			Handler:    _System_PeersAdd_Handler,
		},
		{
			MethodName: "PeersList",
			Handler:    _System_PeersList_Handler,
		},
		{
			MethodName: "PeersStatus",
			Handler:    _System_PeersStatus_Handler,
		},
		{
			MethodName: "BlockByNumber",
			Handler:    _System_BlockByNumber_Handler,
		},
		{
			MethodName: "WhitelistAddList",
			Handler:    _System_WhitelistAddList_Handler,
		},
		{
			MethodName: "WhitelistDeleteList",
			Handler:    _System_WhitelistDeleteList_Handler,
		},
		{
			MethodName: "DDOSContractList",
			Handler:    _System_DDOSContractList_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _System_Subscribe_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Export",
			Handler:       _System_Export_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "server/proto/system.proto",
}
