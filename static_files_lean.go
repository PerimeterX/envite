//go:build lean

package fengshui

import (
	"errors"
	"net/http"
)

var errUnimplemented = errors.New("this is a lean build version of feng shui, it does not include any UI")

type assetFile struct{}

func (a assetFile) Open(name string) (http.File, error) {
	return nil, errUnimplemented
}

func AssetFile() http.FileSystem {
	return assetFile{}
}

func Asset(name string) ([]byte, error) {
	return nil, errUnimplemented
}
