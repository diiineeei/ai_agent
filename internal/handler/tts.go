package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type TTSHandler struct {
	client *texttospeech.Client
}

func NewTTSHandler(client *texttospeech.Client) *TTSHandler { return &TTSHandler{client: client} }

var (
	reCodeBlock      = regexp.MustCompile("(?s)```.*?```")
	reInlineCode     = regexp.MustCompile("`([^`]+)`")
	reAsteriskBold   = regexp.MustCompile(`\*\*(.*?)\*\*`)
	reUnderBold      = regexp.MustCompile(`__(.*?)__`)
	reAsteriskItalic = regexp.MustCompile(`\*(.*?)\*`)
	reUnderItalic    = regexp.MustCompile(`_(.*?)_`)
	reHeading        = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)
	reURL            = regexp.MustCompile(`https?://\S+`)
	reListItem       = regexp.MustCompile(`(?m)^\s*[-*+]\s+`)
	reOrderedItem    = regexp.MustCompile(`(?m)^\s*\d+\.\s+`)
	reEmoji          = regexp.MustCompile(`[\x{1F000}-\x{1FFFF}]|[\x{2600}-\x{27BF}]`)
	reMultiSpace     = regexp.MustCompile(`\s{2,}`)
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

const defaultVoice = "pt-BR-Wavenet-B"

// Speak handles POST /tts
func (h *TTSHandler) Speak(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		jsonError(w, "TTS não disponível: credenciais não configuradas", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Text  string `json:"text"`
		Voice string `json:"voice"`
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

	if !strings.HasSuffix(clean, ".") && !strings.HasSuffix(clean, "!") && !strings.HasSuffix(clean, "?") {
		clean += "."
	}

	voice := body.Voice
	if voice == "" {
		voice = defaultVoice
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: clean},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "pt-BR",
			Name:         voice,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := h.client.SynthesizeSpeech(context.Background(), req)
	if err != nil {
		jsonError(w, "erro ao gerar áudio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "no-store")
	io.Copy(w, bytes.NewReader(resp.AudioContent)) //nolint:errcheck
}
