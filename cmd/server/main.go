package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"ai_agent/internal/agent"
	"ai_agent/internal/handler"
	"ai_agent/internal/model"
	"ai_agent/internal/repository"
	"ai_agent/internal/skills"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/genai"
)

func main() {
	apiKey := mustEnv("GEMINI_API_KEY")
	defaultModel := getEnv("MODEL", "gemini-2.5-flash")
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
	agentConfigRepo := repository.NewMongoAgentConfigRepository(db.Collection("agent_configs"))
	feedbackRepo := repository.NewMongoFeedbackRepository(db.Collection("feedback"))

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
	registry.RegisterSeedOnly("suggest_questions", "Sugere perguntas relevantes que o usuário pode fazer ao assistente com base no histórico da conversa e nas características do agente.")
	registry.RegisterSeedOnly("chess", "Permite jogar xadrez contra o agente de IA diretamente no browser.")

	// Seed default skills into MongoDB (idempotent)
	if err := registry.Seed(ctx); err != nil {
		log.Fatalf("seed de skills: %v", err)
	}

	// Seed default agent config if none exists
	count, err := agentConfigRepo.CountAll(ctx)
	if err != nil {
		log.Fatalf("contando agent configs: %v", err)
	}
	if count == 0 {
		_, err := agentConfigRepo.Create(ctx, model.AgentConfig{
			Name:              "Padrão",
			SystemInstruction: "Você é um assistente útil. Responda sempre em português.",
			Model:             defaultModel,
			EnabledSkills:     []string{"weather", "search_documents", "suggest_questions"},
		})
		if err != nil {
			log.Fatalf("seed de agent config: %v", err)
		}
	}

	// Handlers
	chatHandler := handler.NewChatHandler(geminiClient, sessionRepo, agentConfigRepo, registry)
	fileHandler := handler.NewFileHandler(fileRepo, embedder)
	skillHandler := handler.NewSkillHandler(skillRepo)
	agentConfigHandler := handler.NewAgentConfigHandler(agentConfigRepo, geminiClient)
	feedbackHandler := handler.NewFeedbackHandler(feedbackRepo)
	suggestHandler := handler.NewSuggestHandler(geminiClient, sessionRepo, agentConfigRepo)

	// Routes (Go 1.22+ ServeMux with method+pattern)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /prompt", chatHandler.SendPrompt)
	mux.HandleFunc("GET /sessions", chatHandler.ListSessions)
	mux.HandleFunc("PUT /sessions/{id}/name", chatHandler.RenameSession)
	mux.HandleFunc("GET /history", chatHandler.GetHistory)
	mux.HandleFunc("DELETE /history", chatHandler.DeleteHistory)
	mux.HandleFunc("POST /files", fileHandler.Upload)
	mux.HandleFunc("GET /files", fileHandler.List)
	mux.HandleFunc("DELETE /files/{id}", fileHandler.Delete)
	mux.HandleFunc("GET /skills", skillHandler.List)
	mux.HandleFunc("PUT /skills/{name}/toggle", skillHandler.Toggle)
	mux.HandleFunc("POST /feedback", feedbackHandler.Submit)
	mux.HandleFunc("GET /feedback", feedbackHandler.ForSession)
	mux.HandleFunc("GET /feedback/stats", feedbackHandler.Stats)
	mux.HandleFunc("GET /suggest-questions", suggestHandler.Suggest)
	mux.HandleFunc("GET /agent-configs", agentConfigHandler.List)
	mux.HandleFunc("POST /agent-configs", agentConfigHandler.Create)
	mux.HandleFunc("POST /agent-configs/improve-instruction", agentConfigHandler.ImproveInstruction)
	mux.HandleFunc("GET /agent-configs/{id}", agentConfigHandler.GetByID)
	mux.HandleFunc("PUT /agent-configs/{id}", agentConfigHandler.Update)
	mux.HandleFunc("DELETE /agent-configs/{id}", agentConfigHandler.Delete)

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
