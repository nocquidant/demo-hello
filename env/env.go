package env

// VERSION indicates which version of the binary is running.
var VERSION string

// GITCOMMIT indicates which git hash the binary was built off of
var GITCOMMIT string

// The name of the application
var NAME string

// The HTTP port of the application
var PORT int

// The remote service to call through the '/request' endpoint
var REMOTE_URL string
