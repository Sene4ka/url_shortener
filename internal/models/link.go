package models

type Link struct {
	Id  string `db:"id" json:"id"`
	Url string `db:"url" json:"url"`
}
