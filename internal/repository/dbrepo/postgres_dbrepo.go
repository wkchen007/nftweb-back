package dbrepo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/wkchen007/nftweb-back/internal/models"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) AllNFTs() ([]*models.NFT, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name, description, meta, image
		from
			nft
		where demo = '1'
		order by
			id
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nfts []*models.NFT

	for rows.Next() {
		var nft models.NFT
		err := rows.Scan(
			&nft.ID,
			&nft.Name,
			&nft.Desc,
			&nft.Meta,
			&nft.Image,
		)
		if err != nil {
			return nil, err
		}

		nfts = append(nfts, &nft)
	}

	return nfts, nil
}

func (m *PostgresDBRepo) GetTokenItem(ids []int) ([]models.TokenItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		select id, meta, image
		from nft
		where id in (%s)
	`, strings.Join(placeholders, ","))

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.TokenItem

	for rows.Next() {
		var token models.TokenItem
		err := rows.Scan(
			&token.TokenID,
			&token.TokenURI,
			&token.ImageURI,
		)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (m *PostgresDBRepo) GetBoxItem() (models.TokenItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select id, meta, image
		from nft
		where demo = '0'
		limit 1
	`

	var token models.TokenItem
	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&token.TokenID,
		&token.TokenURI,
		&token.ImageURI,
	)
	if err != nil {
		return models.TokenItem{}, err
	}

	return token, nil
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password,wallet_address,
			created_at, updated_at from users where email = $1`

	var user models.User
	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.WalletAddress,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
