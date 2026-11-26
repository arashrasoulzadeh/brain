package pluginsdk

import _ "embed"

// PureHelperSource is loaded inside the user-function VM, not executed in the host realm.
//
//go:embed src/helpers.mjs
var PureHelperSource string
