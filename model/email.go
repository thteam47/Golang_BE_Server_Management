package model

type Email struct {
	From string   `json:"from,omitempty" bson:"from,omitempty"`
	To   []string `json:"to,omitempty" bson:"to,omitempty"`
	Body string   `json:"body,omitempty" bson:"body,omitempty"`
}
