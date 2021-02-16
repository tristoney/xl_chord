// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

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

// ChordClient is the client API for Chord service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChordClient interface {
	// GetPredecessor returns the node that is considered to be the current node's predecessor
	GetPredecessor(ctx context.Context, in *BaseReq, opts ...grpc.CallOption) (*NodeResp, error)
	// GetSuccessor returns the node that is considered to be the current node's successor
	GetSuccessor(ctx context.Context, in *BaseReq, opts ...grpc.CallOption) (*NodeResp, error)
	// Notify tell the Chord server that the Node think it is the predecessor, which
	// has the potential to initiate the transferring of keys
	Notify(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*BaseResp, error)
	// FindSuccessor finds the node's successor by ID
	FindSuccessor(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*NodeResp, error)
	// CheckPredecessor checks if the predecessor alive
	ChekPredecessor(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*BaseResp, error)
	// SetPredecessor sets the predecessor for a node
	SetPredecessor(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*BaseResp, error)
	// SetSuccessor sets the successor for a node
	SetSuccessor(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*BaseResp, error)
	// GetVal returns the value in the Chord ring for a given key
	GetVal(ctx context.Context, in *GetValReq, opts ...grpc.CallOption) (*GetValResp, error)
	// SetKey writes a k-v pair to the chord ring
	SetKey(ctx context.Context, in *SetKeyReq, opts ...grpc.CallOption) (*SetKeyResp, error)
	// DeleteKey delete the k-v pair of the chord ring
	DeleteKey(ctx context.Context, in *DeleteKeyReq, opts ...grpc.CallOption) (*DeleteKeyResp, error)
	// MultiDelete delete the k-v pairs of the chord ring;
	MultiDelete(ctx context.Context, in *MultiDeleteReq, opts ...grpc.CallOption) (*MultiDeleteResp, error)
	// GetKeys returns the k-v pairs between the given range of the Chord ring
	GetKeys(ctx context.Context, in *GetKeysReq, opts ...grpc.CallOption) (*GetKeysResp, error)
}

type chordClient struct {
	cc grpc.ClientConnInterface
}

func NewChordClient(cc grpc.ClientConnInterface) ChordClient {
	return &chordClient{cc}
}

