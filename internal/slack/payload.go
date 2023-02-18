package slack

//base model
type model struct {
	Type        string    `json:"type,omitempty"`
	Text        any       `json:"text,omitempty"`
	Title       any       `json:"title,omitempty"`
	Emoji       *bool     `json:"emoji,omitempty"`
	Style       string    `json:"style,omitempty"`
	Value       string    `json:"value,omitempty"`
	Accessory   Element   `json:"accessory,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	AltText     string    `json:"alt_text,omitempty"`
	Color       string    `json:"color,omitempty"`
	Blocks      []Block   `json:"blocks,omitempty"`
	Fields      []Element `json:"fields,omitempty"`
	Elements    []Element `json:"elements,omitempty"`
	Attachments []Payload `json:"attachments,omitempty"`
}

//payload
type Payload interface {
	AddHeader(text string) Payload
	AddImage(text string, url string, altText string) Payload
	AddSection(sectionCallback func(section Section) Section) Payload
	AddActions(actionsCallback func(actions Actions) Actions) Payload
	AddDivider() Payload
	AsNotification(title string, color string) Payload
}

func NewPayload() Payload {
	return &model{}
}

func (pl *model) AddHeader(text string) Payload {
	pl.Blocks = append(pl.Blocks, NewBlockHeader(text))
	return pl
}

func (pl *model) AddImage(text string, url string, altText string) Payload {
	pl.Blocks = append(pl.Blocks, NewElementImageWithTitle(text, url, altText))
	return pl
}

func (pl *model) AddSection(sectionCallback func(section Section) Section) Payload {
	pl.Blocks = append(pl.Blocks, sectionCallback(NewBlockSection()))
	return pl
}

func (pl *model) AddActions(actionsCallback func(actions Actions) Actions) Payload {
	pl.Blocks = append(pl.Blocks, actionsCallback(NewBlockActions()))
	return pl
}

func (pl *model) AddDivider() Payload {
	pl.Blocks = append(pl.Blocks, NewBlockDivider())
	return pl
}

func (pl *model) AsNotification(title string, color string) Payload {
	return &model{
		Text: title,
		Attachments: []Payload{
			&model{Color: color, Blocks: pl.Blocks},
		},
	}
}

//block
type Header interface{}
type Section interface {
	SetMarkdown(text string) Section
	SetText(text string, emoji bool) Section
	SetAccessoryImage(url string, altText string) Section
	SetAccessoryButton(text string, value string, style string) Section //style: "primary"/"danger"/""
	AddFieldMarkdown(text string) Section
	AddFieldText(text string, emoji bool) Section
	AddElementMarkdown(text string) Section
	AddElementText(text string, emoji bool) Section
}

type Actions interface {
	AddButton(text string, value string, style string) Actions //style: "primary"/"danger"/""
}

type Divider interface {
}

type Block interface {
}

//block-header
func NewBlockHeader(text string) Section {
	return &model{Type: "header", Text: NewElementText(text, true)}
}

//block-section
func NewBlockSection() Section {
	return &model{Type: "section"}
}

func (sc *model) SetMarkdown(text string) Section {
	sc.Text = NewElementMarkdown(text)
	return sc
}

func (sc *model) SetText(text string, emoji bool) Section {
	sc.Text = NewElementText(text, emoji)
	return sc
}

func (sc *model) SetAccessoryImage(url string, altText string) Section {
	sc.Accessory = NewElementImage(url, altText)
	return sc
}

func (sc *model) SetAccessoryButton(text string, value string, style string) Section {
	sc.Accessory = NewElementButton(text, value, style)
	return sc
}

func (sc *model) AddFieldMarkdown(text string) Section {
	sc.Fields = append(sc.Fields, NewElementMarkdown(text))
	return sc
}

func (sc *model) AddFieldText(text string, emoji bool) Section {
	sc.Fields = append(sc.Fields, NewElementText(text, emoji))
	return sc
}

func (sc *model) AddElementMarkdown(text string) Section {
	sc.Elements = append(sc.Elements, NewElementMarkdown(text))
	return sc
}

func (sc *model) AddElementText(text string, emoji bool) Section {
	sc.Elements = append(sc.Elements, NewElementText(text, emoji))
	return sc
}

//block-actions
func NewBlockActions() Actions {
	return &model{Type: "actions"}
}

func (ac *model) AddButton(text string, value string, style string) Actions {
	ac.Elements = append(ac.Elements, NewElementButton(text, value, style))
	return ac
}

//block-divider
func NewBlockDivider() Divider {
	return &model{Type: "divider"}
}

//element
type Element interface {
}

func NewElementButton(text string, value string, style string) Element {
	return &model{Type: "button", Text: NewElementText(text, true), Value: value, Style: style}
}

func NewElementImage(url string, altText string) Element {
	return &model{Type: "image", ImageURL: url, AltText: altText}
}

func NewElementImageWithTitle(title string, url string, altText string) Element {
	return &model{Type: "image", Title: NewElementText(title, false), ImageURL: url, AltText: altText}
}

func NewElementMarkdown(text string) Element {
	return &model{Type: "mrkdwn", Text: text}
}

func NewElementText(text string, emoji bool) Element {
	return &model{Type: "plain_text", Text: text, Emoji: &emoji}
}
