// apparmor.d - Full set of apparmor profiles
// Copyright (C) 2021-2023 Alexandre Pujol <alexandre@pujol.io>
// SPDX-License-Identifier: GPL-2.0-only

package aa

import (
	"github.com/arduino/go-paths-helper"
)

// Default Apparmor magic directory: /etc/apparmor.d/.
var MagicRoot = paths.New("/etc/apparmor.d")

// AppArmorProfileFiles represents a full set of apparmor profiles
type AppArmorProfileFiles map[string]*AppArmorProfileFile

// AppArmorProfileFile represents a full apparmor profile file.
// Warning: close to the BNF grammar of apparmor profile but not exactly the same (yet):
//   - Some rules are not supported yet (subprofile, hat...)
//   - The structure is simplified as it only aims at writing profile, not parsing it.
type AppArmorProfileFile struct {
	Preamble Rules
	Profiles []*Profile
}

func NewAppArmorProfile() *AppArmorProfileFile {
	return &AppArmorProfileFile{}
}

// DefaultTunables return a minimal working profile to build the profile
// It should not be used when loading file from /etc/apparmor.d
func DefaultTunables() *AppArmorProfileFile {
	return &AppArmorProfileFile{
		Preamble: Rules{
			&Variable{Name: "bin", Values: []string{"/{,usr/}{,s}bin"}, Define: true},
			&Variable{Name: "lib", Values: []string{"/{,usr/}lib{,exec,32,64}"}, Define: true},
			&Variable{Name: "multiarch", Values: []string{"*-linux-gnu*"}, Define: true},
			&Variable{Name: "HOME", Values: []string{"/home/*"}, Define: true},
			&Variable{Name: "user_share_dirs", Values: []string{"/home/*/.local/share"}, Define: true},
			&Variable{Name: "etc_ro", Values: []string{"/{,usr/}etc/"}, Define: true},
			&Variable{Name: "int", Values: []string{"[0-9]{[0-9],}{[0-9],}{[0-9],}{[0-9],}{[0-9],}{[0-9],}{[0-9],}{[0-9],}{[0-9],}"}, Define: true},
			&Variable{Name: "user_cache_dirs", Values: []string{"/home/*/.cache"}, Define: true},
		},
	}
}

// String returns the formatted representation of a profile file as a string
func (f *AppArmorProfileFile) String() string {
	return renderTemplate("apparmor", f)
}

// Validate the profile file
func (f *AppArmorProfileFile) Validate() error {
	if err := f.Preamble.Validate(); err != nil {
		return err
	}
	for _, p := range f.Profiles {
		if err := p.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GetDefaultProfile ensure a profile is always present in the profile file and
// return it, as a default profile.
func (f *AppArmorProfileFile) GetDefaultProfile() *Profile {
	if len(f.Profiles) == 0 {
		f.Profiles = append(f.Profiles, &Profile{})
	}
	return f.Profiles[0]
}

// Sort the rules in the profile
// Follow: https://apparmor.pujol.io/development/guidelines/#guidelines
func (f *AppArmorProfileFile) Sort() {
	for _, p := range f.Profiles {
		p.Sort()
	}
}

// MergeRules merge similar rules together.
// Steps:
//   - Remove identical rules
//   - Merge rule access. Eg: for same path, 'r' and 'w' becomes 'rw'
//
// Note: logs.regCleanLogs helps a lot to do a first cleaning
func (f *AppArmorProfileFile) MergeRules() {
	for _, p := range f.Profiles {
		p.Merge()
	}
}

// Format the profile for better readability before printing it.
// Follow: https://apparmor.pujol.io/development/guidelines/#the-file-block
func (f *AppArmorProfileFile) Format() {
	for _, p := range f.Profiles {
		p.Format()
	}
}
