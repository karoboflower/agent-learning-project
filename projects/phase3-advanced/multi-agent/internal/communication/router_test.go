package communication

import (
	"testing"
	"time"
)

func TestNewMessageRouter(t *testing.T) {
	router := NewMessageRouter()

	if router == nil {
		t.Fatal("NewMessageRouter returned nil")
	}

	if router.handlers == nil {
		t.Error("handlers map not initialized")
	}
}

func TestMessageRouter_RegisterHandler(t *testing.T) {
	router := NewMessageRouter()

	handler := func(msg *Message) error {
		return nil
	}

	router.RegisterHandler("TEST", handler)

	if !router.HasHandler("TEST") {
		t.Error("Handler not registered")
	}

	if router.GetHandlerCount() != 1 {
		t.Errorf("Expected 1 handler, got %d", router.GetHandlerCount())
	}
}

func TestMessageRouter_UnregisterHandler(t *testing.T) {
	router := NewMessageRouter()

	handler := func(msg *Message) error {
		return nil
	}

	router.RegisterHandler("TEST", handler)
	router.UnregisterHandler("TEST")

	if router.HasHandler("TEST") {
		t.Error("Handler should be unregistered")
	}

	if router.GetHandlerCount() != 0 {
		t.Errorf("Expected 0 handlers, got %d", router.GetHandlerCount())
	}
}

func TestMessageRouter_Route(t *testing.T) {
	router := NewMessageRouter()

	called := false
	handler := func(msg *Message) error {
		called = true
		return nil
	}

	router.RegisterHandler("TEST", handler)

	msg := &Message{
		Type: "TEST",
	}

	err := router.Route(msg)
	if err != nil {
		t.Fatalf("Route failed: %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestMessageRouter_Route_NoHandler(t *testing.T) {
	router := NewMessageRouter()

	msg := &Message{
		Type: "UNKNOWN",
	}

	err := router.Route(msg)
	if err == nil {
		t.Error("Expected error for unknown message type")
	}
}

func TestNewMessageQueue(t *testing.T) {
	queue := NewMessageQueue(10)

	if queue == nil {
		t.Fatal("NewMessageQueue returned nil")
	}

	if queue.size != 10 {
		t.Errorf("Expected size 10, got %d", queue.size)
	}

	if !queue.IsEmpty() {
		t.Error("New queue should be empty")
	}
}

func TestMessageQueue_Enqueue(t *testing.T) {
	queue := NewMessageQueue(10)

	msg := &Message{
		MessageID: "msg-001",
		Type:      "TEST",
	}

	err := queue.Enqueue(msg)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("Expected size 1, got %d", queue.Size())
	}

	if queue.IsEmpty() {
		t.Error("Queue should not be empty")
	}
}

func TestMessageQueue_Dequeue(t *testing.T) {
	queue := NewMessageQueue(10)

	msg := &Message{
		MessageID: "msg-001",
		Type:      "TEST",
	}

	queue.Enqueue(msg)

	dequeued, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}

	if dequeued.MessageID != msg.MessageID {
		t.Errorf("Expected message ID %s, got %s", msg.MessageID, dequeued.MessageID)
	}

	if queue.Size() != 0 {
		t.Errorf("Expected size 0, got %d", queue.Size())
	}
}

func TestMessageQueue_DequeueEmpty(t *testing.T) {
	queue := NewMessageQueue(10)

	_, err := queue.Dequeue()
	if err == nil {
		t.Error("Expected error when dequeuing from empty queue")
	}
}

func TestMessageQueue_FullQueue(t *testing.T) {
	queue := NewMessageQueue(2)

	queue.Enqueue(&Message{MessageID: "msg-001"})
	queue.Enqueue(&Message{MessageID: "msg-002"})

	if !queue.IsFull() {
		t.Error("Queue should be full")
	}

	err := queue.Enqueue(&Message{MessageID: "msg-003"})
	if err == nil {
		t.Error("Expected error when queue is full")
	}
}

func TestNewAckManager(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	if am == nil {
		t.Fatal("NewAckManager returned nil")
	}

	if am.timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", am.timeout)
	}
}

func TestAckManager_RegisterMessage(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	am.RegisterMessage("msg-001")

	ack, err := am.GetAck("msg-001")
	if err != nil {
		t.Fatalf("GetAck failed: %v", err)
	}

	if ack.Status != AckStatusPending {
		t.Errorf("Expected status PENDING, got %s", ack.Status)
	}
}

