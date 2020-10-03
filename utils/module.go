// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type Module struct {
	Repo       string
	Tag        string
	Version    string
	Alias      string
	Dir        string
	Subdir     string
	Iid        string
	PullPolicy string
	//ref to an env var name
	UsernameRef string
	PasswordRef string
}

type UpConfig struct {
	Version string
	RefDir  string
	//choice of cwd | refdir
	//default to be cwd
	WorkDir    string
	AbsWorkDir string
	TaskFile   string
	Verbose    string
	ModuleName string
	//default: /bin/sh, or the path given: /usr/local/bin/bash, or simply: GOSH
	ShellType           string
	MaxCallLayers       string
	Timeout             string
	MaxModuelCallLayers string
	//TODO: get rid of pointer as it will result in nil pointer loading issue
	Secure     *SecureSetting
	Modules    Modules
	ModuleLock bool
	//Exec Profile
	EntryTask          string
	Pure               bool
	ModRepoUsernameRef string
	ModRepoPasswordRef string
}

type Modules []Module

type ModuleLockMap map[string]string

func LoadModuleLockRevs() *ModuleLockMap {
	lockfile := "./modlock.yml"
	if _, err := os.Stat(lockfile); !os.IsNotExist(err) {
		yml, err := ioutil.ReadFile(lockfile)
		LogErrorAndPanic("load locked file", err, "read file problem, please fix it")
		revs := ModuleLockMap{}
		err = yaml.Unmarshal(yml, &revs)
		LogErrorAndPanic("load locked revs", err, "the lock file has got configuration problem, please fix it")
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

func (ms *Modules) PullModules() {

	lockMap := LoadModuleLockRevs()
	if ms != nil {
		for _, m := range *ms {
			m.Normalize()
			if m.Repo != "" {
				m.PullRepo(lockMap, MainConfig.ModuleLock)
			} else {
				Pf("module: [%s] uses directory: [%s]\n", m.Alias, m.Dir)
			}
		}
	}

}

func (ms *Modules) ReportModules() {
	if ms != nil && len(*ms) != 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"idx", "alias", "dir", "repo", "version", "pullpolicy", "instanceid", "subdir"})
		for idx, m := range *ms {
			m.Normalize()
			table.Append([]string{
				strconv.Itoa(idx + 1),
				m.Alias,
				m.Dir,
				m.Repo,
				m.Version,
				m.PullPolicy,
				m.Iid,
				m.Subdir,
			})
		}
		table.Render()
	}
}

func (ms *Modules) PullCascadedModules(clonedMainModList *[]string, clonedSubModList *[]string) {

	lockMap := LoadModuleLockRevs()
	if ms != nil {
		for _, m := range *ms {
			m.Normalize()

			if m.Repo != "" {
				if idx := IndexOf(*clonedMainModList, m.Alias); idx != -1 {
					if _, err := os.Stat(m.Dir); !os.IsNotExist(err) {
						cfg := NewUpConfig(m.Dir, "upconfig.yml")
						submods := &cfg.Modules
						*clonedMainModList = RemoveIndex(*clonedMainModList, idx)
						submods.PullCascadedModules(clonedMainModList, clonedSubModList)
					}
				}

				if !Contains(*clonedSubModList, m.Alias) {
					m.PullRepo(lockMap, MainConfig.ModuleLock)
					*clonedSubModList = append(*clonedSubModList, m.Alias)
					if _, err := os.Stat(m.Dir); !os.IsNotExist(err) {
						cfg := NewUpConfig(m.Dir, "upconfig.yml")
						submods := &cfg.Modules
						submods.PullCascadedModules(clonedMainModList, clonedSubModList)
					}
				}
			} else {
				Pf("module: [%s] uses directory: [%s]\n", m.Alias, m.Dir)
			}
		}
	}

}

func (ms *Modules) PullMainModules() (clonedList []string) {
	clonedList = []string{}

	lockMap := LoadModuleLockRevs()
	if ms != nil {
		for _, m := range *ms {
			m.Normalize()
			if m.Repo != "" {
				if !Contains(clonedList, m.Alias) {
					m.PullRepo(lockMap, MainConfig.ModuleLock)
					clonedList = append(clonedList, m.Alias)
				}
			} else {
				Pf("module: [%s] uses directory: [%s]\n", m.Alias, m.Dir)
			}
		}
	}

	return

}

func (m *Module) getVersionAndPath() (string, string) {
	var versionDecided string
	lockMap := LoadModuleLockRevs()
	if MainConfig.ModuleLock && lockMap != nil {
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
			Auth: func() *http.BasicAuth {
				auth := http.BasicAuth{}
				gu := MainConfig.ModRepoUsernameRef
				gp := MainConfig.ModRepoPasswordRef
				var gvalid, ivalid bool

				var guv, gpv string
				if gu != "" && gp != "" {
					guv = os.Getenv(gu)
					gpv = os.Getenv(gp)
					if guv != "" && gpv != "" {
						gvalid = true
					}
				}

				u := m.UsernameRef
				p := m.PasswordRef

				var uv, pv string
				if u != "" && p != "" {
					uv = os.Getenv(u)
					pv = os.Getenv(p)
					if uv != "" && pv != "" {
						ivalid = true
					}
				}

				if ivalid {
					auth.Username = uv
					auth.Password = pv
					return &auth
				}

				if gvalid {
					auth.Username = guv
					auth.Password = gpv
					return &auth
				}
				//fall back to empty auth for public accessible repo
				return &auth
			}(),
			URL:      m.Repo,
			Progress: os.Stdout,
		})
		LogErrorAndExit("Clone Module", err, `Clone errored, please fix the issue first and retry
Please either ues global repo settings:
	ModRepoUsernameRef: GIT_USERNAME
	ModRepoPasswordRef: GIT_PASSWORD

Or individual repo settings:
    UsernameRef: AUTH_TEST_MODULE_GIT_USERNAME
    PasswordRef: AUTH_TEST_MODULE_GIT_PASSWORD

They refer to the environment variable username and password
`)
	}

	if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
		if m.PullPolicy == "always" {
			Pf("removing %s ...", clonePath)
			err := os.RemoveAll(clonePath)
			LogErrorAndPanic("Remove directory", err, Spf("removing [%s] failed", clonePath))
			clone()
		} else if m.PullPolicy == "skip" {
			LogWarn("module repo exist: skipped", Spf("repo: [%s]", clonePath))
		} else if m.PullPolicy == "manual" {
			InvalidAndPanic(Spf("repo: [%s] already exist", clonePath),
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
		err := RunSimpleCmd(clonePath, cmd)
		if err != nil {
			LogWarn("checkout version", `
You may want to re-pull the repo again to ensure it is up to date to avoid missing branch, commit or tag
`)
		}
	}

}

func GetHeadRev(repodir string) string {
	r, err := git.PlainOpen(repodir)
	LogErrorAndPanic("Open repo", err, Spf("please check repo:[%s]", repodir))
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
		InvalidAndPanic("module validation", Spf("You need to use a alias to name the module: dir [%s]", m.Dir))
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
				InvalidAndPanic("module validation", Spf("a alias is needed to avoid confusion i use subdir [%s]", m.Subdir))
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
