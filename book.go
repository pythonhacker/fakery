// Generate fake books
package gofakelib

import (
	"github.com/gertd/go-pluralize"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	bookLoader DataLoader
	booksData  BooksData
)

func init() {
	// We prepare book data from different words
	bookLoader.Init("book.json")
	// Convert to structure
	ConvertMapToStruct(bookLoader.Preload(GenericLocale), &booksData)
}

type Book struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	Publisher string `json:"publisher"`
	Year      int    `json:"year"`
	ISBN10    string `json:"isbn_10,omitempty"`
	ISBN13    string `json:"isbn_13,omitempty"`
	Language  string `json:"language"`
	Genre     string `json:"genre"`
	PageCount int    `json:"page_count"`
	Format    string `json:"format"`
	Base
}

type BooksData struct {
	LiteraryNouns        []string `json:"literary_nouns"`
	LiteraryAdjectives   []string `json:"literary_adjectives"`
	LiteraryGerunds      []string `json:"literary_gerunds"`
	LiteraryActionVerbs  []string `json:"literary_action_verbs"`
	LiteraryConjunctions []string `json:"literary_conjunctions"`
	LiteraryPrepositions []string `json:"literary_prepositions"`
	PublisherFirstNames  []string `json:"publisher_first_names"`
	PublisherSecondNames []string `json:"publisher_second_names"`
	PublisherLastNames   []string `json:"publisher_last_names"`

	Genres []string `json:"genres"`
}

var numbers = []string{"One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight",
	"Nine", "Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Nineteen", "Twenty",
	"Thirty", "Fifty", "Hundred", "Thousand", "Ten Thousand", "Fifty Thousand",
	"An Infinity of", "Countless",
}

var formats = WeightedArray{
	Items: []WeightedItem{
		{Item: "Paperback", Weight: 0.60},
		{Item: "Mass Market Paperback", Weight: 0.15},
		{Item: "Hardback", Weight: 0.10},
		{Item: "EPUB", Weight: 0.05},
		{Item: "Kindle Edition", Weight: 0.10},
	},
}

func (b Book) String() string {
	return b.Base.String(b)
}

// validateISBN10 checks if the provided string is a valid ISBN-10 number
func (f *Faker) ValidateISBN(isbn string) bool {
	// Remove any hyphens or spaces
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	// ISBN-10 must be exactly 10 characters
	if len(isbn) == 10 {
		return f.validateISBN10(isbn)
	} else if len(isbn) == 13 {
		return f.validateISBN13(isbn)
	}

	return false
}

func (f *Faker) validateISBN13(isbn string) bool {
	match, err := regexp.MatchString(`^\d{13}$`, isbn)
	if err != nil || !match {
		return false
	}

	// Calculate the check digit
	var sum int
	for i, char := range isbn[:12] {
		digit, _ := strconv.Atoi(string(char))
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	// Calculate check digit: (10 - (sum mod 10)) mod 10
	calculatedCheckDigit := (10 - (sum % 10)) % 10
	lastDigit, _ := strconv.Atoi(string(isbn[12]))

	// Verify the check digit
	return calculatedCheckDigit == lastDigit
}

// validate ISBN10 numbers
func (f *Faker) validateISBN10(isbn string) bool {

	// Check if first 9 characters are digits
	for i := 0; i < 9; i++ {
		if !unicode.IsDigit(rune(isbn[i])) {
			return false
		}
	}

	// Last character can be a digit or 'X'
	lastChar := isbn[9]
	if !unicode.IsDigit(rune(lastChar)) && lastChar != 'X' && lastChar != 'x' {
		return false
	}

	// Calculate the weighted sum
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(isbn[i]))
		sum += digit * (10 - i)
	}

	// Add the check digit
	var checkValue int
	if lastChar == 'X' || lastChar == 'x' {
		checkValue = 10
	} else {
		checkValue, _ = strconv.Atoi(string(lastChar))
	}
	sum += checkValue

	// Valid ISBN-10 must satisfy: (11 - (sum % 11)) % 11 = checkValue
	// which simplifies to: sum % 11 == 0
	return sum%11 == 0
}

