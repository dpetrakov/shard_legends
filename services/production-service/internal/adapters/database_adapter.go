package adapters

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/shard-legends/production-service/internal/database"
	"github.com/shard-legends/production-service/internal/storage"
)

// DatabaseAdapter адаптирует database.DB для storage.DatabaseInterface
type DatabaseAdapter struct {
	db *database.DB
}

// NewDatabaseAdapter создает новый адаптер для базы данных
func NewDatabaseAdapter(db *database.DB) storage.DatabaseInterface {
	return &DatabaseAdapter{db: db}
}

// QueryRow выполняет запрос, ожидающий одну строку результата
func (a *DatabaseAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) storage.Row {
	row := a.db.Pool().QueryRow(ctx, query, args...)
	return &RowAdapter{row: row}
}

// Query выполняет запрос, возвращающий множество строк
func (a *DatabaseAdapter) Query(ctx context.Context, query string, args ...interface{}) (storage.Rows, error) {
	rows, err := a.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &RowsAdapter{rows: rows}, nil
}

// Exec выполняет запрос без возврата строк
func (a *DatabaseAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := a.db.Pool().Exec(ctx, query, args...)
	return err
}

// BeginTx начинает транзакцию
func (a *DatabaseAdapter) BeginTx(ctx context.Context) (storage.Tx, error) {
	tx, err := a.db.Pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &TxAdapter{tx: tx}, nil
}

// Health проверяет состояние базы данных
func (a *DatabaseAdapter) Health(ctx context.Context) error {
	return a.db.Health(ctx)
}

// RowAdapter адаптирует pgx.Row для storage.Row
type RowAdapter struct {
	row pgx.Row
}

// Scan сканирует результат строки в переданные указатели
func (r *RowAdapter) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// RowsAdapter адаптирует pgx.Rows для storage.Rows
type RowsAdapter struct {
	rows pgx.Rows
}

// Next переходит к следующей строке
func (r *RowsAdapter) Next() bool {
	return r.rows.Next()
}

// Scan сканирует текущую строку в переданные указатели
func (r *RowsAdapter) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Err возвращает ошибку, возникшую во время итерации
func (r *RowsAdapter) Err() error {
	return r.rows.Err()
}

// Close закрывает rows
func (r *RowsAdapter) Close() {
	r.rows.Close()
}

// TxAdapter адаптирует database transaction для storage.Tx
type TxAdapter struct {
	tx pgx.Tx
}

// QueryRow выполняет запрос в контексте транзакции
func (t *TxAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) storage.Row {
	row := t.tx.QueryRow(ctx, query, args...)
	return &RowAdapter{row: row}
}

// Query выполняет запрос в контексте транзакции
func (t *TxAdapter) Query(ctx context.Context, query string, args ...interface{}) (storage.Rows, error) {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &RowsAdapter{rows: rows}, nil
}

// Exec выполняет запрос в контексте транзакции
func (t *TxAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.Exec(ctx, query, args...)
	return err
}

// Commit подтверждает транзакцию
func (t *TxAdapter) Commit() error {
	return t.tx.Commit(context.Background())
}

// Rollback отменяет транзакцию
func (t *TxAdapter) Rollback() error {
	return t.tx.Rollback(context.Background())
}