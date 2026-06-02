// Package platform embeds the ADL model — platform.agent and every metamodel and
// blueprint — so services (e.g. cmd/graph) can build the graph from the model
// without depending on files on disk at runtime. The embedded model travels with
// the static binary, so it runs unchanged on scratch images and any cluster.
package platform

import "embed"

// Model is the embedded ADL model: platform.agent plus metamodels and blueprints.
//
//go:embed platform.agent metamodels/*.agent blueprints/*.agent
var Model embed.FS
