package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"ai_agent/internal/agent"
	"ai_agent/internal/handler"
	"ai_agent/internal/repository"
	"ai_agent/internal/skills"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/genai"
)

func main() {
	apiKey := mustEnv("GEMINI_API_KEY")
	model := getEnv("MODEL", "gemini-2.5-flash")
	port := getEnv("HTTP_PORT", "8080")
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	mongoDBName := getEnv("MONGODB_DB", "ai_agent")

	ctx := context.Background()

	// Connect MongoDB
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("conectando ao MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			log.Printf("desconectando do MongoDB: %v", err)
		}
	}()

	db := mongoClient.Database(mongoDBName)

	// Repositories
	sessionRepo := repository.NewMongoSessionRepository(db.Collection("sessions"))
	fileRepo := repository.NewMongoFileRepository(db.Collection("files"))
	skillRepo := repository.NewMongoSkillRepository(db.Collection("skills"))
	settingsRepo := repository.NewMongoSettingsRepository(db.Collection("settings"))

	// Embedder
	embedder := agent.NewEmbedder()

	// Gemini client
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("criando cliente Gemini: %v", err)
	}

	// Skill registry
	registry := skills.NewSkillRegistry(skillRepo)
	registry.Register("weather", "Retorna as condições climáticas atuais para uma cidade.", func() skills.Skill {
		return &skills.WeatherSkill{}
	})
	registry.Register("search_documents", "Realiza busca semântica nos documentos enviados pelo usuário.", func() skills.Skill {
		return skills.NewSearchDocumentsSkill(fileRepo, embedder)
	})

	// Seed default skills into MongoDB (idempotent)
	if err := registry.Seed(ctx); err != nil {
		log.Fatalf("seed de skills: %v", err)
	}

	// Seed system instruction if not yet persisted
	existing, err := settingsRepo.GetSystemInstruction(ctx)
	if err != nil {
		log.Fatalf("lendo system instruction: %v", err)
	}
	if existing == "" {
		seed := "Você é um assistente útil. Responda sempre em português."
		if err := settingsRepo.SetSystemInstruction(ctx, seed); err != nil {
			log.Fatalf("seed de system instruction: %v", err)
		}
	}

	// Handlers
	chatHandler := handler.NewChatHandler(geminiClient, model, "", sessionRepo, registry, settingsRepo)
	fileHandler := handler.NewFileHandler(fileRepo, embedder)
	skillHandler := handler.NewSkillHandler(skillRepo)
	settingsHandler := handler.NewSettingsHandler(settingsRepo)

	// Routes (Go 1.22+ ServeMux with method+pattern)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /prompt", chatHandler.SendPrompt)
	mux.HandleFunc("GET /history", chatHandler.GetHistory)
	mux.HandleFunc("DELETE /history", chatHandler.DeleteHistory)
	mux.HandleFunc("POST /files", fileHandler.Upload)
	mux.HandleFunc("GET /files", fileHandler.List)
	mux.HandleFunc("DELETE /files/{id}", fileHandler.Delete)
	mux.HandleFunc("GET /skills", skillHandler.List)
	mux.HandleFunc("PUT /skills/{name}/toggle", skillHandler.Toggle)
	mux.HandleFunc("GET /settings/system-instruction", settingsHandler.GetSystemInstruction)
	mux.HandleFunc("PUT /settings/system-instruction", settingsHandler.SetSystemInstruction)

	log.Printf("servidor iniciado em :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("servidor: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("variável de ambiente %q é obrigatória", key)
	}
	return v
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
