package skills

import (
	"context"
	"fmt"
	"log"

	"ai_agent/internal/agent"
	"ai_agent/internal/model"
	"ai_agent/internal/repository"
)

// factory is a constructor that creates a Skill instance with injected dependencies.
// nil indicates a "seed-only" skill managed externally (e.g. contextual skills).
type factory func() Skill

type skillMeta struct {
	description string
	fn          factory // nil = seed-only, not loaded via registry
}

// SkillRegistry manages the lifecycle of skills: seeding defaults into MongoDB,
// loading enabled skills at request time, and wiring them into the agent.
type SkillRegistry struct {
	skillRepo repository.SkillRepository
	entries   map[string]skillMeta
}

func NewSkillRegistry(skillRepo repository.SkillRepository) *SkillRegistry {
	return &SkillRegistry{
		skillRepo: skillRepo,
		entries:   make(map[string]skillMeta),
	}
}

// Register adds a named factory function. Call before Seed.
func (r *SkillRegistry) Register(name, description string, f factory) {
	r.entries[name] = skillMeta{description: description, fn: f}
}

// RegisterSeedOnly registers a skill for MongoDB seeding and management UI only.
// The skill is NOT loaded via the registry; the caller manages instantiation.
func (r *SkillRegistry) RegisterSeedOnly(name, description string) {
	r.entries[name] = skillMeta{description: description, fn: nil}
}

// IsEnabled reports whether the named skill is globally enabled in MongoDB.
func (r *SkillRegistry) IsEnabled(ctx context.Context, name string) bool {
	enabled, err := r.skillRepo.ListEnabled(ctx)
	if err != nil {
		return false
	}
	for _, doc := range enabled {
		if doc.Name == name {
			return true
		}
	}
	return false
}

// Seed upserts all registered skills into MongoDB using $setOnInsert,
// so the user's enabled/disabled state is preserved across restarts.
func (r *SkillRegistry) Seed(ctx context.Context) error {
	for name, meta := range r.entries {
		doc := model.SkillDocument{
			Name:        name,
			Description: meta.description,
			Enabled:     true,
		}
		if err := r.skillRepo.Seed(ctx, doc); err != nil {
			return fmt.Errorf("seeding skill %q: %w", name, err)
		}
	}
	return nil
}

// LoadEnabled queries MongoDB for enabled skills and registers their
// FunctionDeclarations with the given agent. Call this once per request.
func (r *SkillRegistry) LoadEnabled(ctx context.Context, a agent.Agent) error {
	enabled, err := r.skillRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("loading enabled skills: %w", err)
	}
	for _, doc := range enabled {
		meta, ok := r.entries[doc.Name]
		if !ok {
			log.Printf("WARNING: skill %q is enabled in DB but has no registered factory", doc.Name)
			continue
		}
		if meta.fn == nil {
			continue // seed-only / contextual skill, managed externally
		}
		skill := meta.fn()
		if err := a.AddFunctionCall(skill.Declaration()); err != nil {
			return fmt.Errorf("registering skill %q with agent: %w", doc.Name, err)
		}
	}
	return nil
}

// LoadByNames loads only globally-enabled skills whose names are in the provided list.
// It acts as the intersection of globally enabled skills and the agent's configured skills.
func (r *SkillRegistry) LoadByNames(ctx context.Context, a agent.Agent, names []string) error {
	if len(names) == 0 {
		return nil
	}
	allowed := make(map[string]bool, len(names))
	for _, n := range names {
		allowed[n] = true
	}
	enabled, err := r.skillRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("loading enabled skills: %w", err)
	}
	for _, doc := range enabled {
		if !allowed[doc.Name] {
			continue
		}
		meta, ok := r.entries[doc.Name]
		if !ok {
			log.Printf("WARNING: skill %q is enabled in DB but has no registered factory", doc.Name)
			continue
		}
		if meta.fn == nil {
			continue // seed-only / contextual skill, managed externally
		}
		skill := meta.fn()
		if err := a.AddFunctionCall(skill.Declaration()); err != nil {
			return fmt.Errorf("registering skill %q with agent: %w", doc.Name, err)
		}
	}
	return nil
}
