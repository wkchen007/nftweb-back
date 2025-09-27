package repository

import (
	"database/sql"

	"github.com/wkchen007/nftweb-back/internal/models"
)

type DatabaseRepo interface {
	Connection() *sql.DB
	AllNFTs() ([]*models.NFT, error)
	GetTokenItem(id []int) ([]models.TokenItem, error)
	GetBoxItem() (models.TokenItem, error)
	GetUserByEmail(email string) (*models.User, error)
}
