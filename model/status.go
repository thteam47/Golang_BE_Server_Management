package model

import "time"

type StatusDetail struct {
	Status string    `json:"status,omitempty" bson:"status,omitempty"`
	Time   time.Time `json:"time,omitempty" bson:"time,omitempty"`
}
type ListStatus struct {
	ChangeStatus []StatusDetail `json:"changeStatus,omitempty" bson:"changeStatus,omitempty"`
}