package bot

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-bot/internal/xui"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBotSender is a mock implementation of the BotSender interface for testing.
// It captures messages sent via Send() or Request() for later inspection.
type MockBotSender struct {
	SentMessages []tgbotapi.Chattable
}

func (m *MockBotSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SentMessages = append(m.SentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *MockBotSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.SentMessages = append(m.SentMessages, c)
	return &tgbotapi.APIResponse{Ok: true}, nil
}

// MockXUIService is a mock implementation of the XUIService.
type MockXUIService struct {
	GetClientTrafficsFunc func(ctx context.Context, email string) ([]xui.ClientTraffic, error)
}

func (m *MockXUIService) GetClientTraffics(ctx context.Context, email string) ([]xui.ClientTraffic, error) {
	if m.GetClientTrafficsFunc != nil {
		return m.GetClientTrafficsFunc(ctx, email)
	}
	return nil, errors.New("GetClientTrafficsFunc not implemented")
}

func TestHandleGetClientCommand_TableFormat(t *testing.T) {
	// --- Arrange ---

	// 1. Создаем мок для XUI сервиса, который вернет тестовые данные
	mockXUIService := &MockXUIService{
		GetClientTrafficsFunc: func(ctx context.Context, email string) ([]xui.ClientTraffic, error) {
			return []xui.ClientTraffic{
				{
					ID:         1,
					Email:      "test@example.com",
					Enable:     true,
					Up:         1024 * 1024 * 1024,      // 1 GB
					Down:       2 * 1024 * 1024 * 1024,  // 2 GB
					Total:      10 * 1024 * 1024 * 1024, // 10 GB
					ExpiryTime: 1735689600000,           // 2025-01-01 00:00:00 UTC
				},
			}, nil
		},
	}

	// 2. Создаем мок для нашего BotSender
	mockBot := &MockBotSender{}

	// 3. Создаем фейковое сообщение от пользователя.
	// Важно сымитировать не только текст, но и "сущность" (Entity) команды,
	// так как метод CommandArguments() работает на основе этих данных.
	message := &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 12345},
		Text: "/getclient test@example.com",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 10}, // Длина команды "/getclient"
		},
	}

	// --- Act ---
	handleGetClientCommand(context.Background(), message, mockBot, mockXUIService)

	// --- Assert ---
	// Проверяем, что бот попытался отправить ровно одно сообщение
	require.Len(t, mockBot.SentMessages, 1, "Expected one message to be sent")

	// Проверяем, что это сообщение - tgbotapi.MessageConfig
	msg, ok := mockBot.SentMessages[0].(tgbotapi.MessageConfig)
	require.True(t, ok, "Sent chattable should be of type MessageConfig")

	// Проверяем содержимое сообщения
	assert.Equal(t, tgbotapi.ModeHTML, msg.ParseMode, "ParseMode should be HTML")
	assert.Contains(t, msg.Text, "<pre>", "Message should contain a <pre> tag")
	assert.Contains(t, msg.Text, "</pre>", "Message should contain a </pre> tag")
	assert.Contains(t, msg.Text, "Email", "Message should contain 'Email' header")
	assert.Contains(t, msg.Text, "Usage (GB)", "Message should contain 'Usage (GB)' header")
	assert.Contains(t, msg.Text, "Expiry", "Message should contain 'Expiry' header")
	assert.Contains(t, msg.Text, "Status", "Message should contain 'Status' header")
	assert.Contains(t, msg.Text, "test@example.com", "Message should contain the client's email")
	assert.Contains(t, msg.Text, "3.00/10.00", "Message should contain the correct usage")
	assert.Contains(t, msg.Text, "01.01.2025", "Message should contain the correct expiry date")
	assert.True(t, strings.Contains(msg.Text, "✅"), "Message should contain the correct status icon")
}
