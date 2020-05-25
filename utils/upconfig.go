// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/olekukonko/tablewriter"
	"os"
	"path"
	"reflect"
	"strings"
)

type SecureSetting struct {
	Type   string
	Key    string
	KeyRef string
}

type Module struct {
	Repo       string
	Tag        string
	Version    string
	Alias      string
	Dir        string
	Subdir     string
	Iid        string
	PullPolicy string
}

type UpConfig struct {
	Version string
	RefDir  string
	//choice of cwd | refdir
	//default to be cwd
	WorkDir       string
	TaskFile      string
	Verbose       string
	ModuleName    string
	MaxCallLayers string
	Secure        *SecureSetting
	Modules       *[]Module
	ModuleLock    bool
}

type Modules []Module

type ModuleLockMap map[string]string

func (ms Modules) LocateModule(modname string) *Module {
	for _, m := range ms {
		m.Normalize()
		if m.Alias == modname {
			return &m
		}
	}
	return nil
}

func (m *Module) PullRepo(revMap *ModuleLockMap, uselock bool) {
	println("pull repo")
	m.Details()
	clone := func() {
		_, err := git.PlainClone(m.Dir, false, &git.CloneOptions{
			URL:      m.Repo,
			Progress: os.Stdout,
		})
		LogErrorAndExit("Clone Module", err, "Clone errored, please fix the issue first and retry")
	}
	Ptmpdebug("12", m.Dir)
	if _, err := os.Stat(m.Dir); !os.IsNotExist(err) {
		if m.PullPolicy == "always" {
			Pf("removing %s ...", m.Dir)
			err := os.RemoveAll(m.Dir)
			LogErrorAndExit("Remove directory", err, Spf("removing [%s] failed", m.Dir))
			clone()
		} else if m.PullPolicy == "skip" {
			LogWarn("module repo exist: skipped", Spf("repo: [%s]", m.Dir))
		} else if m.PullPolicy == "manual" {
			InvalidAndExit(Spf("repo: [%s] already exist", m.Dir),
				`manual resolution need:
1. You can git pull to update the module
2. If you work on the module, then you will need to commit and push your code accordingly, or
3. You will need to just delete it by yourself, or
4. Use pull policy: skip to not to do anything until you decide`)
		}
	} else {
		clone()
	}

	println("checkout version")
	if m.Version != "" {
		var versionDecided string

		if uselock {
			if lockedVersion, ok := (*revMap)[m.Alias]; ok {
				if lockedVersion != m.Version {
					if !strings.Contains(lockedVersion, m.Version) {
						LogWarn("Locked version differs, use locked version", Spf("locked: %s, configured: %s", lockedVersion, m.Version))
						versionDecided = lockedVersion
					}
				}
			}
		}

		if versionDecided == "" {
			versionDecided = m.Version
		}

		cmd := Spf("git checkout %s", versionDecided)
		Pf("checkout version: %s ...\n", versionDecided)
		Pln(cmd)
		err := RunShellCmd(m.Dir, cmd)
		if err != nil {
			LogWarn("checkout version", `
You may want to re-pull the repo again to ensure it is up to date to avoid missing branch, commit or tag
`)
		}
	}

}

func GetHeadRev(repodir string) string {
	r, err := git.PlainOpen(repodir)
	LogErrorAndExit("Open repo", err, Spf("please check repo:[%s]", repodir))
	h, err := r.ResolveRevision(plumbing.Revision("HEAD"))
	return (h.String())
}

func (m *Module) Details() {
	if m != nil {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"property", "value"})
		table.Append([]string{"alias", m.Alias})
		table.Append([]string{"dir", m.Dir})
		table.Append([]string{"repo", m.Repo})
		table.Append([]string{"version", m.Version})
		table.Append([]string{"pullpolicy", m.PullPolicy})
		table.Append([]string{"instanceid", m.Iid})
		table.Append([]string{"subdir", m.Subdir})
		table.Render()
	}
}

func (m *Module) Normalize() {
	if m.Dir != "" && m.Alias == "" {
		InvalidAndExit("module validation", Spf("You need to use a alias to name the module: dir [%s]", m.Dir))
	}

	if m.Iid == "" {
		m.Iid = "nonamed"
	}

	if m.Repo != "" {
		if m.Version == "" {
			m.Version = "master"
		}

		if m.PullPolicy == "" {
			m.PullPolicy = "skip"
		}

		if m.Alias == "" {
			m.Alias = GetGitRepoName(m.Repo)
		}

		if m.Dir == "" {
			m.Dir = path.Join(GetDefaultModuleDir(), m.Alias)
		}

		if m.Subdir != "" {
			if m.Alias == "" {
				InvalidAndExit("module validation", Spf("You need to use a alias to name the module: subdir [%s]", m.Subdir))
			}
		}

	}
}

func (cfg *UpConfig) SetVerbose(cmdV string) {
	if cmdV != "" {
		cfg.Verbose = cmdV
	}
}

func (cfg *UpConfig) SetRefdir(refdir string) {
	if refdir != "" {
		cfg.RefDir = refdir
	}
}

func (cfg *UpConfig) GetWorkdir() (wkdir string) {
	cwd, err := os.Getwd()
	if err != nil {
		LogErrorAndExit("GetWorkdir", err, "working directory error")
	}

	if cfg.WorkDir == "cwd" {
		wkdir = cwd
	} else if cfg.WorkDir == "refdir" {
		//assume refdir is relative path
		abpath := path.Join(cwd, cfg.RefDir)
		if _, err := os.Stat(abpath); !os.IsNotExist(err) {
			wkdir = abpath
		} else {
			if _, err := os.Stat(cfg.RefDir); !os.IsNotExist(err) {
				wkdir = cfg.RefDir
			}
		}
	} else {
		InvalidAndExit("GetWorkdir", "Work dir setup is not proper")
	}
	return
}

func (cfg *UpConfig) SetWorkdir(workdir string) {
	if workdir != "" {
		if Contains([]string{"cwd", "refdir"}, workdir) {
			cfg.WorkDir = workdir
		}
	} else {
		cfg.WorkDir = ""
	}
}

func (cfg *UpConfig) SetTaskfile(taskfile string) {
	if taskfile != "" {
		cfg.TaskFile = taskfile
	}
}

func (cfg *UpConfig) SetModulename(modulename string) {
	if modulename != "" {
		cfg.ModuleName = modulename
	}
}

func (cfg *UpConfig) ShowCoreConfig(mark string) {
	e := reflect.ValueOf(cfg).Elem()
	et := reflect.Indirect(e).Type()
	fmt.Printf("%s config:\n", mark)
	for i := 0; i < e.NumField(); i++ {
		if f := e.Field(i); f.Kind() == reflect.String {
			fname := et.Field(i).Name
			fmt.Printf("%20s -> %s\n", fname, f.String())
		}
	}
}

