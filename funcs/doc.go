/*
Package funcs is an internal package that provides gomplate namespaces and
functions to be used in 'text/template' templates.

The different namespaces can be added individually:

	f := template.FuncMap{}
	for k, v := range funcs.CreateMathFuncs(ctx) {
		f[k] = v
	}
	for k, v := range funcs.CreateNetFuncs(ctx) {
		f[k] = v
	}

Even though the functions are exported, these are not intended to be called
programmatically by external consumers, but instead only to be used as template
functions.

Deprecated: This package will be made internal in a future major version.
*/
package funcs
