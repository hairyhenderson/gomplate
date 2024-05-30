package funcs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/parsers"
	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"
)

// CreateDataSourceFuncs -
func CreateDataSourceFuncs(ctx context.Context, sr datafs.DataSourceReader) map[string]interface{} {
	ns := &dataSourceFuncs{
		ctx: ctx,
		sr:  sr,
	}

	f := map[string]interface{}{}

	// undocumented but available
	f["_datasource"] = func() interface{} { return ns }

	f["datasource"] = ns.Datasource
	f["ds"] = ns.Datasource
	f["datasourceExists"] = ns.DatasourceExists
	f["datasourceReachable"] = ns.DatasourceReachable
	f["defineDatasource"] = ns.DefineDatasource
	f["include"] = ns.Include
	f["listDatasources"] = ns.ListDatasources

	return f
}

// dataSourceFuncs - datasource reading functions
type dataSourceFuncs struct {
	ctx context.Context
	sr  datafs.DataSourceReader
}

// Include - Reads from the named datasource, without parsing the data, which
// is returned as a string.
func (d *dataSourceFuncs) Include(alias string, args ...string) (string, error) {
	_, b, err := d.sr.ReadSource(d.ctx, alias, args...)
	if err != nil {
		return "", err
	}

	return string(b), err
}

// Datasource - Reads from the named datasource, and returns the parsed datafs.
func (d *dataSourceFuncs) Datasource(alias string, args ...string) (interface{}, error) {
	ct, b, err := d.sr.ReadSource(d.ctx, alias, args...)
	if err != nil {
		return nil, err
	}

	return parsers.ParseData(ct, string(b))
}

// DefineDatasource -
func (d *dataSourceFuncs) DefineDatasource(alias, value string) (string, error) {
	if alias == "" {
		return "", fmt.Errorf("datasource alias must be provided")
	}
	if d.DatasourceExists(alias) {
		slog.DebugContext(d.ctx, "defineDatasource: ignoring attempt to redefine datasource", "alias", alias)
		return "", nil
	}
	srcURL, err := urlhelpers.ParseSourceURL(value)
	if err != nil {
		return "", fmt.Errorf("parse datasource URL: %w", err)
	}

	d.sr.Register(alias, config.DataSource{URL: srcURL})
	return "", nil
}

// DatasourceExists -
func (d *dataSourceFuncs) DatasourceExists(alias string) bool {
	_, ok := d.sr.Lookup(alias)
	return ok
}

// DatasourceReachable - Determines if the named datasource is reachable with
// the given arguments. Reads from the datasource, and discards the returned datafs.
func (d *dataSourceFuncs) DatasourceReachable(alias string, args ...string) bool {
	// first, if the datasource doesn't exist, we can't reach it
	if !d.DatasourceExists(alias) {
		return false
	}

	_, _, err := d.sr.ReadSource(d.ctx, alias, args...)
	return err == nil
}

// Show all datasources  -
func (d *dataSourceFuncs) ListDatasources() []string {
	return d.sr.List()
}
