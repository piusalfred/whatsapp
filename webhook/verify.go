package webhook

import (
	"context"
	"net/http"
)

type (
	VerificationRequest struct {
		Mode      string `json:"hub.mode"`
		Challenge string `json:"hub.challenge"`
		Token     string `json:"hub.verify_token"`
	}

	// SubscriptionVerifier is a function that processes the verification request.
	// The function must return nil if the verification request is valid.
	// It mainly checks if hub.mode is set to subscribe and if the hub.verify_token matches
	// the one set in the App Dashboard.
	SubscriptionVerifier func(context.Context, *VerificationRequest) error
)

// VerifySubscriptionHandler verifies the subscription to the webhook.
// Your endpoint must be able to process two types of HTTPS requests: Verification Requests and Event Notifications.
// Since both requests use HTTPs, your server must have a valid TLS or SSL certificate correctly configured and
// installed. Self-signed certificates are not supported.
// Anytime you configure the Webhooks product in your App Dashboard, we'll send a GET request to your endpoint URL.
// Verification requests include the following query string parameters, appended to the end of your endpoint URL.
// They will look something like this:
// GET https://www.your-clever-domain-name.com/webhooks?
// hub.mode=subscribe&
// hub.challenge=1158201444&
// hub.verify_token=meatyhamhock
// hub.mode This value will always be set to subscribe.
// hub.challenge An int you must pass back to us.
// hub.verify_token A string that that we grab from the Verify Token field in your app's App Dashboard. You will set
// this string when you complete the Webhooks configuration settings steps.
//
// Whenever your endpoint receives a verification request, it must:
//
// - Verify that the hub.verify_token value matches the string you set in the Verify Token field when you configure
// the Webhooks product in your App Dashboard (you haven't set up this token string yet).
//
// - Respond with the hub.challenge value. If you are in your App Dashboard and configuring your Webhooks product
// (and thus, triggering a Verification Request), the dashboard will indicate if your endpoint validated the request
//
//	correctly. If you are using the Graph API's /app/subscriptions endpoint to configure the Webhooks product, the API
//
// will indicate success or failure with a response.
func VerifySubscriptionHandler(verifier SubscriptionVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the query parameters from the request.
		q := r.URL.Query()
		mode := q.Get("hub.mode")
		challenge := q.Get("hub.challenge")
		token := q.Get("hub.verify_token")
		if err := verifier(r.Context(), &VerificationRequest{
			Mode:      mode,
			Challenge: challenge,
			Token:     token,
		}); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
	}
}
