package whatsapp

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

	Addresses struct {
		Addresses []Address `json:"addresses"`
	}

	Email struct {
		Email string `json:"email"`
		Type  string `json:"type"`
	}

	Emails struct {
		Emails []Email `json:"emails"`
	}

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

	Phones struct {
		Phones []Phone `json:"phones"`
	}

	Url struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	Urls struct {
		Urls []Url `json:"urls"`
	}

	Contact struct {
		Addresses Addresses `json:"addresses,omitempty"`
		Birthday  string    `json:"birthday"`
		Emails    Emails    `json:"emails,omitempty"`
		Name      Name      `json:"name"`
		Org       Org       `json:"org"`
		Phones    Phones    `json:"phones,omitempty"`
		Urls      Urls      `json:"urls,omitempty"`
	}

	Contacts struct {
		Contacts []Contact `json:"contacts"`
	}
)
