package whatsapp

import "context"

type (

	// Location represents a location
	//
	//	"location": {
	//		"longitude": LONG_NUMBER,
	//		"latitude": LAT_NUMBER,
	//		"name": LOCATION_NAME,
	//		"address": LOCATION_ADDRESS
	//	  }
	Location struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Name      string  `json:"name"`
		Address   string  `json:"address"`
	}

	SendLocationFunc func(ctx context.Context, params *RequestParams)
)
