package mysql

import (
	"context"
	"database/sql"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	getQuery = `
SELECT umr.message_id, mj.content, mj.username, mj.sender_id, mj.created_at, mj.updated_at
FROM user_message_recipients umr
    INNER JOIN (
        SELECT m.id, m.content, m.sender_id, m.created_at, m.updated_at, u.username
        FROM messages m
            INNER JOIN (
                SELECT id, username
                FROM users
            ) u ON m.sender_id = u.id
    ) mj ON umr.message_id = mj.id
`
	getQueryByUserID = getQuery + " WHERE umr.recipient_id = ?;"

	getQueryByMessageID = getQuery + " WHERE umr.message_id = ?;"
)

var _ store.MessageStore = (*messageStore)(nil)

type messageStore struct {
	db *sql.DB
}

func (s *messageStore) Create(ctx context.Context, msg store.Message, recipientUserIDs []int64) error {
	err := s.create(ctx, msg, recipientUserIDs)
	if err == nil {
		return nil
	}

	if sqlErr, ok := err.(*mysql.MySQLError); ok {
		if sqlErr.Number == 1062 {
			return store.ErrDuplicate
		}
	}
	return err
}

func (s *messageStore) GetByID(ctx context.Context, msgID int64) (*store.Message, error) {
	row := s.db.QueryRowContext(ctx, getQueryByMessageID, msgID)

	var msg store.Message

	err := row.Scan(&msg.ID, &msg.Content, &msg.Sender, &msg.SenderID, &msg.SentDateTime, &msg.UpdatedDateTime)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (s *messageStore) Get(ctx context.Context, userID int64) ([]*store.Message, error) {
	rows, err := s.db.QueryContext(ctx, getQueryByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*store.Message
	for rows.Next() {
		var msg store.Message

		if err := rows.Scan(&msg.ID, &msg.Content, &msg.Sender, &msg.SenderID, &msg.SentDateTime, &msg.UpdatedDateTime); err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

// Update returns ErrNotFound if the message does not exist.
func (s *messageStore) Update(ctx context.Context, msg store.Message, recipientUserIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.Exec("UPDATE messages SET content=?, updated_at=? WHERE id=?", msg.Content, msg.UpdatedDateTime, msg.ID)
	if err != nil {
		return err
	}

	// Check if food does not exist
	if affected, err := res.RowsAffected(); err != nil {
		_ = tx.Rollback()
		return err
	} else if affected < 1 {
		_ = tx.Rollback()
		return store.ErrNotFound
	}

	// Delete the existing recipients.
	_, err = tx.Exec("DELETE FROM user_message_recipients WHERE message_id=?", msg.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Add the new recipients.
	err = s.createRecipients(ctx, tx, msg.ID, recipientUserIDs)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Delete returns ErrNotFound if the food does not exist.
func (s *messageStore) Delete(ctx context.Context, messageID int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM messages WHERE id=?", messageID)
	if err != nil {
		return err
	}

	// Check if food does not exist
	if affected, err := res.RowsAffected(); err != nil {
		return err
	} else if affected < 1 {
		return store.ErrNotFound
	}

	return nil
}

func (s *messageStore) create(ctx context.Context, msg store.Message, recipientUserIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, "INSERT INTO messages(content, sender_id, created_at, updated_at) VALUES (?, ?, ?, ?)",
		msg.Content, msg.SenderID, msg.SentDateTime, msg.SentDateTime)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	messageID, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = s.createRecipients(ctx, tx, messageID, recipientUserIDs)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *messageStore) createRecipients(ctx context.Context, tx *sql.Tx, messageID int64, recipientUserIDs []int64) error {
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO user_message_recipients(message_id, recipient_id) VALUES (?, ?)")
	if err != nil {
		return err
	}

	for _, rec := range recipientUserIDs {
		_, err := stmt.ExecContext(ctx, messageID, rec)
		if err != nil {
			return err
		}
	}

	return nil
}
