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
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	//TODO: get rid of pointer as it will result in nil pointer loading issue
	Secure     *SecureSetting
	Modules    []Module
	ModuleLock bool
}

type Modules []Module

type ModuleLockMap map[string]string

func LoadModuleLockRevs() *ModuleLockMap {
	lockfile := "./modlock.yml"
	if _, err := os.Stat(lockfile); !os.IsNotExist(err) {
		yml, err := ioutil.ReadFile(lockfile)
		LogErrorAndExit("load locked file", err, "read file problem, please fix it")
		revs := ModuleLockMap{}
		err = yaml.Unmarshal(yml, &revs)
		LogErrorAndExit("load locked revs", err, "the lock file has got configuration problem, please fix it")
		return &revs
	} else {
		return nil
	}

}

func (ms Modules) LocateModule(modname string) *Module {
	for _, m := range ms {
		m.Normalize()
		if m.Alias == modname {
			return &m
		}
	}
	return nil
}

func (m *Module) getVersionAndPath() (string, string) {
	var versionDecided string
	lockMap := LoadModuleLockRevs()
	if m.Version != "" {
		if MainConfig.ModuleLock {
			if lockedVersion, ok := (*lockMap)[m.Alias]; ok {
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
	}

	clonePath := m.Dir
	if versionDecided != "" {
		clonePath = Spf("%s@%s", m.Dir, versionDecided)
	}

	return versionDecided, clonePath
}

func (m *Module) PullRepo(revMap *ModuleLockMap, uselock bool) {

	clonePath := m.Dir
	m.ShowDetails()
	clone := func() {
		_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
			URL:      m.Repo,
			Progress: os.Stdout,
		})
		LogErrorAndExit("Clone Module", err, "Clone errored, please fix the issue first and retry")
	}

	if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
		if m.PullPolicy == "always" {
			Pf("removing %s ...", clonePath)
			err := os.RemoveAll(clonePath)
			LogErrorAndExit("Remove directory", err, Spf("removing [%s] failed", clonePath))
			clone()
		} else if m.PullPolicy == "skip" {
			LogWarn("module repo exist: skipped", Spf("repo: [%s]", clonePath))
		} else if m.PullPolicy == "manual" {
			InvalidAndExit(Spf("repo: [%s] already exist", clonePath),
				`manual resolution need:
1. You can git pull to update the module
2. If you work on the module, then you will need to commit and push your code accordingly, or
3. You will need to just delete it by yourself, or
4. Use pull policy: skip to not to do anything until you decide`)
		}
	} else {
		clone()
	}

	Pln("checkout version")
	versionDecided := m.Version
	if versionDecided != "" {
		cmd := Spf("git checkout %s", versionDecided)
		Pf("checkout version: %s ...\n", versionDecided)
		Pln(cmd)
		err := RunShellCmd(clonePath, cmd)
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

func (m *Module) ShowDetails() {
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
			if m.Subdir != "" {
				InvalidAndExit("module validation", Spf("a alias is needed to avoid confusion i use subdir [%s]", m.Subdir))
			} else {
				m.Alias = GetGitRepoName(m.Repo)
			}
		}

		if m.Dir == "" {
			_, clonePath := m.getVersionAndPath()
			//m.Dir = path.Join(GetDefaultModuleDir(), m.Alias)
			m.Dir = Spf("%s%s", path.Join(GetDefaultModuleDir(), m.Alias), clonePath)
		}

		if m.Alias == "" {
			m.Alias = GetGitRepoName(m.Repo)
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

func (cfg *UpConfig) GetWorkdirOld() (wkdir string) {
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

//return abs path
func (cfg *UpConfig) GetWorkdir() (wkdir string) {
	cwd, err := os.Getwd()
	if err != nil {
		LogErrorAndExit("GetWorkdir", err, "working directory error")
	}

	if cfg.WorkDir == "cwd" {
		wkdir = cwd
	} else if cfg.WorkDir == "refdir" {
		if path.IsAbs(cfg.RefDir) {
			if _, err := os.Stat(cfg.RefDir); !os.IsNotExist(err) {
				wkdir = cfg.RefDir
			}
		} else {
			abspath := path.Clean(path.Join(cwd, cfg.RefDir))
			if _, err := os.Stat(abspath); !os.IsNotExist(err) {
				Pdebug(abspath)
				wkdir = abspath
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

