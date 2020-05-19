package main

import (
	"github.com/Masterminds/semver"
	"github.com/parnurzeal/gorequest"
)

type Updater interface {
	CurrentVersion() (*semver.Version, error)
	ApplyUpdate(v *semver.Version, stopParent bool) error
}

func CheckUpdate() (*semver.Version, error) {
	info := new(Info)
	_, _, _ = gorequest.New().Get("http://localhost:3000/info").EndStruct(&info)
	return semver.NewVersion(info.Version)
}
