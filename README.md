filerelay
=========

Translation of `filerelaytest.c` into Go.

This requires `libimobiledevice` to be installed and has only been tested with the `libimobiledevice` installed by
homebrew on OSX.

The cgo code makes some assumptions about go internal structure layout which may not be valid on non-`x86_64`
architectures.


