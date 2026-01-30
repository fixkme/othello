package pb

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	marshaler   *protojson.MarshalOptions
	unmarshaler *protojson.UnmarshalOptions
)

func init() {
	marshaler = &protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "    ",
		AllowPartial:    true,
		UseProtoNames:   true,
		UseEnumNumbers:  true,
		EmitUnpopulated: true,
	}
	unmarshaler = &protojson.UnmarshalOptions{}
}

func PbToJsonText(message proto.Message) (string, error) {
	data, err := marshaler.Marshal(message)
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func JsonTextToPb(text string, message proto.Message) error {
	return unmarshaler.Unmarshal([]byte(text), message)
}
