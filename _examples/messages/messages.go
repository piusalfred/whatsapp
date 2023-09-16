package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/piusalfred/whatsapp"
	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

func main() {
	err := setup()
	if err != nil {
		fmt.Printf("error setting up: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}
}

const quitMessage = `Press Ctrl+C to quit at any point to quit`

var ErrInterrupted = fmt.Errorf("interrupted")

// setup runs a simple interactive commandline tool to show that you have successfully
// configured your whatsapp business account to send messages.
func setup() error {
	writer := os.Stdout
	if _, err2 := writer.WriteString(quitMessage); err2 != nil {
		return err2
	}
	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	hook := whttp.LogRequestHook(logger)
	respHook := whttp.LogResponseHook(logger)

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		sig := <-sigChan
		err := fmt.Errorf("%w: received signal: %v", ErrInterrupted, sig)
		cancel(err)
	}()
	configer := &configer{reader: bufio.NewReader(os.Stdin)}
	config, err := configer.Read(ctx)
	if err != nil {
		return err
	}
	client, err := whatsapp.NewClientWithConfig(config,
		whatsapp.WithBaseClient(&whatsapp.BaseClient{Client: whttp.NewClient(
			whttp.WithHTTPClient(http.DefaultClient),
			whttp.WithRequestHooks(hook),
			whttp.WithResponseHooks(respHook),
		)}))
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the recipient phone number (this number must be registered in FB portal): ")
	recipient, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Println("Sending Template Message (make sure you reply): ")
	// Send a Template ( We will use the default template called hello_world)
	tmpl := &whatsapp.Template{
		LanguageCode:   "en_US",
		LanguagePolicy: "",
		Name:           "hello_world",
		Components:     nil,
	}

	response, err := client.SendTemplate(ctx, recipient, tmpl)
	if err != nil {
		return err
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "send template", slog.Group("response", response))

	message := &whatsapp.TextMessage{
		Message:    "😺Find me at https://github.com/piusalfred/whatsapp 👩🏻‍🦰",
		PreviewURL: true,
	}

	response, err = client.SendTextMessage(ctx, recipient, message)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "send template",
		slog.Group("response", response))

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

	contacts := []*models.Contact{contact1, contact1}

	response, err = client.SendContacts(ctx, recipient, contacts)
	if err != nil {
		fmt.Printf("error sending contacts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("response: %+v\n", response)

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

	bodyText := `
			Real Madrid is one of the most successful football
		clubs in the world, with a rich history and a proud tradition.
Founded in 1902, the club has won countless domestic and international
titles over the years, cementing its place among the greatest teams of all time.
With a squad of some of the most talented and skilled players in the world,
Real Madrid has consistently dominated the sport, winning a record 14 European 
Champions League titles and 35 La Liga titles. The club's legendary players,
such as Cristiano Ronaldo, Zinedine Zidane, Alfredo Di Stefano, and Raul,
have left an indelible mark on the history of the game, and their legacy
continues to inspire generations of football fans around the world. With a 
loyal and passionate fanbase, state-of-the-art facilities, and an unwavering
ccommitment to excellence, 
Real Madrid is truly one of the greatest clubs in the history of football`

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

	select {
	case <-ctx.Done():
		return fmt.Errorf("interupted: %w", ctx.Err())

	default:
		return nil
	}
}

var _ whatsapp.ConfigReader = (*configer)(nil)

// configer implements whatsapp.ConfigReaderFunc it basically asks user to enter the required
// configuration values via the commandline.
type configer struct {
	reader *bufio.Reader
}

func (c *configer) Read(ctx context.Context) (*whatsapp.Config, error) {
	doneChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)
	var config whatsapp.Config

	go func() {
		fmt.Println("Enter your access token: ")
		token, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}
		config.AccessToken = strings.TrimSpace(token)

		fmt.Println("Enter your phone number ID: ")
		phoneID, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}

		config.PhoneNumberID = strings.TrimSpace(phoneID)

		fmt.Println("Enter your business account ID: ")
		businessID, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}

		config.BusinessAccountID = strings.TrimSpace(businessID)

		fmt.Println("Enter API version:(Lowest version is v16.0) ")
		version, err := c.reader.ReadString('\n')
		if err != nil {
			errChan <- err

			return
		}

		config.Version = strings.TrimSpace(version)

		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
		return &config, nil

	case err := <-errChan:
		return nil, err

	case <-ctx.Done():
		return nil, fmt.Errorf("interrupted: %w", ctx.Err())
	}
}
