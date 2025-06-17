// Generate different types of emojis
package fakery

import (
	source "fakery/source"
)

// Structure representing an emoji
type Emoji struct {
	Symbol      string `json:"symbol"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Alias       string `json:"alias"`
	Base
}

func (e Emoji) String() string {
	return e.Base.String(e)
}

// Return random emoji symbol
func (f *Faker) EmojiSymbol() string {
	idx := f.IntRange(len(source.Emojis))
	return source.Emojis[idx].Symbol
}

// Return random emoji category
func (f *Faker) EmojiCategory() string {
	idx := f.IntRange(len(source.Emojis))
	return source.Emojis[idx].Category
}

// Return random emoji description
func (f *Faker) EmojiDescription() string {
	idx := f.IntRange(len(source.Emojis))
	// manage description
	desc := source.Emojis[idx].Description
	// Lower and capitalize description string
	return f.Capitalize(desc)
}

// Return random emoji tags (aliases)
func (f *Faker) EmojiAlias() string {
	idx := f.IntRange(len(source.Emojis))
	aliases := source.Emojis[idx].Aliases
	if len(aliases) == 1 {
		return aliases[0]
	}
	return f.RandomString(aliases)
}

// Return random emoji struct
func (f *Faker) Emoji() *Emoji {
	idx := f.IntRange(len(source.Emojis))
	emojiData := source.Emojis[idx]

	return &Emoji{
		Symbol:      emojiData.Symbol,
		Category:    emojiData.Category,
		Description: f.Capitalize(emojiData.Description),
		Alias:       f.RandomString(emojiData.Aliases),
	}
}
