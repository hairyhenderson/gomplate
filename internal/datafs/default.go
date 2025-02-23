package datafs

import (
	"sync"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/autofs"
)

// DefaultProvider is the default filesystem provider used by gomplate
var DefaultProvider = sync.OnceValue(
	func() fsimpl.FSProvider {
		fsp := fsimpl.NewMux()

		// start with all go-fsimpl filesystems
		fsp.Add(autofs.FS)

		// override go-fsimpl's filefs with wdfs to handle working directories
		fsp.Add(wdFSProvider)

		// gomplate-only filesystems
		fsp.Add(EnvFS)
		fsp.Add(StdinFS)
		fsp.Add(mergeFSProvider)

		return fsp
	})()
