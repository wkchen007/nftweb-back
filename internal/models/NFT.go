package models

type NFT struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Meta  string `json:"meta"`
	Image string `json:"image"`
}

type TokenItem struct {
	TokenID  string `json:"tokenId"`
	TokenURI string `json:"tokenURI,omitempty"`
	ImageURI string `json:"imageURI,omitempty"`
}
