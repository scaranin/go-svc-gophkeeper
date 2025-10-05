package models

import (
	"encoding/json"
	"fmt"

	v1 "go-svc-gophkeeper/gen/go/v1"

	"google.golang.org/protobuf/encoding/protojson"
)

// MetadataWrapper обертка для работы с protobuf Metadata
type MetadataWrapper struct {
	*v1.Metadata
}

// NewMetadata создает новую обертку для Metadata
func NewMetadata() *MetadataWrapper {
	return &MetadataWrapper{
		Metadata: &v1.Metadata{},
	}
}

// FromProto создает обертку из protobuf Metadata
func FromProto(protoMetadata *v1.Metadata) *MetadataWrapper {
	if protoMetadata == nil {
		return NewMetadata()
	}
	return &MetadataWrapper{Metadata: protoMetadata}
}

// ToBytes сериализует Metadata в JSON для хранения в БД
func (mw *MetadataWrapper) ToBytes() ([]byte, error) {
	if mw.Metadata == nil {
		return json.Marshal(map[string]interface{}{})
	}

	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	data, err := marshaler.Marshal(mw.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return data, nil
}

// FromBytes десериализует Metadata из JSON bytes
func (mw *MetadataWrapper) FromBytes(data []byte) error {
	if len(data) == 0 {
		mw.Metadata = &v1.Metadata{}
		return nil
	}

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	metadata := &v1.Metadata{}
	if err := unmarshaler.Unmarshal(data, metadata); err != nil {
		var jsonData map[string]interface{}
		if jsonErr := json.Unmarshal(data, &jsonData); jsonErr != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		metadata = convertJSONToProtoMetadata(jsonData)
	}

	mw.Metadata = metadata
	return nil
}

// convertJSONToProtoMetadata конвертирует простой JSON map в protobuf Metadata
func convertJSONToProtoMetadata(jsonData map[string]interface{}) *v1.Metadata {
	metadata := &v1.Metadata{}

	if name, ok := jsonData["name"].(string); ok {
		metadata.Name = name
	}
	if description, ok := jsonData["description"].(string); ok {
		metadata.Description = description
	}
	if website, ok := jsonData["website"].(string); ok {
		metadata.Website = website
	}
	if tags, ok := jsonData["tags"].(string); ok {
		metadata.Tags = tags
	}
	if customFields, ok := jsonData["custom_fields"].(map[string]interface{}); ok {
		metadata.CustomFields = make(map[string]string)
		for k, v := range customFields {
			if strVal, ok := v.(string); ok {
				metadata.CustomFields[k] = strVal
			}
		}
	}

	return metadata
}

// ToProto возвращает protobuf представление
func (mw *MetadataWrapper) ToProto() *v1.Metadata {
	if mw.Metadata == nil {
		return &v1.Metadata{}
	}
	return mw.Metadata
}
