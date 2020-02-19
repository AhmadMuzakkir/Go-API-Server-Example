package mysql

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	testDbHost   = "127.0.0.1"
	testDbPort   = 3307
	testUsername = "root"
	testPassword = "test_password"
	testDatabase = "test_database"
)

func getTestStore(t *testing.T) (*Store, func()) {
	s, err := Connect(testDbHost, testDbPort, testUsername, testPassword, testDatabase)
	if err != nil {
		t.Fatalf(
			"error connecting to the test mysql database: address=%q, port=%d username=%q, password=%q, database=%q: %s",
			testDbHost, testDbPort, testUsername, testPassword, testDatabase, err,
		)
	}

	driver, err := mysql.WithInstance(s.db, &mysql.Config{
		DatabaseName: testDatabase,
	})
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"mysql", driver)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		assert.FailNow(t, err.Error())
	}

	if err := m.Up(); !assert.NoError(t, err) {
		t.FailNow()
	}

	cleanup := func() {
		_ = s.db.Close()
	}

	return s, cleanup
}

// Test User and Token tables
func TestUseAndToken(t *testing.T) {
	s, cleanup := getTestStore(t)
	defer cleanup()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = s.userStore.Create(context.Background(), "username", string(passwordHash))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user, err := s.userStore.GetByUsername(context.Background(), "username")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, "username", user.Username)
	assert.Equal(t, string(passwordHash), user.PasswordHash)

	token := make([]byte, 32)
	_, err = rand.Read(token)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	updatedAt := time.Now()

	err = s.tokenStore.Create(context.Background(), user.ID, fmt.Sprintf("%x", token), updatedAt)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	userToken, err := s.tokenStore.GetUserID(context.Background(), fmt.Sprintf("%x", token))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, user.ID, userToken.UserID)
	assert.Equal(t, updatedAt.Nanosecond(), userToken.UpdatedAt.Nanosecond())
}

func TestMessage(t *testing.T) {
	s, cleanup := getTestStore(t)
	defer cleanup()

	user1 := addUser(t, s, "username1", "password1")
	user2 := addUser(t, s, "username2", "password2")
	user3 := addUser(t, s, "username3", "password3")

	// Add message

	msg := store.Message{
		Content:      "message content",
		SenderID:     user1.ID,
		SentDateTime: time.Now(),
	}

	err := s.messageStore.Create(context.Background(), msg, []int64{user2.ID, user3.ID})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Test the message for user2

	user2Msg, err := s.messageStore.Get(context.Background(), user2.ID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Len(t, user2Msg, 1) {
		t.FailNow()
	}

	assert.Equal(t, msg.Content, user2Msg[0].Content)
	assert.Equal(t, msg.SentDateTime.Nanosecond(), user2Msg[0].SentDateTime.Nanosecond())
	assert.Equal(t, msg.SentDateTime.Nanosecond(), user2Msg[0].UpdatedDateTime.Nanosecond())
	assert.Equal(t, "username1", user2Msg[0].Sender)

	// Update the message, and remove user3 from the recipients

	msg.ID = user2Msg[0].ID
	msg.Content = "updated message content"
	msg.UpdatedDateTime = time.Now()

	err = s.messageStore.Update(context.Background(), msg, []int64{user2.ID})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Test the updated message for user2

	user2Msg, err = s.messageStore.Get(context.Background(), user2.ID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Len(t, user2Msg, 1) {
		t.FailNow()
	}

	assert.Equal(t, msg.Content, user2Msg[0].Content)
	assert.Equal(t, msg.SentDateTime.Nanosecond(), user2Msg[0].SentDateTime.Nanosecond())
	assert.Equal(t, msg.UpdatedDateTime.Nanosecond(), user2Msg[0].UpdatedDateTime.Nanosecond())

	// Test that user3 will not getting the message

	user3Msg, err := s.messageStore.Get(context.Background(), user3.ID)
	if !assert.Len(t, user3Msg, 0) {
		t.FailNow()
	}
}

func addUser(t *testing.T, s *Store, username, password string) *store.User {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = s.userStore.Create(context.Background(), username, string(passwordHash))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	user, err := s.userStore.GetByUsername(context.Background(), username)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return user
}