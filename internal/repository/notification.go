package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (
			id, batch_id, recipient, channel, content, priority, status,
			idempotency_key, retry_count, max_retries, scheduled_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		n.ID, n.BatchID, n.Recipient, string(n.Channel), n.Content,
		string(n.Priority), string(n.Status), n.IdempotencyKey,
		n.RetryCount, n.MaxRetries, n.ScheduledAt,
	)
	return err
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	var n domain.Notification
	query := `SELECT * FROM notifications WHERE id = $1`
	err := r.db.GetContext(ctx, &n, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepository) GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	var list []*domain.Notification
	query := `SELECT * FROM notifications WHERE batch_id = $1 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &list, query, batchID)
	return list, err
}

func (r *NotificationRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Notification, error) {
	var n domain.Notification
	query := `SELECT * FROM notifications WHERE idempotency_key = $1`
	err := r.db.GetContext(ctx, &n, query, key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepository) List(ctx context.Context, f domain.ListFilter) ([]*domain.Notification, error) {
	query := `SELECT * FROM notifications WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if f.Status != nil {
		query += fmt.Sprintf(` AND status = $%d`, argNum)
		args = append(args, string(*f.Status))
		argNum++
	}
	if f.Channel != nil {
		query += fmt.Sprintf(` AND channel = $%d`, argNum)
		args = append(args, string(*f.Channel))
		argNum++
	}
	if f.From != nil {
		query += fmt.Sprintf(` AND created_at >= $%d`, argNum)
		args = append(args, *f.From)
		argNum++
	}
	if f.To != nil {
		query += fmt.Sprintf(` AND created_at <= $%d`, argNum)
		args = append(args, *f.To)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argNum, argNum+1)
	if f.Limit <= 0 {
		f.Limit = 50
	}
	args = append(args, f.Limit, f.Offset)

	var list []*domain.Notification
	err := r.db.SelectContext(ctx, &list, query, args...)
	return list, err
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, providerMsgID, errMsg *string) error {
	query := `UPDATE notifications SET status = $1, provider_msg_id = $2, error_message = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, string(status), providerMsgID, errMsg, id)
	return err
}

func (r *NotificationRepository) IncrementRetry(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `UPDATE notifications SET retry_count = retry_count + 1, error_message = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, errMsg, id)
	return err
}

func (r *NotificationRepository) CancelPending(ctx context.Context, id uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET status = 'cancelled', updated_at = NOW() WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}