// generateISBN10 creates a valid ISBN-10 number
func (f *Faker) generateISBN10() string {
	// Generate 9 random digits
	digits := make([]int, 9)
	for i := 0; i < 9; i++ {
		digits[i] = f.IntRange(10)
	}

	// Calculate check digit (10th digit)
	sum := 0
	for i := 0; i < 9; i++ {
		sum += digits[i] * (10 - i)
	}
	checkDigit := (11 - (sum % 11)) % 11

	// Convert to string
	isbn := ""
	for _, d := range digits {
		isbn += strconv.Itoa(d)
	}

	// Add check digit (X is used for 10)
	if checkDigit == 10 {
		isbn += "X"
	} else {
		isbn += strconv.Itoa(checkDigit)
	}

	// Format with hyphens: group-publisher-title-check
	// Common format: x-xxx-xxxxx-x
	parts := []string{
		isbn[0:1],
		isbn[1:4],
		isbn[4:9],
		isbn[9:10],
	}

	return strings.Join(parts, "-")
}

// generateISBN13 creates a valid ISBN-13 number
func (f *Faker) generateISBN13() string {
	// Start with 978 or 979 (EAN prefix for books)
	prefix := 978
	if f.RandDigitNonZero() > 7 { // About 20% chance for 979
		prefix = 979
	}

	// Generate 9 random digits
	digits := make([]int, 9)
	for i := 0; i < 9; i++ {
		digits[i] = f.IntRange(10)
	}

	// Combine prefix and random digits
	allDigits := make([]int, 12)
	allDigits[0] = prefix / 100
	allDigits[1] = (prefix / 10) % 10
	allDigits[2] = prefix % 10
	copy(allDigits[3:], digits)

	// Calculate check digit (13th digit) using ISBN-13 algorithm
	sum := 0
	for i := 0; i < 12; i++ {
		weight := 1
		if i%2 == 1 {
			weight = 3
		}
		sum += allDigits[i] * weight
	}
	checkDigit := (10 - (sum % 10)) % 10

	// Convert to string
	isbn := ""
	for _, d := range allDigits {
		isbn += strconv.Itoa(d)
	}
	isbn += strconv.Itoa(checkDigit)

	// Format with hyphens: EAN-group-publisher-title-check
	// Common format: xxx-x-xxx-xxxxx-x
	parts := []string{
		isbn[0:3],
		isbn[3:4],
		isbn[4:7],
		isbn[7:12],
		isbn[12:13],
	}

	return strings.Join(parts, "-")
}

