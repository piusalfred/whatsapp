Remeber to replace WhatsApp Business API credentials with your own.

```go

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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/piusalfred/whatsapp"
	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

func main() {
	writer := os.Stdout
	phoneID := "1XXXXXXX2711"
	whatsappID := "1XXXXXX4XXX"
	token := "EAALLrT0o"
	hook := whttp.DebugHook(writer)
	client := whatsapp.NewClient(
		whatsapp.WithHTTPClient(http.DefaultClient),
		whatsapp.WithPhoneNumberID(phoneID),
		whatsapp.WithBusinessAccountID(whatsappID),
		whatsapp.WithAccessToken(token),
		whatsapp.WithHooks(hook))

	recipient := "255767001828"
	ctx := context.Background()

	// Send a Template ( We will use the default template called hello_world)
	tmpl := &whatsapp.Template{
		LanguageCode:   "en_US",
		LanguagePolicy: "",
		Name:           "hello_world",
		Components:     nil,
	}

	response, err := client.SendTemplate(ctx, recipient, tmpl)
	if err != nil {
		fmt.Printf("error sending template message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("response: %+v\n", response)

	// Sending a text message
	message := &whatsapp.TextMessage{
		Message:    "Hello there!!!\n 😺Find me at https://github.com/piusalfred/whatsapp 👩🏻‍🦰",
		PreviewURL: true,
	}

	response, err = client.SendTextMessage(ctx, recipient, message)
	if err != nil {
		fmt.Printf("error sending text message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("response: %+v\n", response)

	// Estadio Santiago Bernabeu
	location := &models.Location{
		Longitude: -3.688344,
		Latitude:  40.453053,
		Name:      "Estadio Santiago Bernabeu",
		Address:   "Av. de Concha Espina, 1, 28036 Madrid, Spain",
	}
	//
	response, err = client.SendLocationMessage(ctx, recipient, location)
	if err != nil {
		fmt.Printf("error sending location message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("response: %+v\n", response)

	name := &models.Name{
		FormattedName: "John Doe Jr",
		FirstName:     "John",
		LastName:      "Doe",
		MiddleName:    "Jackson",
		Suffix:        "Jr",
		Prefix:        "Dr",
	}

	birthday := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	homePhone := &models.Phone{
		Phone: "555-1234",
		Type:  "home",
		WaID:  "",
	}

	workPhone := &models.Phone{
		Phone: "555-1234",
		Type:  "work",
	}

	email := &models.Email{
		Email: "thejohndoejr@example.dummy",
		Type:  "work",
	}

	organization := &models.Org{
		Company:    "John Doe and Sons Co. LTD",
		Department: "Serious Stuffs Department",
		Title:      "Commander In Chief",
	}
	//
	address := &models.Address{
		Street:      "123 Main St",
		City:        "Anytown",
		State:       "CA",
		Zip:         "12345",
		Country:     "United States",
		CountryCode: "US",
		Type:        "home",
	}
	//
	contact1 := models.NewContact("John Doe Jr",
		models.WithContactName(name),
		models.WithContactBirthdays(birthday),
		models.WithContactPhones(homePhone, workPhone),
		models.WithContactEmails(email),
		models.WithContactOrganization(organization),
		models.WithContactAddresses(address),
	)

	contacts := []*models.Contact{contact1}
	//
	response, err = client.SendContacts(ctx, recipient, contacts)
	if err != nil {
		fmt.Printf("error sending contacts: %v\n", err)
		os.Exit(1)
	}
	//
	//fmt.Printf("response: %+v\n", response)
	//
	//// Sending an image
	media := &whatsapp.MediaMessage{
		Type:      whatsapp.MediaTypeImage,
		MediaLink: "https://cdn.pixabay.com/photo/2022/12/04/16/17/leaves-7634894_1280.jpg",
	}
	//
	response, err = client.SendMedia(ctx, recipient, media, nil)
	//
	if err != nil {
		fmt.Printf("error sending media: %v\n", err)
		os.Exit(1)
	}
	//
	fmt.Printf("response: %+v\n", response)
	//
	//
	header := &models.InteractiveHeader{
		Text: "choose what you want to do",
		Type: "image",
		Image: &models.Media{
			Link: "https://cdn.pixabay.com/photo/2022/12/04/16/17/leaves-7634894_1280.jpg",
		},
	}

	bodyText := "Real Madrid is one of the most successful football clubs in the world, with a rich history and a proud tradition. Founded in 1902, the club has won countless domestic and international titles over the years, cementing its place among the greatest teams of all time. With a squad of some of the most talented and skilled players in the world, Real Madrid has consistently dominated the sport, winning a record 14 European Champions League titles and 35 La Liga titles. The club's legendary players, such as Cristiano Ronaldo, Zinedine Zidane, Alfredo Di Stefano, and Raul, have left an indelible mark on the history of the game, and their legacy continues to inspire generations of football fans around the world. With a loyal and passionate fanbase, state-of-the-art facilities, and an unwavering commitment to excellence, Real Madrid is truly one of the greatest clubs in the history of football."

	replyButton1 := &models.InteractiveReplyButton{
		ID:    "btn0001",
		Title: "Real Madrid",
	}

	replyButton2 := &models.InteractiveReplyButton{
		ID:    "btn0002",
		Title: "Barcelona",
	}

	replyButton3 := &models.InteractiveReplyButton{
		ID:    "btn0003",
		Title: "Atletico Madrid",
	}

	buttonsList := models.CreateInteractiveRelyButtonList(
		replyButton1, replyButton2, replyButton3)

	action := &models.InteractiveAction{
		Button:            "",
		Buttons:           buttonsList,
		CatalogID:         "",
		ProductRetailerID: "",
		Sections:          nil,
	}

	interactive := models.Interactive{
		Type:   models.InteractiveMessageButton,
		Action: action,
		Header: header,
	}

	models.WithInteractiveFooter("https://github.com/piusalfred/whatsapp")(&interactive)
	models.WithInteractiveBody(bodyText)(&interactive)

	response, err = client.SendInteractiveMessage(ctx, recipient, &interactive)
	if err != nil {
		fmt.Printf("error sending interactive message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("response: %+v\n", response)
}

```