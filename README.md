# WhatsApp Cloud API Go Client

[![GoDoc](https://godoc.org/github.com/piusalfred/whatsapp?status.svg)](https://godoc.org/github.com/piusalfred/whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/piusalfred/whatsapp)](https://goreportcard.com/report/github.com/piusalfred/whatsapp)
![Status](https://img.shields.io/badge/status-alpha-red)

**Note:** This library is currently in alpha and not yet stable. Breaking changes may occur.

---

A highly configurable Go client library for the [WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp), providing functionalities for:

- **Webhooks** (Business and Message)
- **Messaging**
- **QR Code Management**
- **Phone Number Management**
- **Media Management**

## Features

- **Webhooks Handling**: Easily set up webhook endpoints for both business and message notifications with support for middleware and signature verification.
- **Messaging**: Send and receive messages, including text, images, and interactive messages.
- **QR Code Management**: Create and manage QR codes linked to your WhatsApp business account.
- **Phone Number Management**: Manage phone numbers associated with your WhatsApp business account.
- **Media Management**: Upload, download, and manage media files like images, videos, and documents.

## Installation

To install the package, run:

```bash
go get github.com/yourusername/whatsapp-cloud-api
```


## Prerequisites

- [**Whatsapp Cloud API Get Started Guide**](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started) 