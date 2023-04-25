// apparmor.d - Full set of apparmor profiles
// Copyright (C) 2023 Alexandre Pujol <alexandre@pujol.io>
// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arduino/go-paths-helper"
	"github.com/roddhjav/apparmor.d/pkg/aa"
	"github.com/roddhjav/apparmor.d/pkg/logging"
)

// Initialize a new clean apparmor.d build directory
func Synchronise() error {
	dirs := paths.PathList{RootApparmord, Root.Join("root")}
	for _, dir := range dirs {
		if err := dir.RemoveAll(); err != nil {
			return err
		}
	}
	for _, path := range []string{"./apparmor.d", "./root"} {
		cmd := exec.Command("rsync", "-a", path, Root.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	logging.Success("Initialize a new clean apparmor.d build directory")
	return nil
}

// Ignore profiles and files as defined in dists/ignore/
func Ignore() error {
	for _, name := range []string{"main.ignore", Distribution + ".ignore"} {
		path := paths.New("dists/ignore/" + name)
		if !path.Exist() {
			continue
		}
		lines, _ := path.ReadFileAsLines()
		for _, line := range lines {
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			profile := Root.Join(line)
			if profile.NotExist() {
				files, err := RootApparmord.ReadDirRecursiveFiltered(nil, paths.FilterNames(line))
				if err != nil {
					return err
				}
				for _, path := range files {
					if err := path.RemoveAll(); err != nil {
						return err
					}
				}
			} else {
				if err := profile.RemoveAll(); err != nil {
					return err
				}
			}
		}
		logging.Success("Ignore profiles/files in %s", path)
	}
	return nil
}

// Merge all profiles in a new apparmor.d directory
func Merge() error {
	var dirToMerge = []string{
		"groups/*/*", "groups",
		"profiles-*-*/*", "profiles-*",
	}

	idx := 0
	for idx < len(dirToMerge)-1 {
		dirMoved, dirRemoved := dirToMerge[idx], dirToMerge[idx+1]
		files, err := filepath.Glob(RootApparmord.Join(dirMoved).String())
		if err != nil {
			return err
		}
		for _, file := range files {
			err := os.Rename(file, RootApparmord.Join(filepath.Base(file)).String())
			if err != nil {
				return err
			}
		}

		files, err = filepath.Glob(RootApparmord.Join(dirRemoved).String())
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := paths.New(file).RemoveAll(); err != nil {
				return err
			}
		}
		idx = idx + 2
	}
	logging.Success("Merge all profiles")
	return nil
}

// Set the distribution specificities
func Configure() error {
	switch Distribution {
	case "arch":
		if err := setLibexec("/{usr/,}lib"); err != nil {
			return err
		}

	case "opensuse":
		if err := setLibexec("/{usr/,}libexec"); err != nil {
			return err
		}

	case "debian", "ubuntu", "whonix":
		if err := setLibexec("/{usr/,}libexec"); err != nil {
			return err
		}

		// Copy Ubuntu specific profiles
		if err := copyTo(DistDir.Join("ubuntu"), RootApparmord); err != nil {
			return err
		}
		if Distribution == "ubuntu" {
			break
		}

		// Copy debian specific profiles
		if err := copyTo(DistDir.Join("debian"), RootApparmord); err != nil {
			return err
		}

		// Remove ABI on abstractions files
		files, _ := RootApparmord.Join("abstractions").ReadDir(paths.FilterOutDirectories())
		for _, file := range files {
			if !file.Exist() {
				continue
			}
			content, _ := file.ReadFile()
			profile := BuildABI(string(content))
			if err := file.WriteFile([]byte(profile)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%s is not a supported distribution", Distribution)

	}
	return nil
}

func setLibexec(libexec string) error {
	aa.Tunables["libexec"] = []string{libexec}
	file, err := RootApparmord.Join("tunables", "multiarch.d", "apparmor.d").Append()
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(`@{libexec}=` + libexec)
	return err
}

func copyTo(src *paths.Path, dst *paths.Path) error {
	files, err := src.ReadDirRecursiveFiltered(nil, paths.FilterOutDirectories())
	if err != nil {
		return err
	}
	for _, file := range files {
		destination, err := file.RelFrom(src)
		if err != nil {
			return err
		}
		destination = dst.JoinPath(destination)
		if err := file.CopyTo(destination); err != nil {
			return err
		}
	}
	return nil
}

// Set flags on some profiles according to manifest defined in `dists/flags/`
func SetFlags() error {
	for _, name := range []string{"main.flags", Distribution + ".flags"} {
		path := paths.New("dists/flags/" + name)
		if !path.Exist() {
			continue
		}
		lines, _ := path.ReadFileAsLines()
		for _, line := range lines {
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			manifest := strings.Split(line, " ")
			profile := manifest[0]
			file := RootApparmord.Join(profile)
			if !file.Exist() {
				logging.Warning("Profile %s not found", profile)
				continue
			}

			// If flags is set, overwrite profile flag
			if len(manifest) > 1 {
				flags := " flags=(" + manifest[1] + ") {"
				content, err := file.ReadFile()
				if err != nil {
					return err
				}

				// Remove all flags definition, then set manifest' flags
				res := regFlag.ReplaceAllLiteralString(string(content), "")
				res = regProfileHeader.ReplaceAllLiteralString(res, flags)
				if err := file.WriteFile([]byte(res)); err != nil {
					return err
				}
			}
		}
		logging.Success("Set profile flags from %s", path)
	}
	return nil
}

// Set AppArmor for full system policy
// See https://gitlab.com/apparmor/apparmor/-/wikis/FullSystemPolicy
// https://gitlab.com/apparmor/apparmor/-/wikis/AppArmorInSystemd#early-policy-loads
func SetFullSystemPolicy() error {
	if !Full {
		return nil
	}

	for _, name := range []string{"init", "systemd"} {
		err := paths.New("apparmor.d/groups/_full/" + name).CopyTo(RootApparmord.Join(name))
		if err != nil {
			return err
		}
	}
	logging.Success("Configure AppArmor for full system policy")
	return nil
}