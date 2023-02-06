package whatsapp

// The phone numbers in Cloud API requests can be provided in any dialable format, as long as they include their country code.

// Here are some examples of supported phone number formats:

// "1-000-000-0000"
// "1 (000) 000-0000"
// "1 000 000 0000"
// "1 (000) 000 0000"

const RegexPhoneNumber = `^(\+?1)?[ .-]?\(?[2-9]\d{2}\)?[ .-]?\d{3}[ .-]?\d{4}$`

func FormatPhoneNumber(phone string) string {

}
