package funcs

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/hairyhenderson/gomplate/v5/conv"
)

// CreateFakeFuncs -
func CreateFakeFuncs(ctx context.Context) map[string]any {
	ns := &FakeFuncs{ctx: ctx, faker: gofakeit.New(0)}
	return map[string]any{
		"fake": func() any { return ns },
	}
}

// FakeFuncs generates realistic fake data with gofakeit.
//
// Each namespace owns its Faker rather than sharing gofakeit.GlobalFaker, so
// nothing else in the process can change its source. The Faker is created in
// locked mode, which serialises every read from the source, so the namespace
// is safe for concurrent use.
type FakeFuncs struct {
	ctx   context.Context
	faker *gofakeit.Faker
}

// Seeded returns a copy of the namespace whose values are derived from seed,
// so the same seed always produces the same sequence of values.
func (f *FakeFuncs) Seeded(seed any) *FakeFuncs {
	return &FakeFuncs{ctx: f.ctx, faker: seededFaker(conv.ToString(seed))}
}

// seededFaker derives the source's state from a hash of seed, as
// net.GenerateMAC does. It avoids gofakeit.New, which treats a zero seed as a
// request for a random one, and Faker.Seed, which reseeds
// gofakeit.GlobalFaker rather than its receiver.
func seededFaker(seed string) *gofakeit.Faker {
	sum := sha256.Sum256([]byte(seed))
	src := rand.NewPCG(binary.LittleEndian.Uint64(sum[:8]), binary.LittleEndian.Uint64(sum[8:16]))
	return gofakeit.NewFaker(src, true)
}

// Person

func (f *FakeFuncs) Name() string       { return f.faker.Name() }
func (f *FakeFuncs) FirstName() string  { return f.faker.FirstName() }
func (f *FakeFuncs) MiddleName() string { return f.faker.MiddleName() }
func (f *FakeFuncs) LastName() string   { return f.faker.LastName() }
func (f *FakeFuncs) NamePrefix() string { return f.faker.NamePrefix() }
func (f *FakeFuncs) NameSuffix() string { return f.faker.NameSuffix() }
func (f *FakeFuncs) Hobby() string      { return f.faker.Hobby() }

// Contact

func (f *FakeFuncs) Email() string          { return f.faker.Email() }
func (f *FakeFuncs) Phone() string          { return f.faker.Phone() }
func (f *FakeFuncs) PhoneFormatted() string { return f.faker.PhoneFormatted() }

// Address

func (f *FakeFuncs) Street() string         { return f.faker.Street() }
func (f *FakeFuncs) City() string           { return f.faker.City() }
func (f *FakeFuncs) State() string          { return f.faker.State() }
func (f *FakeFuncs) StateAbr() string       { return f.faker.StateAbr() }
func (f *FakeFuncs) Zip() string            { return f.faker.Zip() }
func (f *FakeFuncs) Country() string        { return f.faker.Country() }
func (f *FakeFuncs) CountryAbr() string     { return f.faker.CountryAbr() }
func (f *FakeFuncs) TimeZoneRegion() string { return f.faker.TimeZoneRegion() }

// Company and job

func (f *FakeFuncs) CompanySuffix() string { return f.faker.CompanySuffix() }
func (f *FakeFuncs) JobTitle() string      { return f.faker.JobTitle() }
func (f *FakeFuncs) BuzzWord() string      { return f.faker.BuzzWord() }
func (f *FakeFuncs) BS() string            { return f.faker.BS() }
func (f *FakeFuncs) Slogan() string        { return f.faker.Slogan() }

// Internet

func (f *FakeFuncs) DomainName() string   { return f.faker.DomainName() }
func (f *FakeFuncs) DomainSuffix() string { return f.faker.DomainSuffix() }
func (f *FakeFuncs) URL() string          { return f.faker.URL() }
func (f *FakeFuncs) IPv4Address() string  { return f.faker.IPv4Address() }
func (f *FakeFuncs) IPv6Address() string  { return f.faker.IPv6Address() }
func (f *FakeFuncs) HTTPMethod() string   { return f.faker.HTTPMethod() }
func (f *FakeFuncs) UserAgent() string    { return f.faker.UserAgent() }

// Product and text

func (f *FakeFuncs) ProductName() string        { return f.faker.ProductName() }
func (f *FakeFuncs) ProductCategory() string    { return f.faker.ProductCategory() }
func (f *FakeFuncs) ProductDescription() string { return f.faker.ProductDescription() }
func (f *FakeFuncs) Word() string               { return f.faker.Word() }
func (f *FakeFuncs) Phrase() string             { return f.faker.Phrase() }
