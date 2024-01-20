/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

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
		return contacts
	}

	return nil
}
