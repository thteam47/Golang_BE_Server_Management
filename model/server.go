package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var NilServer Server

type Server struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username    string             `json:"username,omitempty" bson:"username,omitempty"`
	Password    string             `json:"password,omitempty" bson:"password,omitempty"`
	ServerName  string             `json:"servername,omitempty" bson:"servername,omitempty"`
	Ip          string             `json:"ip,omitempty" bson:"ip,omitempty"`
	Port        int64              `json:"port,omitempty" bson:"port,omitempty"`
	Status      string             `json:"status,omitempty" bson:"status,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}
