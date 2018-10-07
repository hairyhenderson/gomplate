package funcs

import (
	"sync"
	"time"

	"github.com/brianvoe/gofakeit"
)

/*
TODO: verify struct types work
*/

var (
	fakef     *FakeFuncs
	fakefInit sync.Once
)

// FakeNS - keNS - the Fake namespace
func FakeNS() *FakeFuncs {
	fakefInit.Do(func() { fakef = &FakeFuncs{} })

	return fakef
}

// AddFakeFuncs -
func AddFakeFuncs(f map[string]interface{}) {
	f["Fake"] = FakeNS
}

// FakeFuncs -
type FakeFuncs struct {
}

// NOT IMPLEMENTED
// gofakeit.Struct()

// Address -
func (f *FakeFuncs) Address() *gofakeit.AddressInfo {
	return gofakeit.Address()
}

// Street -
func (f *FakeFuncs) Street() (street string) {
	return gofakeit.Street()
}

// StreetNumber -
func (f *FakeFuncs) StreetNumber() string {
	return gofakeit.StreetNumber()
}

// StreetPrefix -
func (f *FakeFuncs) StreetPrefix() string {
	return gofakeit.StreetPrefix()
}

// StreetName -
func (f *FakeFuncs) StreetName() string {
	return gofakeit.StreetName()
}

// StreetSuffix -
func (f *FakeFuncs) StreetSuffix() string {
	return gofakeit.StreetSuffix()
}

// City -
func (f *FakeFuncs) City() (city string) {
	return gofakeit.City()
}

// State -
func (f *FakeFuncs) State() string {
	return gofakeit.State()
}

// StateAbr -
func (f *FakeFuncs) StateAbr() string {
	return gofakeit.StateAbr()
}

// Zip -
func (f *FakeFuncs) Zip() string {
	return gofakeit.Zip()
}

// Country -
func (f *FakeFuncs) Country() string {
	return gofakeit.Country()
}

// CountryAbr -
func (f *FakeFuncs) CountryAbr() string {
	return gofakeit.CountryAbr()
}

// Latitude -
func (f *FakeFuncs) Latitude() float64 {
	return gofakeit.Latitude()
}

// LatitudeInRange -
func (f *FakeFuncs) LatitudeInRange(min, max float64) (float64, error) {
	return gofakeit.LatitudeInRange(min, max)
}

// Longitude -
func (f *FakeFuncs) Longitude() float64 {
	return gofakeit.Longitude()
}

// LongitudeInRange -
func (f *FakeFuncs) LongitudeInRange(min, max float64) (float64, error) {
	return gofakeit.LongitudeInRange(min, max)
}

// BeerName -
func (f *FakeFuncs) BeerName() string {
	return gofakeit.BeerName()
}

// BeerStyle -
func (f *FakeFuncs) BeerStyle() string {
	return gofakeit.BeerStyle()
}

// BeerHop -
func (f *FakeFuncs) BeerHop() string {
	return gofakeit.BeerHop()
}

// BeerYeast -
func (f *FakeFuncs) BeerYeast() string {
	return gofakeit.BeerYeast()
}

// BeerMalt -
func (f *FakeFuncs) BeerMalt() string {
	return gofakeit.BeerMalt()
}

// BeerIbu -
func (f *FakeFuncs) BeerIbu() string {
	return gofakeit.BeerIbu()
}

// BeerAlcohol -
func (f *FakeFuncs) BeerAlcohol() string {
	return gofakeit.BeerAlcohol()
}

// BeerBlg -
func (f *FakeFuncs) BeerBlg() string {
	return gofakeit.BeerBlg()
}

// Bool -
func (f *FakeFuncs) Bool() bool {
	return gofakeit.Bool()
}

// Color -
func (f *FakeFuncs) Color() string {
	return gofakeit.Color()
}

// SafeColor -
func (f *FakeFuncs) SafeColor() string {
	return gofakeit.SafeColor()
}

// HexColor -
func (f *FakeFuncs) HexColor() string {
	return gofakeit.HexColor()
}