func (f *Faker) BookTitle() string {

	var items []string
	// Patterns for which a "The" can be prefixed
	var theAllowedPatterns = "2479"
	// Generate a random pattern
	pattern := f.IntRange(11) + 1
	//	fmt.Println(pattern)

	// Easy placeholders to avoid long code
	nouns := booksData.LiteraryNouns
	adjectives := booksData.LiteraryAdjectives
	gerunds := booksData.LiteraryGerunds
	conjunctions := booksData.LiteraryConjunctions
	prepositions := booksData.LiteraryPrepositions
	actionVerbs := booksData.LiteraryActionVerbs

	switch pattern {
	case 1: // Single Noun
		// return f.RandomString(nouns)
		// The + Noun
		items = []string{"The", f.RandomString(nouns)}

	case 2: // Adjective + Noun
		items = []string{f.RandomString(adjectives), f.RandomString(nouns)}

	case 3: // Possessive + Noun
		items = []string{f.FirstName() + "'s", f.RandomString(nouns)}

	case 4: // Noun + Preposition + Noun
		items = []string{f.RandomString(nouns), f.RandomString(prepositions), "the", f.RandomString(nouns)}

	case 5: // Gerund Phrase
		items = []string{f.RandomString(gerunds), f.RandomString(prepositions), f.RandomString(nouns)}

	case 6: // Article + Adjective + Noun
		adjective := f.RandomString(adjectives)
		items = []string{strings.Title(DetermineArticle(adjective, f)), adjective, f.RandomString(nouns)}

	case 7: // Number + Noun
		number := f.RandomString(numbers)
		if number == "One" {
			items = []string{number, f.RandomString(nouns)}
		} else {
			items = []string{number, pluralize.NewClient().Plural(f.RandomString(nouns))}
		}

	case 8: // Prepositional Phrase
		prep := f.RandomString(prepositions)
		items = []string{strings.Title(prep), "the", f.RandomString(nouns)}

	case 9: // Noun of Noun
		lastNoun := f.RandomString(nouns)
		if f.RollDice() < 6 {
			// produce this format 5/6 times
			// E.g: Fire of the Sanctuary
			items = []string{f.RandomString(nouns), "of", DetermineArticle(lastNoun, f), lastNoun}
		} else {
			// produce this rarely
			// E.g: Sentinel of Wall
			items = []string{f.RandomString(nouns), "of", lastNoun}
		}

	case 10: // Alliterative (Noun and Noun)
		n1 := f.RandomString(nouns)
		// Find nouns that start with the same letter
		sameLetterNouns := []string{}
		firstChar := []rune(n1)[0]

		for _, n := range nouns {
			if []rune(n)[0] == firstChar && n != n1 {
				sameLetterNouns = append(sameLetterNouns, n)
			}
		}

		n2 := n1
		if len(sameLetterNouns) > 0 {
			n2 = sameLetterNouns[f.IntRange(len(sameLetterNouns))]
		} else {
			n2 = f.RandomString(nouns)
		}

		items = []string{n1, f.RandomString(conjunctions), n2}

	case 11: // How to Verb
		noun := f.RandomString(nouns)
		items = []string{"How to", f.RandomString(actionVerbs), DetermineArticle(noun, f), noun}
	}

	title := strings.Join(items, " ")
	// Once in a while add a "The" prefix
	if f.RollDice() == 6 && strings.Contains(theAllowedPatterns, strconv.Itoa(pattern)) {
		title = "The " + title
	}
	return title
}

func (f *Faker) BookPublisher() string {

	var items []string
	pattern := f.IntRange(3) + 1

	firstNames := booksData.PublisherFirstNames
	secondNames := booksData.PublisherSecondNames
	lastNames := booksData.PublisherLastNames

	switch pattern {
	case 1: // FirstWord + SecondWord
		items = []string{f.RandomString(firstNames), f.RandomString(secondNames), f.RandomString(lastNames)}
	case 2:
		items = []string{f.RandomString(firstNames), f.RandomString(lastNames)}
	case 3:
		items = []string{f.RandomString(secondNames), f.RandomString(lastNames)}
	}

	return strings.Join(items, " ")
}

func (f *Faker) BookGenre() string {
	return f.RandomString(booksData.Genres)
}

func (f *Faker) BookAuthor() string {
	return f.Name()
}

func (f *Faker) BookFormat() string {
	format, _ := f.RandomItem(&formats)
	return format
}

func (f *Faker) BookYear() int {
	return f.RandIntBetween(1970, time.Now().Year()-1)
}

func (f *Faker) BookISBN() string {
	if f.Choice() == 1 {
		return f.generateISBN10()
	} else {
		return f.generateISBN13()
	}
}

func (f *Faker) Book() *Book {
	var b Book

	b.Title = f.BookTitle()
	b.Author = f.BookAuthor()
	b.Genre = f.BookGenre()
	b.Year = f.BookYear()
	b.Publisher = f.BookPublisher()
	b.Language = strings.Split(f.locale, "_")[0]
	if f.Choice() == 1 {
		b.ISBN10 = f.generateISBN10()
	} else {
		b.ISBN13 = f.generateISBN13()
	}
	// Page count
	b.PageCount = f.RandIntBetween(100, 501)

	b.Format, _ = f.RandomItem(&formats)

	return &b
}
