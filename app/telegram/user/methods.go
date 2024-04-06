package user

import (
	"context"
	"fmt"
	"github.com/gotd/td/tg"
)

func (c *Client) GetUsername(phoneNum string) (string, error) {
	rawUsers, err := c.api.ContactsImportContacts(context.Background(), []tg.InputPhoneContact{{Phone: phoneNum}})
	if err != nil {
		return "", fmt.Errorf("failed to import contacts: %w", err)
	}
	users := rawUsers.GetUsers()

	if len(users) == 0 {
		return "", nil
	}

	user, ok := users[0].(*tg.User)
	if !ok {
		return "", nil
	}

	if err := c.removeContact(user); err != nil {
		return "", fmt.Errorf("failed to remove contact: %w", err)
	}

	return user.Username, nil
}

func (c *Client) GetSelfUsername() (string, error) {
	status, err := c.auth.Status(context.Background())
	if err != nil {
		return "", err
	}

	if !status.Authorized {
		return "", fmt.Errorf("user client is not authorized")
	}

	return status.User.Username, nil
}

func (c *Client) removeContact(user *tg.User) error {
	_, err := c.api.ContactsDeleteContacts(context.Background(), []tg.InputUserClass{&tg.InputUser{
		UserID:     user.ID,
		AccessHash: user.AccessHash,
	}})
	if err != nil {
		return fmt.Errorf("failed to delete contacts: %w", err)
	}

	return nil
}
