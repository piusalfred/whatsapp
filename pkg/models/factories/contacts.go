package factories

import (
	"time"

	"github.com/piusalfred/whatsapp/pkg/models"
)

type (
	ContactOption func(*models.Contact)
)

func NewContact(name string, options ...ContactOption) *models.Contact {
	contact := &models.Contact{
		Name: &models.Name{
			FormattedName: name,
		},
	}
	for _, option := range options {
		option(contact)
	}

	return contact
}

func WithContactName(name *models.Name) ContactOption {
	return func(c *models.Contact) {
		c.Name = name
	}
}

func WithContactAddresses(addresses ...*models.Address) ContactOption {
	return func(c *models.Contact) {
		c.Addresses = addresses
	}
}

func WithContactOrganization(organization *models.Org) ContactOption {
	return func(c *models.Contact) {
		c.Org = organization
	}
}

func WithContactURLs(urls ...*models.Url) ContactOption {
	return func(c *models.Contact) {
		c.Urls = urls
	}
}

func WithContactPhones(phones ...*models.Phone) ContactOption {
	return func(c *models.Contact) {
		c.Phones = phones
	}
}

func WithContactBirthdays(birthday time.Time) ContactOption {
	return func(c *models.Contact) {
		// should be formatted as YYYY-MM-DD
		bd := birthday.Format("2006-01-02")
		c.Birthday = bd
	}
}

func WithContactEmails(emails ...*models.Email) ContactOption {
	return func(c *models.Contact) {
		c.Emails = emails
	}
}

// NewContacts ...
func NewContacts(contacts []*models.Contact) models.Contacts {
	if contacts != nil {
		return models.Contacts(contacts)
	}

	return nil
}
