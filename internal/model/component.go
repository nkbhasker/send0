package model

type Component struct {
	Base
	Name    string `json:"name"`
	Content string `json:"content"`
}