func TestAckManager_Confirm(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	am.RegisterMessage("msg-001")

	err := am.Confirm("msg-001", true, "")
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}

	ack, _ := am.GetAck("msg-001")
	if ack.Status != AckStatusConfirmed {
		t.Errorf("Expected status CONFIRMED, got %s", ack.Status)
	}
}

func TestAckManager_ConfirmFailed(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	am.RegisterMessage("msg-001")

	err := am.Confirm("msg-001", false, "test error")
	if err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}

	ack, _ := am.GetAck("msg-001")
	if ack.Status != AckStatusFailed {
		t.Errorf("Expected status FAILED, got %s", ack.Status)
	}

	if ack.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", ack.Error)
	}
}

func TestAckManager_WaitForAck(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	am.RegisterMessage("msg-001")

	// Confirm in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		am.Confirm("msg-001", true, "")
	}()

	// Wait for ack
	ack, err := am.WaitForAck("msg-001")
	if err != nil {
		t.Fatalf("WaitForAck failed: %v", err)
	}

	if ack.Status != AckStatusConfirmed {
		t.Errorf("Expected status CONFIRMED, got %s", ack.Status)
	}
}

func TestAckManager_WaitForAck_Timeout(t *testing.T) {
	am := NewAckManager(100 * time.Millisecond)

	am.RegisterMessage("msg-001")

	// Don't confirm - let it timeout
	_, err := am.WaitForAck("msg-001")
	if err == nil {
		t.Error("Expected timeout error")
	}

	ack, _ := am.GetAck("msg-001")
	if ack.Status != AckStatusTimeout {
		t.Errorf("Expected status TIMEOUT, got %s", ack.Status)
	}
}

func TestAckManager_CleanupExpired(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	// Register and confirm messages
	am.RegisterMessage("msg-001")
	am.Confirm("msg-001", true, "")

	am.RegisterMessage("msg-002")
	am.Confirm("msg-002", true, "")

	// Set old timestamp
	ack001, _ := am.GetAck("msg-001")
	ack001.Timestamp = time.Now().Add(-10 * time.Second)

	// Cleanup
	count := am.CleanupExpired(5 * time.Second)

	if count != 1 {
		t.Errorf("Expected 1 expired ack, got %d", count)
	}

	// msg-001 should be removed
	_, err := am.GetAck("msg-001")
	if err == nil {
		t.Error("Expected error for removed ack")
	}

	// msg-002 should still exist
	_, err = am.GetAck("msg-002")
	if err != nil {
		t.Error("msg-002 should still exist")
	}
}

func TestAckManager_GetPendingCount(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	am.RegisterMessage("msg-001")
	am.RegisterMessage("msg-002")
	am.RegisterMessage("msg-003")

	if am.GetPendingCount() != 3 {
		t.Errorf("Expected 3 pending, got %d", am.GetPendingCount())
	}

	am.Confirm("msg-001", true, "")

	if am.GetPendingCount() != 2 {
		t.Errorf("Expected 2 pending, got %d", am.GetPendingCount())
	}
}

func TestAckManager_GetAckStats(t *testing.T) {
	am := NewAckManager(5 * time.Second)

	am.RegisterMessage("msg-001")
	am.RegisterMessage("msg-002")
	am.RegisterMessage("msg-003")

	am.Confirm("msg-001", true, "")
	am.Confirm("msg-002", false, "error")

	stats := am.GetAckStats()

	if stats[AckStatusPending] != 1 {
		t.Errorf("Expected 1 pending, got %d", stats[AckStatusPending])
	}

	if stats[AckStatusConfirmed] != 1 {
		t.Errorf("Expected 1 confirmed, got %d", stats[AckStatusConfirmed])
	}

	if stats[AckStatusFailed] != 1 {
		t.Errorf("Expected 1 failed, got %d", stats[AckStatusFailed])
	}
}

func TestSerializeMessage(t *testing.T) {
	msg := &Message{
		MessageID: "msg-001",
		Type:      "TEST",
		From:      "agent-001",
		To:        "agent-002",
		Timestamp: "2026-01-29T10:00:00Z",
		Priority:  5,
		Payload: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}
}

