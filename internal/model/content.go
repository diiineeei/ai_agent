package model

import "time"

type Part struct {
	Text         string `bson:"text,omitempty"          json:"text,omitempty"`
	FunctionCall *struct {
		Name string         `bson:"name"  json:"name"`
		Args map[string]any `bson:"args"  json:"args"`
	} `bson:"function_call,omitempty" json:"function_call,omitempty"`
	FunctionResponse *struct {
		Name     string `bson:"name"     json:"name"`
		Response any    `bson:"response" json:"response"`
	} `bson:"function_response,omitempty" json:"function_response,omitempty"`
}

type Content struct {
	Role      string    `bson:"role"       json:"role"`
	Parts     []Part    `bson:"parts"      json:"parts"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
