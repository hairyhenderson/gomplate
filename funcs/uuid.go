package funcs

import (
	"context"
	"sync"

	"github.com/hairyhenderson/gomplate/v3/conv"

	"github.com/google/uuid"
)

var (
	uuidNS     *UUIDFuncs
	uuidNSInit sync.Once
)

// UUIDNS -
func UUIDNS() *UUIDFuncs {
	uuidNSInit.Do(func() {
		uuidNS = &UUIDFuncs{}
	})
	return uuidNS
}

// AddUUIDFuncs -
func AddUUIDFuncs(f map[string]interface{}) {
	for k, v := range CreateUUIDFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateUUIDFuncs -
func CreateUUIDFuncs(ctx context.Context) map[string]interface{} {
	ns := UUIDNS()
	ns.ctx = ctx
	return map[string]interface{}{"uuid": UUIDNS}
}

// UUIDFuncs -
type UUIDFuncs struct {
	ctx context.Context
}

// V1 - return a version 1 UUID (based on the current MAC Address and the
// current date/time). Use V4 instead in most cases.
func (f *UUIDFuncs) V1() (string, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// V4 - return a version 4 (random) UUID
func (f *UUIDFuncs) V4() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Nil -
func (f *UUIDFuncs) Nil() (string, error) {
	return uuid.Nil.String(), nil
}

// IsValid - checks if the given UUID is in the correct format. It does not
// validate whether the version or variant are correct.
func (f *UUIDFuncs) IsValid(in interface{}) (bool, error) {
	_, err := f.Parse(in)
	return err == nil, nil
}

// Parse - parse a UUID for further manipulation or inspection.
//
// Both the standard UUID forms of xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx and
// urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx are decoded as well as the
// Microsoft encoding {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx} and the raw hex
// encoding: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
func (f *UUIDFuncs) Parse(in interface{}) (uuid.UUID, error) {
	u, err := uuid.Parse(conv.ToString(in))
	if err != nil {
		return uuid.Nil, err
	}
	return u, err
}