func TestDeserializeMessage(t *testing.T) {
	jsonData := []byte(`{
		"message_id": "msg-001",
		"type": "TEST",
		"from": "agent-001",
		"to": "agent-002",
		"timestamp": "2026-01-29T10:00:00Z",
		"priority": 5,
		"payload": {"key": "value"}
	}`)

	msg, err := DeserializeMessage(jsonData)
	if err != nil {
		t.Fatalf("DeserializeMessage failed: %v", err)
	}

	if msg.MessageID != "msg-001" {
		t.Errorf("Expected message ID msg-001, got %s", msg.MessageID)
	}

	if msg.Type != "TEST" {
		t.Errorf("Expected type TEST, got %s", msg.Type)
	}

	if msg.Priority != 5 {
		t.Errorf("Expected priority 5, got %d", msg.Priority)
	}
}

func TestMessageBuilder(t *testing.T) {
	msg := NewMessageBuilder().
		SetMessageID("msg-001").
		SetType("TEST").
		SetFrom("agent-001").
		SetTo("agent-002").
		SetPriority(8).
		AddPayloadField("key", "value").
		AddMetadataField("trace", "123").
		Build()

	if msg.MessageID != "msg-001" {
		t.Errorf("Expected message ID msg-001, got %s", msg.MessageID)
	}

	if msg.Type != "TEST" {
		t.Errorf("Expected type TEST, got %s", msg.Type)
	}

	if msg.Priority != 8 {
		t.Errorf("Expected priority 8, got %d", msg.Priority)
	}

	if msg.Payload["key"] != "value" {
		t.Error("Payload field not set correctly")
	}

	if msg.Metadata["trace"] != "123" {
		t.Error("Metadata field not set correctly")
	}

	// Check timestamp was auto-set
	if msg.Timestamp == "" {
		t.Error("Timestamp should be auto-set")
	}
}

func TestMessageValidator_Validate(t *testing.T) {
	validator := NewMessageValidator()

	validMsg := &Message{
		MessageID: "msg-001",
		Type:      "TEST",
		From:      "agent-001",
		To:        "agent-002",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := validator.Validate(validMsg)
	if err != nil {
		t.Errorf("Valid message should pass validation: %v", err)
	}
}

func TestMessageValidator_Validate_Invalid(t *testing.T) {
	validator := NewMessageValidator()

	tests := []struct {
		name string
		msg  *Message
	}{
		{"missing message_id", &Message{Type: "TEST", From: "a", To: "b", Timestamp: time.Now().Format(time.RFC3339)}},
		{"missing type", &Message{MessageID: "msg-001", From: "a", To: "b", Timestamp: time.Now().Format(time.RFC3339)}},
		{"missing from", &Message{MessageID: "msg-001", Type: "TEST", To: "b", Timestamp: time.Now().Format(time.RFC3339)}},
		{"missing to", &Message{MessageID: "msg-001", Type: "TEST", From: "a", Timestamp: time.Now().Format(time.RFC3339)}},
		{"missing timestamp", &Message{MessageID: "msg-001", Type: "TEST", From: "a", To: "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.msg)
			if err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}

func TestMessageValidator_ValidatePayload(t *testing.T) {
	validator := NewMessageValidator()

	msg := &Message{
		Payload: map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		},
	}

	err := validator.ValidatePayload(msg, []string{"field1", "field2"})
	if err != nil {
		t.Errorf("Valid payload should pass validation: %v", err)
	}

	err = validator.ValidatePayload(msg, []string{"field1", "field3"})
	if err == nil {
		t.Error("Expected validation error for missing field")
	}
}

func BenchmarkSerializeMessage(b *testing.B) {
	msg := &Message{
		MessageID: "msg-001",
		Type:      "TEST",
		From:      "agent-001",
		To:        "agent-002",
		Timestamp: "2026-01-29T10:00:00Z",
		Priority:  5,
		Payload: map[string]interface{}{
			"key": "value",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SerializeMessage(msg)
	}
}

func BenchmarkDeserializeMessage(b *testing.B) {
	jsonData := []byte(`{"message_id":"msg-001","type":"TEST","from":"agent-001","to":"agent-002","timestamp":"2026-01-29T10:00:00Z","priority":5,"payload":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeserializeMessage(jsonData)
	}
}
