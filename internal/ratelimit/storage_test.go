package ratelimit_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/J0es1ick/test-assignment/internal/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStorage(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    storage := &ratelimit.Database{DB: db}

    t.Run("should create new bucket if not exists", func(t *testing.T) {
        mock.ExpectBegin()
        
        mock.ExpectQuery("SELECT capacity, tokens, rate, last_refill FROM ratelimit").
            WillReturnError(sql.ErrNoRows)
            
        mock.ExpectExec("INSERT INTO ratelimit").
            WithArgs("test", 10, 10, "1s", sqlmock.AnyArg()).
            WillReturnResult(sqlmock.NewResult(1, 1))
            
        mock.ExpectCommit()

        err := storage.Update(context.Background(), "test", func(b *ratelimit.TokenBucket) (*ratelimit.TokenBucket, error) {
            if b == nil {
                return ratelimit.NewTokenBucket(10, time.Second), nil
            }
            return b, nil
        })

        assert.NoError(t, err)
    })

    t.Run("should update existing bucket", func(t *testing.T) {
        now := time.Now()
        mock.ExpectBegin()
        
        rows := sqlmock.NewRows([]string{"capacity", "tokens", "rate", "last_refill"}).
            AddRow(10, 5, "1s", now)
        mock.ExpectQuery("SELECT capacity, tokens, rate, last_refill FROM ratelimit").
            WillReturnRows(rows)
            
        mock.ExpectExec("INSERT INTO ratelimit").
            WithArgs("test", 10, 4, "1s", sqlmock.AnyArg()).
            WillReturnResult(sqlmock.NewResult(1, 1))
            
        mock.ExpectCommit()

        err := storage.Update(context.Background(), "test", func(b *ratelimit.TokenBucket) (*ratelimit.TokenBucket, error) {
            b.Take()
            return b, nil
        })

        assert.NoError(t, err)
    })

    assert.NoError(t, mock.ExpectationsWereMet())
}