// RGBColor -
func (f *FakeFuncs) RGBColor() []int {
	return gofakeit.RGBColor()
}

// Company -
func (f *FakeFuncs) Company() (company string) {
	return gofakeit.Company()
}

// CompanySuffix -
func (f *FakeFuncs) CompanySuffix() string {
	return gofakeit.CompanySuffix()
}

// BuzzWord -
func (f *FakeFuncs) BuzzWord() string {
	return gofakeit.BuzzWord()
}

// BS -
func (f *FakeFuncs) BS() string {
	return gofakeit.BS()
}

// Contact -
func (f *FakeFuncs) Contact() *gofakeit.ContactInfo {
	return gofakeit.Contact()
}

// Phone -
func (f *FakeFuncs) Phone() string {
	return gofakeit.Phone()
}

// PhoneFormatted -
func (f *FakeFuncs) PhoneFormatted() string {
	return gofakeit.PhoneFormatted()
}

// Email -
func (f *FakeFuncs) Email() string {
	return gofakeit.Email()
}

// Currency -
func (f *FakeFuncs) Currency() *gofakeit.CurrencyInfo {
	return gofakeit.Currency()
}

// CurrencyShort -
func (f *FakeFuncs) CurrencyShort() string {
	return gofakeit.CurrencyShort()
}

// CurrencyLong -
func (f *FakeFuncs) CurrencyLong() string {
	return gofakeit.CurrencyLong()
}

// Price -
func (f *FakeFuncs) Price(min, max float64) float64 {
	return gofakeit.Price(min, max)
}

// Date -
func (f *FakeFuncs) Date() time.Time {
	return gofakeit.Date()
}

// DateRange -
func (f *FakeFuncs) DateRange(start, end time.Time) time.Time {
	return gofakeit.DateRange(start, end)
}

// Month -
func (f *FakeFuncs) Month() string {
	return gofakeit.Month()
}

// Day -
func (f *FakeFuncs) Day() int {
	return gofakeit.Day()
}

// WeekDay -
func (f *FakeFuncs) WeekDay() string {
	return gofakeit.WeekDay()
}

// Year -
func (f *FakeFuncs) Year() int {
	return gofakeit.Year()
}

// Hour -
func (f *FakeFuncs) Hour() int {
	return gofakeit.Hour()
}

// Minute -
func (f *FakeFuncs) Minute() int {
	return gofakeit.Minute()
}

// Second -
func (f *FakeFuncs) Second() int {
	return gofakeit.Second()
}

// NanoSecond -
func (f *FakeFuncs) NanoSecond() int {
	return gofakeit.NanoSecond()
}

// TimeZone -
func (f *FakeFuncs) TimeZone() string {
	return gofakeit.TimeZone()
}

// TimeZoneFull -
func (f *FakeFuncs) TimeZoneFull() string {
	return gofakeit.TimeZoneFull()
}

// TimeZoneAbv -
func (f *FakeFuncs) TimeZoneAbv() string {
	return gofakeit.TimeZoneAbv()
}

// TimeZoneOffset -
func (f *FakeFuncs) TimeZoneOffset() float32 {
	return gofakeit.TimeZoneOffset()
}

// Seed -
func (f *FakeFuncs) Seed(seed int64) {
	gofakeit.Seed(seed)
	return
}

// MimeType -
func (f *FakeFuncs) MimeType() string {
	return gofakeit.MimeType()
}

// Extension -
func (f *FakeFuncs) Extension() string {
	return gofakeit.Extension()
}

// Generate -
func (f *FakeFuncs) Generate(dataVal string) string {
	return gofakeit.Generate(dataVal)
}

// HackerPhrase -
func (f *FakeFuncs) HackerPhrase() string {
	return gofakeit.HackerPhrase()
}

// HackerAbbreviation -
func (f *FakeFuncs) HackerAbbreviation() string {
	return gofakeit.HackerAbbreviation()
}

// HackerAdjective -
func (f *FakeFuncs) HackerAdjective() string {
	return gofakeit.HackerAdjective()
}

