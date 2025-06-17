// Handling different type of words
package fakery

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
func (f *Fakery) Adjective() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adjectives"))
}

// Return a random positive adjective
func (f *Fakery) AdjectivePositive() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adjectives_positive"))
}

// Return a random negative adjective
func (f *Fakery) AdjectiveNegative() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adjectives_negative"))
}

// Return a random adverb
func (f *Fakery) Adverb() string {
	return f.RandomString(f.LoadGenericLocale(&wordsLoader).Get("adverbs"))
}
