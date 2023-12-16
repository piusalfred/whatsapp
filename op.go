package whatsapp

// Op represents a request to the WhatsApp API an operation.
type Op string

const (
	OpSendMessage  Op = "send message"
	OpSendMedia    Op = "send media"
	OpSendContact  Op = "send contact"
	OpSendLocation Op = "send location"
	OpSendDocument Op = "send document"
	OpSendTemplate Op = "send template"
	OpReadMessage  Op = "read message"
	OpUploadMedia  Op = "upload media"
)
