// Package tnsnames parses Oracle tnsnames.ora files into a small tree model.
//
// The package is designed for callers that need to load aliases, re-render their
// descriptor strings, or extract common ADDRESS and CONNECT_DATA fields to build
// simple client connection strings.
package tnsnames