// HackerNoun -
func (f *FakeFuncs) HackerNoun() string {
	return gofakeit.HackerNoun()
}

// HackerVerb -
func (f *FakeFuncs) HackerVerb() string {
	return gofakeit.HackerVerb()
}

// HackerIngverb -
func (f *FakeFuncs) HackerIngverb() string {
	return gofakeit.HackerIngverb()
}

// HipsterWord -
func (f *FakeFuncs) HipsterWord() string {
	return gofakeit.HipsterWord()
}

// HipsterSentence -
func (f *FakeFuncs) HipsterSentence(wordCount int) string {
	return gofakeit.HipsterSentence(wordCount)
}

// HipsterParagraph -
func (f *FakeFuncs) HipsterParagraph(paragraphCount int, sentenceCount int, wordCount int, separator string) string {
	return gofakeit.HipsterParagraph(paragraphCount, sentenceCount, wordCount, separator)
}

// ImageURL -
func (f *FakeFuncs) ImageURL(width int, height int) string {
	return gofakeit.ImageURL(width, height)
}

// DomainName -
func (f *FakeFuncs) DomainName() string {
	return gofakeit.DomainName()
}

// DomainSuffix -
func (f *FakeFuncs) DomainSuffix() string {
	return gofakeit.DomainSuffix()
}

// URL -
func (f *FakeFuncs) URL() string {
	return gofakeit.URL()
}

// HTTPMethod -
func (f *FakeFuncs) HTTPMethod() string {
	return gofakeit.HTTPMethod()
}

// IPv4Address -
func (f *FakeFuncs) IPv4Address() string {
	return gofakeit.IPv4Address()
}

// IPv6Address -
func (f *FakeFuncs) IPv6Address() string {
	return gofakeit.IPv6Address()
}

// Username -
func (f *FakeFuncs) Username() string {
	return gofakeit.Username()
}

// Job -
func (f *FakeFuncs) Job() *gofakeit.JobInfo {
	return gofakeit.Job()
}

// JobTitle -
func (f *FakeFuncs) JobTitle() string {
	return gofakeit.JobTitle()
}

// JobDescriptor -
func (f *FakeFuncs) JobDescriptor() string {
	return gofakeit.JobDescriptor()
}

// JobLevel -
func (f *FakeFuncs) JobLevel() string {
	return gofakeit.JobLevel()
}

// LogLevel -
func (f *FakeFuncs) LogLevel(logType string) string {
	return gofakeit.LogLevel(logType)
}

// Categories -
func (f *FakeFuncs) Categories() map[string][]string {
	return gofakeit.Categories()
}

// Name -
func (f *FakeFuncs) Name() string {
	return gofakeit.Name()
}

// FirstName -
func (f *FakeFuncs) FirstName() string {
	return gofakeit.FirstName()
}

// LastName -
func (f *FakeFuncs) LastName() string {
	return gofakeit.LastName()
}

// NamePrefix -
func (f *FakeFuncs) NamePrefix() string {
	return gofakeit.NamePrefix()
}

// NameSuffix -
func (f *FakeFuncs) NameSuffix() string {
	return gofakeit.NameSuffix()
}

// Number -
func (f *FakeFuncs) Number(min int, max int) int {
	return gofakeit.Number(min, max)
}

// Uint8 -
func (f *FakeFuncs) Uint8() uint8 {
	return gofakeit.Uint8()
}

// Uint16 -
func (f *FakeFuncs) Uint16() uint16 {
	return gofakeit.Uint16()
}

// Uint32 -
func (f *FakeFuncs) Uint32() uint32 {
	return gofakeit.Uint32()
}

// Uint64 -
func (f *FakeFuncs) Uint64() uint64 {
	return gofakeit.Uint64()
}

// Int8 -
func (f *FakeFuncs) Int8() int8 {
	return gofakeit.Int8()
}

// Int16 -
func (f *FakeFuncs) Int16() int16 {
	return gofakeit.Int16()
}

// Int32 -
func (f *FakeFuncs) Int32() int32 {
	return gofakeit.Int32()
}

