package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type TTSHandler struct {
	apiKey string
}

func NewTTSHandler(apiKey string) *TTSHandler { return &TTSHandler{apiKey: apiKey} }

var (
	reCodeBlock   = regexp.MustCompile("(?s)```.*?```")
	reInlineCode  = regexp.MustCompile("`([^`]+)`")
	reAsteriskBold = regexp.MustCompile(`\*\*(.*?)\*\*`)
	reUnderBold    = regexp.MustCompile(`__(.*?)__`)
	reAsteriskItalic = regexp.MustCompile(`\*(.*?)\*`)
	reUnderItalic    = regexp.MustCompile(`_(.*?)_`)
	reHeading     = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)
	reURL         = regexp.MustCompile(`https?://\S+`)
	reListItem    = regexp.MustCompile(`(?m)^\s*[-*+]\s+`)
	reOrderedItem = regexp.MustCompile(`(?m)^\s*\d+\.\s+`)
	reEmoji       = regexp.MustCompile(`[\x{1F000}-\x{1FFFF}]|[\x{2600}-\x{27BF}]`)
	reMultiSpace  = regexp.MustCompile(`\s{2,}`)
)

func cleanTextForTTS(text string) string {
	text = reCodeBlock.ReplaceAllString(text, ", código omitido, ")
	text = reInlineCode.ReplaceAllString(text, "$1")
	text = reAsteriskBold.ReplaceAllString(text, "$1")
	text = reUnderBold.ReplaceAllString(text, "$1")
	text = reAsteriskItalic.ReplaceAllString(text, "$1")
	text = reUnderItalic.ReplaceAllString(text, "$1")
	text = reHeading.ReplaceAllString(text, "$1. ")
	text = reURL.ReplaceAllString(text, "link")
	text = reListItem.ReplaceAllString(text, ", ")
	text = reOrderedItem.ReplaceAllString(text, ", ")
	text = reEmoji.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "\n\n", ". ")
	text = strings.ReplaceAll(text, "\n", ", ")
	text = reMultiSpace.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

type cloudTTSRequest struct {
	Input      ttsInput      `json:"input"`
	Voice      ttsVoice      `json:"voice"`
	AudioConfig ttsAudioConfig `json:"audioConfig"`
}

type ttsInput struct {
	Text string `json:"text"`
}

type ttsVoice struct {
	LanguageCode string `json:"languageCode"`
	Name         string `json:"name"`
}

type ttsAudioConfig struct {
	AudioEncoding string  `json:"audioEncoding"`
	SpeakingRate  float64 `json:"speakingRate"`
	Pitch         float64 `json:"pitch"`
}

type cloudTTSResponse struct {
	AudioContent string `json:"audioContent"` // base64 MP3
	Error        *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (h *TTSHandler) cloudTTSAudio(text string) ([]byte, error) {
	payload := cloudTTSRequest{
		Input: ttsInput{Text: text},
		Voice: ttsVoice{
			LanguageCode: "pt-BR",
			Name:         "pt-BR-Neural2-C",
		},
		AudioConfig: ttsAudioConfig{
			AudioEncoding: "MP3",
			SpeakingRate:  1.0,
			Pitch:         0.0,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := "https://texttospeech.googleapis.com/v1/text:synthesize?key=" + h.apiKey
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("cloud tts: %w", err)
	}
	defer resp.Body.Close()

	var result cloudTTSResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("cloud tts decode: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("cloud tts: %s", result.Error.Message)
	}

	audio, err := base64.StdEncoding.DecodeString(result.AudioContent)
	if err != nil {
		return nil, fmt.Errorf("cloud tts base64: %w", err)
	}
	return audio, nil
}

// Speak handles POST /tts
func (h *TTSHandler) Speak(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		jsonError(w, "campo 'text' é obrigatório", http.StatusBadRequest)
		return
	}

	clean := cleanTextForTTS(body.Text)
	if clean == "" {
		jsonError(w, "texto vazio após limpeza", http.StatusBadRequest)
		return
	}

	audio, err := h.cloudTTSAudio(clean)
	if err != nil {
		jsonError(w, "erro ao gerar áudio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "no-store")
	io.Copy(w, bytes.NewReader(audio)) //nolint:errcheck
}
