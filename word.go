// Handling different type of words
package gofakelib

var (
	wordsLoader DataLoader
)

func init() {
	wordsLoader.Init("words.json")
}

type WordsData struct {
	Adjectives         []string `json:"adjectives"`
	AdjectivesPositive []string `json:"adjectives_positive"`
	AdjectivesNegative []string `json:"adjectives_negative"`
	Adverbs            []string `json:"adverbs"`
}

// Return a random adjective
func (f *Faker) Adjective() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adjectives"))
}

// Return a random positive adjective
func (f *Faker) AdjectivePositive() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adjectives_positive"))
}

// Return a random negative adjective
func (f *Faker) AdjectiveNegative() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adjectives_negative"))
}

// Return a random adverb
func (f *Faker) Adverb() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adverbs"))
}
