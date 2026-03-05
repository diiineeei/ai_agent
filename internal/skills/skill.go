package skills

import "ai_agent/internal/agent"

// Skill is the contract that every agent tool must implement.
// Name() must exactly match the "name" field stored in the MongoDB skills collection.
type Skill interface {
	// Name returns the unique identifier used in MongoDB and in the FunctionDeclaration.
	Name() string

	// Declaration returns the full FunctionDeclaration to register with the agent.
	Declaration() *agent.FunctionDeclaration
}
