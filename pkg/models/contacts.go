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

package models

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

	Url struct { ////nolint: revive,stylecheck
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	Urls []*Url

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
)
