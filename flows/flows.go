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

// Package flows provides a set of functions for creating and sending FlowJSON messages.
package flows

//{
// "version": "3.1",
// "data_api_version": "3.0",
// "routing_model": {"MY_FIRST_SCREEN": ["MY_SECOND_SCREEN"] },
// "screens": [...]
//}

const (
	LowestSupportedVersion        = "3.1"
	LowestSupportedDataApiVersion = "3.0"
)

type (
	Flow struct {
		Version        string              `json:"version,omitempty"`
		DataAPIVersion string              `json:"data_api_version,omitempty"`
		RoutingModel   map[string][]string `json:"routing_model,omitempty"`
		Screens        []*Screen           `json:"screens,omitempty"`
	}

	Screen struct {
		ID            string  `json:"id,omitempty"`
		Terminal      bool    `json:"terminal,omitempty"`
		Success       bool    `json:"success,omitempty"`
		Title         string  `json:"title,omitempty"`
		RefreshOnBack bool    `json:"refresh_on_back,omitempty"`
		Data          Data    `json:"data,omitempty"`
		Layout        *Layout `json:"layout,omitempty"`
	}

	DataValue struct {
		Type    string `json:"type,omitempty"`
		Example string `json:"__example__,omitempty"`
	}

	Data map[string]DataValue

	Layout struct {
		Type string `json:"type,omitempty"`
		// Children a list of components that are rendered in the layout.
	}

	// Components
	//Text (Heading, Subheading, Caption, Body)
	//
	//TextEntry
	//
	//CheckboxGroup
	//
	//RadioButtonsGroup
	//
	//Footer
	//
	//OptIn
	//
	//Dropdown
	//
	//EmbeddedLink
	//
	//DatePicker
	//
	//Image

	TextComponent struct {
		Type          string     `json:"type,omitempty"`
		Text          string     `json:"text,omitempty"`
		Visible       bool       `json:"visible,omitempty"`
		FontWeight    FontWeight `json:"font-weight,omitempty"`
		Strikethrough bool       `json:"strikethrough,omitempty"`
	}

	TextInputComponent struct {
		Type          string  `json:"type,omitempty"`
		Label         string  `json:"label,omitempty"`
		InputType     string  `json:"input-type,omitempty"` //enum	{'text','number','email', 'password', 'passcode', 'phone'}
		Required      bool    `json:"required,omitempty"`
		MinChars      int     `json:"min-chars,omitempty"`
		MaxChars      int     `json:"max-chars,omitempty"`
		HelperText    string  `json:"helper-text,omitempty"`
		Name          string  `json:"name,omitempty"`
		Visible       bool    `json:"visible,omitempty"`
		OnClickAction *Action `json:"on-click-action,omitempty"`
	}

	TextAreaComponent struct {
		Type          string  `json:"type,omitempty"`
		Label         string  `json:"label,omitempty"`
		Required      bool    `json:"required,omitempty"`
		MaxLength     int     `json:"max-length,omitempty"`
		Name          string  `json:"name,omitempty"`
		HelperText    string  `json:"helper-text,omitempty"`
		Enabled       bool    `json:"enabled,omitempty"`
		Visible       bool    `json:"visible,omitempty"`
		OnClickAction *Action `json:"on-click-action,omitempty"`
	}

	CheckboxGroupComponent struct {
		Type             string          `json:"type,omitempty"`
		DataSource       []*DropDownData `json:"data-source,omitempty"`
		Name             string          `json:"name,omitempty"`
		MinSelectedItems int             `json:"min-selected-items,omitempty"`
		MaxSelectedItems int             `json:"max-selected-items,omitempty"`
		Enabled          bool            `json:"enabled,omitempty"`
		Label            string          `json:"label,omitempty"`
		Required         bool            `json:"required,omitempty"`
		Visible          bool            `json:"visible,omitempty"`
		OnSelectAction   *Action         `json:"on-select-action,omitempty"`
	}

	RadioButtonsGroupComponent struct {
		Type           string          `json:"type,omitempty"`
		DataSource     []*DropDownData `json:"data-source,omitempty"`
		Name           string          `json:"name,omitempty"`
		Enabled        bool            `json:"enabled,omitempty"`
		Label          string          `json:"label,omitempty"`
		Required       bool            `json:"required,omitempty"`
		Visible        bool            `json:"visible,omitempty"`
		OnSelectAction *Action         `json:"on-select-action,omitempty"`
	}

	FooterComponent struct {
		Type          string  `json:"type,omitempty"`
		Label         string  `json:"label,omitempty"`
		LeftCaption   string  `json:"left-caption,omitempty"`
		CenterCaption string  `json:"center-caption,omitempty"`
		RightCaption  string  `json:"right-caption,omitempty"`
		Enabled       bool    `json:"enabled,omitempty"`
		OnClickAction *Action `json:"on-click-action,omitempty"`
	}

	OptInComponent struct {
		Type          string  `json:"type,omitempty"`
		Label         string  `json:"label,omitempty"`
		Required      bool    `json:"required,omitempty"`
		Name          string  `json:"name,omitempty"`
		OnClickAction *Action `json:"on-click-action,omitempty"`
		Visible       bool    `json:"visible,omitempty"`
	}

	DropdownComponent struct {
		Type           string          `json:"type,omitempty"`
		Label          string          `json:"label,omitempty"`
		DataSource     []*DropDownData `json:"data-source,omitempty"`
		Required       bool            `json:"required,omitempty"`
		Enabled        bool            `json:"enabled,omitempty"`
		Visible        bool            `json:"visible,omitempty"`
		OnSelectAction *Action         `json:"on-select-action,omitempty"`
	}

	DropDownData struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Enabled     bool   `json:"enabled,omitempty"`
		Description string `json:"description,omitempty"`
		Metadata    string `json:"metadata,omitempty"`
	}

	EmbeddedLinkComponent struct {
		Type          string  `json:"type,omitempty"`
		Text          string  `json:"text,omitempty"`
		OnClickAction *Action `json:"on-click-action,omitempty"`
		Visible       bool    `json:"visible,omitempty"`
	}

	DatePickerComponent struct {
		Type        string   `json:"type,omitempty"`
		Label       string   `json:"label,omitempty"`
		MinDate     string   `json:"min-date,omitempty"`
		MaxDate     string   `json:"max-date,omitempty"`
		Name        string   `json:"name,omitempty"`
		Unavailable []string `json:"unavailable-dates,omitempty"`
		Visible     bool     `json:"visible,omitempty"`
		HelperText  string   `json:"helper-text,omitempty"`
		Enabled     bool     `json:"enabled,omitempty"`
		OnSelect    *Action  `json:"on-select-action,omitempty"`
	}

	ImageComponent struct {
		Type        string  `json:"type,omitempty"`
		Src         string  `json:"src,omitempty"`
		Width       int     `json:"width,omitempty"`
		Height      int     `json:"height,omitempty"`
		ScaleType   string  `json:"scale-type,omitempty"`
		AspectRatio float64 `json:"aspect-ratio,omitempty"`
		AltText     string  `json:"alt-text,omitempty"`
	}

	Action struct {
		Name string `json:"name,omitempty"`
	}
)

type ImageScaleType string

const (
	ScaleTypeCover   ImageScaleType = "cover"
	ScaleTypeContain ImageScaleType = "contain"
)

const (
	ComponentTypeFooter = "Footer"
	ComponentTypeOptIn  = "OptIn"
)

type TextComponentType string

const (
	TextComponentTypeHeading    TextComponentType = "TextHeading"
	TextComponentTypeSubheading TextComponentType = "TextSubheading"
	TextComponentTypeBody       TextComponentType = "TextBody"
	TextComponentTypeCaption    TextComponentType = "TextCaption"
)

type FontWeight string

const (
	FontWeightBold       FontWeight = "bold"
	FontWeightItalic     FontWeight = "italic"
	FontWeightBoldItalic FontWeight = "bold_italic"
	FontWeightNormal     FontWeight = "normal"
)
