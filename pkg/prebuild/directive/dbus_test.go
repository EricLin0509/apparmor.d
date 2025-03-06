// apparmor.d - Full set of apparmor profiles
// Copyright (C) 2021-2024 Alexandre Pujol <alexandre@pujol.io>
// SPDX-License-Identifier: GPL-2.0-only

package directive

import (
	"testing"

	"github.com/roddhjav/apparmor.d/pkg/paths"
)

const dbusOwnSystemd1 = `  dbus bind bus=system name=org.freedesktop.systemd1{,.*},
  dbus receive bus=system path=/org/freedesktop/systemd1{,/**}
       interface=org.freedesktop.systemd1{,.*}
       peer=(name="@{busname}"),
  dbus send bus=system path=/org/freedesktop/systemd1{,/**}
       interface=org.freedesktop.systemd1{,.*}
       peer=(name="{@{busname},org.freedesktop.DBus}"),
  dbus (send receive) bus=system path=/org/freedesktop/systemd1{,/**}
       interface=org.freedesktop.DBus.Properties
       member={Get,GetAll,Set,PropertiesChanged}
       peer=(name="{@{busname},org.freedesktop.DBus}"),
  dbus receive bus=system path=/org/freedesktop/systemd1{,/**}
       interface=org.freedesktop.DBus.Introspectable
       member=Introspect
       peer=(name="@{busname}"),
  dbus receive bus=system path=/org/freedesktop/systemd1{,/**}
       interface=org.freedesktop.DBus.ObjectManager
       member=GetManagedObjects
       peer=(name="{@{busname},org.freedesktop.systemd1{,.*}}"),
  dbus send bus=system path=/org/freedesktop/systemd1{,/**}
       interface=org.freedesktop.DBus.ObjectManager
       member={InterfacesAdded,InterfacesRemoved}
       peer=(name="{@{busname},org.freedesktop.DBus}"),`

func TestDbus_Apply(t *testing.T) {
	tests := []struct {
		name    string
		opt     *Option
		profile string
		want    string
		wantErr bool
	}{
		{
			name: "own",
			opt: &Option{
				Name: "dbus",
				ArgMap: map[string]string{
					"bus":  "system",
					"name": "org.freedesktop.systemd1",
					"own":  "",
				},
				ArgList: []string{"own", "bus=system", "name=org.freedesktop.systemd1"},
				File:    paths.New("fake-own"),
				Raw:     "  #aa:dbus own bus=system name=org.freedesktop.systemd1",
			},
			profile: "  #aa:dbus own bus=system name=org.freedesktop.systemd1",
			want:    dbusOwnSystemd1,
		},
		{
			name: "own-interface",
			opt: &Option{
				Name: "dbus",
				ArgMap: map[string]string{
					"bus":        "session",
					"name":       "com.rastersoft.ding",
					"interface+": "org.gtk.Actions",
					"own":        "",
				},
				ArgList: []string{"own", "bus=session", "name=com.rastersoft.ding", "interface+=org.gtk.Actions"},
				File:    paths.New("fake-interface"),
				Raw:     "  #aa:dbus own bus=session name=com.rastersoft.ding interface+=org.gtk.Actions",
			},
			profile: "  #aa:dbus own bus=session name=com.rastersoft.ding interface+=org.gtk.Actions",
			want: `  dbus bind bus=session name=com.rastersoft.ding{,.*},
  dbus receive bus=session path=/com/rastersoft/ding{,/**}
       interface=com.rastersoft.ding{,.*}
       peer=(name="@{busname}"),
  dbus send bus=session path=/com/rastersoft/ding{,/**}
       interface=com.rastersoft.ding{,.*}
       peer=(name="{@{busname},org.freedesktop.DBus}"),
  dbus receive bus=session path=/com/rastersoft/ding{,/**}
       interface=org.gtk.Actions
       peer=(name="@{busname}"),
  dbus send bus=session path=/com/rastersoft/ding{,/**}
       interface=org.gtk.Actions
       peer=(name="{@{busname},org.freedesktop.DBus}"),
  dbus (send receive) bus=session path=/com/rastersoft/ding{,/**}
       interface=org.freedesktop.DBus.Properties
       member={Get,GetAll,Set,PropertiesChanged}
       peer=(name="{@{busname},org.freedesktop.DBus}"),
  dbus receive bus=session path=/com/rastersoft/ding{,/**}
       interface=org.freedesktop.DBus.Introspectable
       member=Introspect
       peer=(name="@{busname}"),
  dbus receive bus=session path=/com/rastersoft/ding{,/**}
       interface=org.freedesktop.DBus.ObjectManager
       member=GetManagedObjects
       peer=(name="{@{busname},com.rastersoft.ding{,.*}}"),
  dbus send bus=session path=/com/rastersoft/ding{,/**}
       interface=org.freedesktop.DBus.ObjectManager
       member={InterfacesAdded,InterfacesRemoved}
       peer=(name="{@{busname},org.freedesktop.DBus}"),`,
		},
		{
			name: "talk",
			opt: &Option{
				Name: "dbus",
				ArgMap: map[string]string{
					"bus":   "system",
					"name":  "org.freedesktop.Accounts",
					"label": "accounts-daemon",
					"talk":  "",
				},
				ArgList: []string{"talk", "bus=system", "name=org.freedesktop.Accounts", "label=accounts-daemon"},
				File:    paths.New("gdm-session-worker"),
				Raw:     "  #aa:dbus talk bus=system name=org.freedesktop.Accounts label=accounts-daemon",
			},
			profile: "  #aa:dbus talk bus=system name=org.freedesktop.Accounts label=accounts-daemon",
			want: `  dbus (send receive) bus=system path=/org/freedesktop/Accounts{,/**}
       interface=org.freedesktop.Accounts{,.*}
       peer=(name="{@{busname},org.freedesktop.Accounts{,.*}}", label=accounts-daemon),
  dbus (send receive) bus=system path=/org/freedesktop/Accounts{,/**}
       interface=org.freedesktop.DBus.Properties
       member={Get,GetAll,Set,PropertiesChanged}
       peer=(name="{@{busname},org.freedesktop.Accounts{,.*}}", label=accounts-daemon),
  dbus send bus=system path=/org/freedesktop/Accounts{,/**}
       interface=org.freedesktop.DBus.Introspectable
       member=Introspect
       peer=(name="{@{busname},org.freedesktop.Accounts{,.*}}", label=accounts-daemon),
  dbus send bus=system path=/org/freedesktop/Accounts{,/**}
       interface=org.freedesktop.DBus.ObjectManager
       member=GetManagedObjects
       peer=(name="{@{busname},org.freedesktop.Accounts{,.*}}", label=accounts-daemon),
  dbus receive bus=system path=/org/freedesktop/Accounts{,/**}
       interface=org.freedesktop.DBus.ObjectManager
       member={InterfacesAdded,InterfacesRemoved}
       peer=(name="{@{busname},org.freedesktop.Accounts{,.*}}", label=accounts-daemon),`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Directives["dbus"].Apply(tt.opt, tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dbus.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Dbus.Apply() = %v, want %v", got, tt.want)
			}
		})
	}
}
