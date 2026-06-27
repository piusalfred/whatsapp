//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package message

import "time"

type (
	Address struct {
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Type        string `json:"type"`
	}

	Addresses []*Address

	Email struct {
		Email string `json:"email"`
		Type  string `json:"type"`
	}

	Emails []*Email

	Name struct {
		FormattedName string `json:"formatted_name"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		MiddleName    string `json:"middle_name"`
		Suffix        string `json:"suffix"`
		Prefix        string `json:"prefix"`
	}

	Org struct {
		Company    string `json:"company"`
		Department string `json:"department"`
		Title      string `json:"title"`
	}

	Phone struct {
		Phone string `json:"phone"`
		Type  string `json:"type"`
		WaID  string `json:"wa_id,omitempty"`
	}

	Phones []*Phone

	URL struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	Urls []*URL

	Contact struct {
		Addresses Addresses `json:"addresses,omitempty"`
		Birthday  string    `json:"birthday"`
		Emails    Emails    `json:"emails,omitempty"`
		Name      *Name     `json:"name"`
		Org       *Org      `json:"org"`
		Phones    Phones    `json:"phones,omitempty"`
		Urls      Urls      `json:"urls,omitempty"`
	}

	Contacts []*Contact

	ContactOption func(*Contact)
)

func NewContact(options ...ContactOption) *Contact {
	contact := &Contact{}
	for _, option := range options {
		option(contact)
	}

	return contact
}

func WithContactName(name *Name) ContactOption {
	return func(c *Contact) {
		c.Name = name
	}
}

func WithContactAddresses(addresses ...*Address) ContactOption {
	return func(c *Contact) {
		c.Addresses = addresses
	}
}

func WithContactOrganization(organization *Org) ContactOption {
	return func(c *Contact) {
		c.Org = organization
	}
}

func WithContactURLs(urls ...*URL) ContactOption {
	return func(c *Contact) {
		c.Urls = urls
	}
}

func WithContactPhones(phones ...*Phone) ContactOption {
	return func(c *Contact) {
		c.Phones = phones
	}
}

func WithContactBirthdays(birthday time.Time) ContactOption {
	return func(c *Contact) {
		// should be formatted as YYYY-MM-DD
		bd := birthday.Format(time.DateOnly)
		c.Birthday = bd
	}
}

func WithContactEmails(emails ...*Email) ContactOption {
	return func(c *Contact) {
		c.Emails = emails
	}
}
