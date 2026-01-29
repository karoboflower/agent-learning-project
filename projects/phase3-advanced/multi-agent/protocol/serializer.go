package protocol

import (
	"encoding/json"
	"fmt"
)

// Serializer 消息序列化器
type Serializer struct {
	prettyPrint bool
}

// NewSerializer 创建新的序列化器
func NewSerializer() *Serializer {
	return &Serializer{
		prettyPrint: false,
	}
}

// Serialize 序列化消息为JSON字节
func (s *Serializer) Serialize(msg *Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	var data []byte
	var err error

	if s.prettyPrint {
		data, err = json.MarshalIndent(msg, "", "  ")
	} else {
		data, err = json.Marshal(msg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	return data, nil
}

// Deserialize 反序列化JSON字节为消息
func (s *Serializer) Deserialize(data []byte) (*Message, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}

	return &msg, nil
}

// SerializeToString 序列化消息为JSON字符串
func (s *Serializer) SerializeToString(msg *Message) (string, error) {
	data, err := s.Serialize(msg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeFromString 从JSON字符串反序列化消息
func (s *Serializer) DeserializeFromString(str string) (*Message, error) {
	return s.Deserialize([]byte(str))
}

// SetPrettyPrint 设置是否格式化输出
func (s *Serializer) SetPrettyPrint(pretty bool) {
	s.prettyPrint = pretty
}

// SerializePayload 序列化负载
func SerializePayload(payload interface{}) (map[string]interface{}, error) {
	// 将结构体序列化为map
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}

// DeserializePayload 反序列化负载
func DeserializePayload(payload map[string]interface{}, target interface{}) error {
	// 将map反序列化为结构体
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %w", err)
	}

	return nil
}