// Int64 -
func (f *FakeFuncs) Int64() int64 {
	return gofakeit.Int64()
}

// Float32 -
func (f *FakeFuncs) Float32() float32 {
	return gofakeit.Float32()
}

// Float64 -
func (f *FakeFuncs) Float64() float64 {
	return gofakeit.Float64()
}

// Numerify -
func (f *FakeFuncs) Numerify(str string) string {
	return gofakeit.Numerify(str)
}

// ShuffleInts -
func (f *FakeFuncs) ShuffleInts(a []int) []int {
	gofakeit.ShuffleInts(a)
	return a
}

// Password -
func (f *FakeFuncs) Password(lower bool, upper bool, numeric bool, special bool, space bool, num int) string {
	return gofakeit.Password(lower, upper, numeric, special, space, num)
}

// CreditCard -
func (f *FakeFuncs) CreditCard() *gofakeit.CreditCardInfo {
	return gofakeit.CreditCard()
}

// CreditCardType -
func (f *FakeFuncs) CreditCardType() string {
	return gofakeit.CreditCardType()
}

// CreditCardNumber -
func (f *FakeFuncs) CreditCardNumber() int {
	return gofakeit.CreditCardNumber()
}

// CreditCardNumberLuhn -
func (f *FakeFuncs) CreditCardNumberLuhn() int {
	return gofakeit.CreditCardNumberLuhn()
}

// CreditCardExp -
func (f *FakeFuncs) CreditCardExp() string {
	return gofakeit.CreditCardExp()
}

// CreditCardCvv -
func (f *FakeFuncs) CreditCardCvv() string {
	return gofakeit.CreditCardCvv()
}

// SSN -
func (f *FakeFuncs) SSN() string {
	return gofakeit.SSN()
}

// Gender -
func (f *FakeFuncs) Gender() string {
	return gofakeit.Gender()
}

// Person -
func (f *FakeFuncs) Person() *gofakeit.PersonInfo {
	return gofakeit.Person()
}

// SimpleStatusCode -
func (f *FakeFuncs) SimpleStatusCode() int {
	return gofakeit.SimpleStatusCode()
}

// StatusCode -
func (f *FakeFuncs) StatusCode() int {
	return gofakeit.StatusCode()
}

// Letter -
func (f *FakeFuncs) Letter() string {
	return gofakeit.Letter()
}

// Lexify -
func (f *FakeFuncs) Lexify(str string) string {
	return gofakeit.Lexify(str)
}

// ShuffleStrings -
func (f *FakeFuncs) ShuffleStrings(a []string) []string {
	gofakeit.ShuffleStrings(a)
	return a
}

// RandString -
func (f *FakeFuncs) RandString(a []string) string {
	return gofakeit.RandString(a)
}

// UUID -
func (f *FakeFuncs) UUID() string {
	return gofakeit.UUID()
}

// UserAgent -
func (f *FakeFuncs) UserAgent() string {
	return gofakeit.UserAgent()
}

// ChromeUserAgent -
func (f *FakeFuncs) ChromeUserAgent() string {
	return gofakeit.ChromeUserAgent()
}

// FirefoxUserAgent -
func (f *FakeFuncs) FirefoxUserAgent() string {
	return gofakeit.FirefoxUserAgent()
}

// SafariUserAgent -
func (f *FakeFuncs) SafariUserAgent() string {
	return gofakeit.SafariUserAgent()
}

// OperaUserAgent -
func (f *FakeFuncs) OperaUserAgent() string {
	return gofakeit.OperaUserAgent()
}

// Word -
func (f *FakeFuncs) Word() string {
	return gofakeit.Word()
}

// Sentence -
func (f *FakeFuncs) Sentence(wordCount int) string {
	return gofakeit.Sentence(wordCount)
}

// Paragraph -
func (f *FakeFuncs) Paragraph(paragraphCount int, sentenceCount int, wordCount int, separator string) string {
	return gofakeit.Paragraph(paragraphCount, sentenceCount, wordCount, separator)
}
