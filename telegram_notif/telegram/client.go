package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"telegram_notif/models"
)

type Client struct {
	token      string
	httpClient *http.Client
}

type keyboardButton struct {
	Text string `json:"text"`
}

type replyKeyboardMarkup struct {
	Keyboard              [][]keyboardButton `json:"keyboard"`
	ResizeKeyboard        bool               `json:"resize_keyboard"`
	OneTimeKeyboard       bool               `json:"one_time_keyboard"`
	InputFieldPlaceholder string             `json:"input_field_placeholder,omitempty"`
}

type replyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 70 * time.Second,
		},
	}
}

func (c *Client) GetUpdates(
	ctx context.Context,
	offset int64,
) ([]models.TelegramUpdate, error) {
	form := url.Values{}
	form.Set("offset", strconv.FormatInt(offset, 10))
	form.Set("timeout", "50")
	form.Set("allowed_updates", `["message"]`)

	var response models.TelegramUpdatesResponse

	if err := c.postForm(
		ctx,
		"getUpdates",
		form,
		&response,
	); err != nil {
		return nil, err
	}

	if !response.OK {
		return nil, fmt.Errorf(
			"Telegram API error: %s",
			response.Description,
		)
	}

	return response.Result, nil
}

func (c *Client) SendMessage(
	ctx context.Context,
	chatID int64,
	message string,
) error {
	return c.sendMessage(
		ctx,
		chatID,
		message,
		"",
	)
}

func (c *Client) SendMessageWithKeyboard(
	ctx context.Context,
	chatID int64,
	message string,
	buttonRows [][]string,
) error {
	keyboard := make(
		[][]keyboardButton,
		0,
		len(buttonRows),
	)

	for _, row := range buttonRows {
		keyboardRow := make(
			[]keyboardButton,
			0,
			len(row),
		)

		for _, buttonText := range row {
			keyboardRow = append(
				keyboardRow,
				keyboardButton{
					Text: buttonText,
				},
			)
		}

		keyboard = append(
			keyboard,
			keyboardRow,
		)
	}

	markup := replyKeyboardMarkup{
		Keyboard:              keyboard,
		ResizeKeyboard:        true,
		OneTimeKeyboard:       true,
		InputFieldPlaceholder: "Pilih bagian Anda",
	}

	markupJSON, err := json.Marshal(markup)
	if err != nil {
		return fmt.Errorf(
			"gagal membuat keyboard Telegram: %w",
			err,
		)
	}

	return c.sendMessage(
		ctx,
		chatID,
		message,
		string(markupJSON),
	)
}

func (c *Client) SendMessageRemoveKeyboard(
	ctx context.Context,
	chatID int64,
	message string,
) error {
	markup := replyKeyboardRemove{
		RemoveKeyboard: true,
	}

	markupJSON, err := json.Marshal(markup)
	if err != nil {
		return fmt.Errorf(
			"gagal menghapus keyboard Telegram: %w",
			err,
		)
	}

	return c.sendMessage(
		ctx,
		chatID,
		message,
		string(markupJSON),
	)
}

func (c *Client) sendMessage(
	ctx context.Context,
	chatID int64,
	message string,
	replyMarkup string,
) error {
	form := url.Values{}

	form.Set(
		"chat_id",
		strconv.FormatInt(chatID, 10),
	)

	form.Set(
		"text",
		message,
	)

	if strings.TrimSpace(replyMarkup) != "" {
		form.Set(
			"reply_markup",
			replyMarkup,
		)
	}

	var response models.TelegramBasicResponse

	if err := c.postForm(
		ctx,
		"sendMessage",
		form,
		&response,
	); err != nil {
		return err
	}

	if !response.OK {
		return fmt.Errorf(
			"Telegram API error: %s",
			response.Description,
		)
	}

	return nil
}

func (c *Client) postForm(
	ctx context.Context,
	method string,
	form url.Values,
	responseTarget any,
) error {
	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/%s",
		c.token,
		method,
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}

	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := json.NewDecoder(
		response.Body,
	).Decode(responseTarget); err != nil {
		return err
	}

	return nil
}
