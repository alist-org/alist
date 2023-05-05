package errs

import "errors"

var NotSupportPluginMode = errors.New("not support plugin mode")
var PluginHasBeenLoaded = errors.New("plugin has been loaded")

var NotFoundPluginByRepository = errors.New("not found plugin by repository")
var NotFoundPluginVersionByRepository = errors.New("not found plugin version by repository")
var PluginHasBeenInstalled = errors.New("plugin has been installed")
