package models

type NFT struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Meta  string `json:"meta"`
	Image string `json:"image"`
}
