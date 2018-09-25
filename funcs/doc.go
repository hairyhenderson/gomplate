/*
Package funcs provides gomplate namespaces and functions to be used in 'text/template' templates.

The different namespaces can be added individually:

	f := template.FuncMap{}
	funcs.AddMathFuncs(f)
	funcs.AddNetFuncs(f)

Even though the functions are exported, these are not intended to be called programmatically
by external consumers, but instead only to be used as template functions.

*/
package funcs
