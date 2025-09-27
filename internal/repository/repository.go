package repository

import (
	"database/sql"

	"github.com/wkchen007/nftweb-back/internal/models"
)

type DatabaseRepo interface {
	Connection() *sql.DB
	AllNFTs() ([]*models.NFT, error)
	GetUserByEmail(email string) (*models.User, error)
}