func (c *chordClient) GetPredecessor(ctx context.Context, in *BaseReq, opts ...grpc.CallOption) (*NodeResp, error) {
	out := new(NodeResp)
	err := c.cc.Invoke(ctx, "/Chord/GetPredecessor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) GetSuccessor(ctx context.Context, in *BaseReq, opts ...grpc.CallOption) (*NodeResp, error) {
	out := new(NodeResp)
	err := c.cc.Invoke(ctx, "/Chord/GetSuccessor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) Notify(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*BaseResp, error) {
	out := new(BaseResp)
	err := c.cc.Invoke(ctx, "/Chord/Notify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) FindSuccessor(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*NodeResp, error) {
	out := new(NodeResp)
	err := c.cc.Invoke(ctx, "/Chord/FindSuccessor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) ChekPredecessor(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*BaseResp, error) {
	out := new(BaseResp)
	err := c.cc.Invoke(ctx, "/Chord/ChekPredecessor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) SetPredecessor(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*BaseResp, error) {
	out := new(BaseResp)
	err := c.cc.Invoke(ctx, "/Chord/SetPredecessor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) SetSuccessor(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*BaseResp, error) {
	out := new(BaseResp)
	err := c.cc.Invoke(ctx, "/Chord/SetSuccessor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) GetVal(ctx context.Context, in *GetValReq, opts ...grpc.CallOption) (*GetValResp, error) {
	out := new(GetValResp)
	err := c.cc.Invoke(ctx, "/Chord/GetVal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) SetKey(ctx context.Context, in *SetKeyReq, opts ...grpc.CallOption) (*SetKeyResp, error) {
	out := new(SetKeyResp)
	err := c.cc.Invoke(ctx, "/Chord/SetKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) DeleteKey(ctx context.Context, in *DeleteKeyReq, opts ...grpc.CallOption) (*DeleteKeyResp, error) {
	out := new(DeleteKeyResp)
	err := c.cc.Invoke(ctx, "/Chord/DeleteKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) MultiDelete(ctx context.Context, in *MultiDeleteReq, opts ...grpc.CallOption) (*MultiDeleteResp, error) {
	out := new(MultiDeleteResp)
	err := c.cc.Invoke(ctx, "/Chord/MultiDelete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chordClient) GetKeys(ctx context.Context, in *GetKeysReq, opts ...grpc.CallOption) (*GetKeysResp, error) {
	out := new(GetKeysResp)
	err := c.cc.Invoke(ctx, "/Chord/GetKeys", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChordServer is the server API for Chord service.
// All implementations should embed UnimplementedChordServer
// for forward compatibility
type ChordServer interface {
	// GetPredecessor returns the node that is considered to be the current node's predecessor
	GetPredecessor(context.Context, *BaseReq) (*NodeResp, error)
	// GetSuccessor returns the node that is considered to be the current node's successor
	GetSuccessor(context.Context, *BaseReq) (*NodeResp, error)
	// Notify tell the Chord server that the Node think it is the predecessor, which
	// has the potential to initiate the transferring of keys
	Notify(context.Context, *NodeReq) (*BaseResp, error)
	// FindSuccessor finds the node's successor by ID
	FindSuccessor(context.Context, *IDReq) (*NodeResp, error)
	// CheckPredecessor checks if the predecessor alive
	ChekPredecessor(context.Context, *IDReq) (*BaseResp, error)
	// SetPredecessor sets the predecessor for a node
	SetPredecessor(context.Context, *NodeReq) (*BaseResp, error)
	// SetSuccessor sets the successor for a node
	SetSuccessor(context.Context, *NodeReq) (*BaseResp, error)
	// GetVal returns the value in the Chord ring for a given key
	GetVal(context.Context, *GetValReq) (*GetValResp, error)
	// SetKey writes a k-v pair to the chord ring
	SetKey(context.Context, *SetKeyReq) (*SetKeyResp, error)
	// DeleteKey delete the k-v pair of the chord ring
	DeleteKey(context.Context, *DeleteKeyReq) (*DeleteKeyResp, error)
	// MultiDelete delete the k-v pairs of the chord ring;
	MultiDelete(context.Context, *MultiDeleteReq) (*MultiDeleteResp, error)
	// GetKeys returns the k-v pairs between the given range of the Chord ring
	GetKeys(context.Context, *GetKeysReq) (*GetKeysResp, error)
}

// UnimplementedChordServer should be embedded to have forward compatible implementations.
type UnimplementedChordServer struct {
}

func (UnimplementedChordServer) GetPredecessor(context.Context, *BaseReq) (*NodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPredecessor not implemented")
}
func (UnimplementedChordServer) GetSuccessor(context.Context, *BaseReq) (*NodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSuccessor not implemented")
}
func (UnimplementedChordServer) Notify(context.Context, *NodeReq) (*BaseResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Notify not implemented")
}
func (UnimplementedChordServer) FindSuccessor(context.Context, *IDReq) (*NodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindSuccessor not implemented")
}
func (UnimplementedChordServer) ChekPredecessor(context.Context, *IDReq) (*BaseResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChekPredecessor not implemented")
}
func (UnimplementedChordServer) SetPredecessor(context.Context, *NodeReq) (*BaseResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPredecessor not implemented")
}
func (UnimplementedChordServer) SetSuccessor(context.Context, *NodeReq) (*BaseResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetSuccessor not implemented")
}
func (UnimplementedChordServer) GetVal(context.Context, *GetValReq) (*GetValResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVal not implemented")
}
func (UnimplementedChordServer) SetKey(context.Context, *SetKeyReq) (*SetKeyResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetKey not implemented")
}
func (UnimplementedChordServer) DeleteKey(context.Context, *DeleteKeyReq) (*DeleteKeyResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteKey not implemented")
}
func (UnimplementedChordServer) MultiDelete(context.Context, *MultiDeleteReq) (*MultiDeleteResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MultiDelete not implemented")
}
func (UnimplementedChordServer) GetKeys(context.Context, *GetKeysReq) (*GetKeysResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKeys not implemented")
}

// UnsafeChordServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChordServer will
// result in compilation errors.
type UnsafeChordServer interface {
	mustEmbedUnimplementedChordServer()
}

func RegisterChordServer(s grpc.ServiceRegistrar, srv ChordServer) {
	s.RegisterService(&Chord_ServiceDesc, srv)
}

func _Chord_GetPredecessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BaseReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).GetPredecessor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/GetPredecessor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).GetPredecessor(ctx, req.(*BaseReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_GetSuccessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BaseReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).GetSuccessor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/GetSuccessor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).GetSuccessor(ctx, req.(*BaseReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).Notify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/Notify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).Notify(ctx, req.(*NodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_FindSuccessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IDReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).FindSuccessor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/FindSuccessor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).FindSuccessor(ctx, req.(*IDReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_ChekPredecessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IDReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).ChekPredecessor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/ChekPredecessor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).ChekPredecessor(ctx, req.(*IDReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_SetPredecessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).SetPredecessor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/SetPredecessor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).SetPredecessor(ctx, req.(*NodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_SetSuccessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).SetSuccessor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/SetSuccessor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).SetSuccessor(ctx, req.(*NodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_GetVal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetValReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).GetVal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/GetVal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).GetVal(ctx, req.(*GetValReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_SetKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetKeyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).SetKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/SetKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).SetKey(ctx, req.(*SetKeyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_DeleteKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteKeyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).DeleteKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/DeleteKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).DeleteKey(ctx, req.(*DeleteKeyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_MultiDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MultiDeleteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).MultiDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/MultiDelete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).MultiDelete(ctx, req.(*MultiDeleteReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chord_GetKeys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetKeysReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).GetKeys(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/GetKeys",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).GetKeys(ctx, req.(*GetKeysReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Chord_ServiceDesc is the grpc.ServiceDesc for Chord service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Chord_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Chord",
	HandlerType: (*ChordServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPredecessor",
			Handler:    _Chord_GetPredecessor_Handler,
		},
		{
			MethodName: "GetSuccessor",
			Handler:    _Chord_GetSuccessor_Handler,
		},
		{
			MethodName: "Notify",
			Handler:    _Chord_Notify_Handler,
		},
		{
			MethodName: "FindSuccessor",
			Handler:    _Chord_FindSuccessor_Handler,
		},
		{
			MethodName: "ChekPredecessor",
			Handler:    _Chord_ChekPredecessor_Handler,
		},
		{
			MethodName: "SetPredecessor",
			Handler:    _Chord_SetPredecessor_Handler,
		},
		{
			MethodName: "SetSuccessor",
			Handler:    _Chord_SetSuccessor_Handler,
		},
		{
			MethodName: "GetVal",
			Handler:    _Chord_GetVal_Handler,
		},
		{
			MethodName: "SetKey",
			Handler:    _Chord_SetKey_Handler,
		},
		{
			MethodName: "DeleteKey",
			Handler:    _Chord_DeleteKey_Handler,
		},
		{
			MethodName: "MultiDelete",
			Handler:    _Chord_MultiDelete_Handler,
		},
		{
			MethodName: "GetKeys",
			Handler:    _Chord_GetKeys_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/chord.proto",
}
