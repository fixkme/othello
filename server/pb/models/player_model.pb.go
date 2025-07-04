// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.27.1
// source: models/player_model.proto

package models

import (
	datas "github.com/fixkme/othello/server/pb/datas"
	_ "github.com/fixkme/othello/server/pb/pbext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// @model
// 玩家数据
type PBPlayerModel struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	PlayerId int64                  `protobuf:"varint,1,opt,name=player_id,json=playerId,proto3" json:"player_id,omitempty"`
	// 玩家信息
	ModelPlayerInfo *datas.PBPlayerInfo `protobuf:"bytes,2,opt,name=model_player_info,json=modelPlayerInfo,proto3" json:"model_player_info,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *PBPlayerModel) Reset() {
	*x = PBPlayerModel{}
	mi := &file_models_player_model_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PBPlayerModel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PBPlayerModel) ProtoMessage() {}

func (x *PBPlayerModel) ProtoReflect() protoreflect.Message {
	mi := &file_models_player_model_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PBPlayerModel.ProtoReflect.Descriptor instead.
func (*PBPlayerModel) Descriptor() ([]byte, []int) {
	return file_models_player_model_proto_rawDescGZIP(), []int{0}
}

func (x *PBPlayerModel) GetPlayerId() int64 {
	if x != nil {
		return x.PlayerId
	}
	return 0
}

func (x *PBPlayerModel) GetModelPlayerInfo() *datas.PBPlayerInfo {
	if x != nil {
		return x.ModelPlayerInfo
	}
	return nil
}

var File_models_player_model_proto protoreflect.FileDescriptor

const file_models_player_model_proto_rawDesc = "" +
	"\n" +
	"\x19models/player_model.proto\x12\x06models\x1a\x17pbext/options_ext.proto\x1a\x17datas/player_data.proto\"o\n" +
	"\vPlayerModel\x12\x1b\n" +
	"\tplayer_id\x18\x01 \x01(\x03R\bplayerId\x12=\n" +
	"\x11model_player_info\x18\x02 \x01(\v2\x11.datas.PlayerInfoR\x0fmodelPlayerInfo:\x04\x80\xb5\x18\x01B,Z*github.com/fixkme/othello/server/pb/modelsb\x06proto3"

var (
	file_models_player_model_proto_rawDescOnce sync.Once
	file_models_player_model_proto_rawDescData []byte
)

func file_models_player_model_proto_rawDescGZIP() []byte {
	file_models_player_model_proto_rawDescOnce.Do(func() {
		file_models_player_model_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_models_player_model_proto_rawDesc), len(file_models_player_model_proto_rawDesc)))
	})
	return file_models_player_model_proto_rawDescData
}

var file_models_player_model_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_models_player_model_proto_goTypes = []any{
	(*PBPlayerModel)(nil),      // 0: models.PlayerModel
	(*datas.PBPlayerInfo)(nil), // 1: datas.PlayerInfo
}
var file_models_player_model_proto_depIdxs = []int32{
	1, // 0: models.PlayerModel.model_player_info:type_name -> datas.PlayerInfo
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_models_player_model_proto_init() }
func file_models_player_model_proto_init() {
	if File_models_player_model_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_models_player_model_proto_rawDesc), len(file_models_player_model_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_models_player_model_proto_goTypes,
		DependencyIndexes: file_models_player_model_proto_depIdxs,
		MessageInfos:      file_models_player_model_proto_msgTypes,
	}.Build()
	File_models_player_model_proto = out.File
	file_models_player_model_proto_goTypes = nil
	file_models_player_model_proto_depIdxs = nil
}
