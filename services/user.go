package services

import (
	"database/sql"
	"log"

	"stocks-info-channel/model"

	"github.com/lib/pq"
)

// GetOrCreateUser fetches a user by phone or creates a new one
func GetOrCreateUser(db *sql.DB, phone string) (*model.User, error) {
	var user model.User
	err := db.QueryRow(`
		SELECT id, phone_number, name, last_message_time,
		       last_two_messages_to_user, last_two_messages_from_user,
		       is_subscribed, subscribed_stocks
		FROM users WHERE phone_number = $1
	`, phone).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.LastMessageTime,
		&user.LastTwoMessagesToUser,
		&user.LastTwoMessagesFromUser,
		&user.IsSubscribed,
		&user.SubscribedStocks,
	)

	if err == sql.ErrNoRows {
		log.Println("User not found, creating new user")
		return createUser(db, phone)
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

// createUser inserts a new user with default values
func createUser(db *sql.DB, phone string) (*model.User, error) {
	user := &model.User{
		PhoneNumber:             phone,
		Name:                    sql.NullString{},
		LastMessageTime:         sql.NullTime{},
		LastTwoMessagesToUser:   pq.StringArray{},
		LastTwoMessagesFromUser: pq.StringArray{},
		IsSubscribed:            true,
		SubscribedStocks:        pq.StringArray{},
	}

	err := db.QueryRow(`
		INSERT INTO users (phone_number, name, last_message_time,
			last_two_messages_to_user, last_two_messages_from_user,
			is_subscribed, subscribed_stocks)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, user.PhoneNumber, user.Name, user.LastMessageTime,
		user.LastTwoMessagesToUser, user.LastTwoMessagesFromUser,
		user.IsSubscribed, user.SubscribedStocks).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}
