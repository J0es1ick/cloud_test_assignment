package ratelimit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/J0es1ick/test-assignment/internal/config"
	_ "github.com/lib/pq"
)

type Storage interface {
	Get(ctx context.Context, key string) (*TokenBucket, bool, error)
	Set(ctx context.Context, key string, bucket *TokenBucket) error
	Update(ctx context.Context, key string, updateFunc func(bucket *TokenBucket) (*TokenBucket, error)) error
}

type Database struct {
	db *sql.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=%d",
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Name,
        cfg.Database.SSLMode,
        int(cfg.Database.ConnectTimeout.Seconds()),
    )

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (s *Database) Init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS ratelimit (
			key TEXT PRIMARY KEY,
            capacity INTEGER NOT NULL,
            tokens INTEGER NOT NULL,
            rate TEXT NOT NULL,
            last_refill TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT NOW(),
            updated_at TIMESTAMP DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *Database) Get(ctx context.Context, key string) (*TokenBucket, bool, error) {
	row := s.db.QueryRowContext(ctx, "SELECT capacity, tokens, rate, last_refill FROM ratelimit WHERE key = $1", key)

	var tb TokenBucket
	err := row.Scan(&tb.capacity, &tb.tokens, &tb.rate, &tb.lastRefill)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &tb, true, nil
}

func (s *Database) Set(ctx context.Context, key string, bucket *TokenBucket) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ratelimit (key, capacity, tokens, rate, last_refill)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO UPDATE SET
			capacity = EXCLUDED.capacity,
			tokens = EXCLUDED.tokens,
			rate = EXCLUDED.rate,
			last_refill = EXCLUDED.last_refill,
			updated_at = NOW();
	`, key, bucket.capacity, bucket.tokens, bucket.rate.String(), bucket.lastRefill)

	return err
}

func (s *Database) Update(ctx context.Context, key string, updateFunc func(bucket *TokenBucket) (*TokenBucket, error)) error {
	tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    bucket, exists, err := s.Get(ctx, key)
    if err != nil {
        return err
    }
    
    if !exists {
        return sql.ErrNoRows
    }
    
    newBucket, err := updateFunc(bucket)
    if err != nil {
        return err
    }
	
    if err := s.Set(ctx, key, newBucket); err != nil {
        return err
    }
    
    return tx.Commit()
}