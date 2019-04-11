// Copyright © 2019 Alibaba Co. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/project"
	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
)

type syncCommand struct {
	cmd *cobra.Command
	ws  *workspace.WorkSpace

	O struct {
		ForceBroken            bool
		ForceSync              bool
		LocalOnly              bool
		NetworkOnly            bool
		DetachHead             bool
		CurrentBranchOnly      bool
		Jobs                   uint64
		ManifestName           string
		NoCloneBundle          bool
		ManifestServerUsername string
		ManifestServerPassword string
		FetchSubmodules        bool
		NoTags                 bool
		OptimizedFetch         bool
		Prune                  bool
		SmartSync              bool
		SmartTag               string
	}
}

func (v *syncCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "sync",
		Short: "Update working tree to the latest revision",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runE(args)
		},
	}
	v.cmd.Flags().BoolVarP(&v.O.ForceBroken,
		"force-broken",
		"f",
		false,
		"continue sync even if a project fails to sync")
	v.cmd.Flags().BoolVar(&v.O.ForceSync,
		"force-sync",
		false,
		"overwrite an existing git directory if it needs to "+
			"point to a different object directory. WARNING: this "+
			"may cause loss of data")
	v.cmd.Flags().BoolVarP(&v.O.LocalOnly,
		"local-only",
		"l",
		false,
		"only update working tree, don't fetch")
	v.cmd.Flags().BoolVarP(&v.O.NetworkOnly,
		"network-only",
		"n",
		false,
		"fetch only, don't update working tree")
	v.cmd.Flags().BoolVarP(&v.O.DetachHead,
		"detach",
		"d",
		false,
		"detach projects back to manifest revision")
	v.cmd.Flags().BoolVarP(&v.O.CurrentBranchOnly,
		"current-branch",
		"c",
		false,
		"fetch only current branch from server")
	v.cmd.Flags().Uint64VarP(&v.O.Jobs,
		"jobs",
		"j",
		v.defaultJobs(),
		fmt.Sprintf("projects to fetch simultaneously"))
	v.cmd.Flags().StringVarP(&v.O.ManifestName,
		"manifest-name",
		"m",
		"",
		"temporary manifest to use for this sync")
	v.cmd.Flags().BoolVar(&v.O.NoCloneBundle,
		"no-clone-bundle",
		false,
		"disable use of /clone.bundle on HTTP/HTTPS")
	v.cmd.Flags().StringVarP(&v.O.ManifestServerUsername,
		"manifest-server-username",
		"u",
		"",
		"username to authenticate with the manifest server")
	v.cmd.Flags().StringVarP(&v.O.ManifestServerPassword,
		"manifest-server-password",
		"p",
		"",
		"password to authenticate with the manifest server")
	v.cmd.Flags().BoolVar(&v.O.FetchSubmodules,
		"fetch-submodules",
		false,
		"fetch submodules from server")
	v.cmd.Flags().BoolVar(&v.O.NoTags,
		"no-tags",
		false,
		"don't fetch tags")
	v.cmd.Flags().BoolVar(&v.O.OptimizedFetch,
		"optimized-fetch",
		false,
		"only fetch projects fixed to sha1 if revision does not exist locally")
	v.cmd.Flags().BoolVar(&v.O.Prune,
		"prune",
		false,
		"delete refs that no longer exist on the remote")
	v.cmd.Flags().BoolVar(&v.O.SmartSync,
		"smart-sync",
		false,
		"smart sync using manifest from the latest known good build")
	v.cmd.Flags().StringVarP(&v.O.SmartTag,
		"smart-tag",
		"t",
		"",
		"smart sync using manifest from a known tag")

	return v.cmd
}

func (v *syncCommand) WorkSpace() *workspace.WorkSpace {
	if v.ws == nil {
		var err error
		v.ws, err = workspace.NewWorkSpace("")
		if err != nil {
			log.Fatal(err)
		}
	}
	return v.ws
}

func (v *syncCommand) defaultJobs() uint64 {
	rlimit := syscall.Rlimit{}
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	defaultJobs := min((rlimit.Cur-5)/3, config.MaxJobs)

	// When running test cases in cmd/, function `defaultJobs` will be evaluated.
	// Do not call `v.WorkSpace()` function, which will fail if not in a workspace.
	if v.ws == nil {
		v.ws, _ = workspace.NewWorkSpace("")
	}
	if v.ws != nil &&
		v.ws.Manifest != nil &&
		v.ws.Manifest.Default != nil &&
		v.ws.Manifest.Default.SyncJ > 0 {
		defaultJobs = min(defaultJobs, v.ws.Manifest.Default.SyncJ)
	}

	return defaultJobs
}

func (v syncCommand) CallManifestServerRPC() {
	// TODO
	log.Panic("not implement CallManifestServerRPC")
}

func (v syncCommand) updateManifestProject() error {
	return nil
	// TODO: sync manifest project
	/*
		mp = v.ws.ManifestProject
		mp.PreSync()
	*/

	/*
	   mp = self.manifest.manifestProject
	   mp.PreSync()

	   if opt.repo_upgraded:
	     _PostRepoUpgrade(self.manifest, quiet=opt.quiet)

	   if not opt.local_only:
	     start = time.time()
	     success = mp.Sync_NetworkHalf(quiet=opt.quiet,
	                                   current_branch_only=opt.current_branch_only,
	                                   no_tags=opt.no_tags,
	                                   optimized_fetch=opt.optimized_fetch,
	                                   submodules=self.manifest.HasSubmodules)
	     finish = time.time()
	     self.event_log.AddSync(mp, event_log.TASK_SYNC_NETWORK,
	                            start, finish, success)

	   if mp.HasChanges:
	     syncbuf = SyncBuffer(mp.config)
	     start = time.time()
	     mp.Sync_LocalHalf(syncbuf, submodules=self.manifest.HasSubmodules)
	     clean = syncbuf.Finish()
	     self.event_log.AddSync(mp, event_log.TASK_SYNC_LOCAL,
	                            start, time.time(), clean)
	     if not clean:
	       sys.exit(1)
	     self._ReloadManifest(manifest_name)
	     if opt.jobs is None:
	       self.jobs = self.manifest.default.sync_j
	*/
}

func (v syncCommand) NetworkHalf(allProjects []*project.Project) error {
	var err error

	// TODO 1: run go routine for multiple jobs
	// TODO 2: loop until all projects fetch, or the same projects failed twice
	// TODO 3: sort projects by its fetch time (reverse order),
	for _, projects := range project.GroupByName(allProjects) {
		for _, p := range projects {
			err = p.Fetch(nil)
			if err != nil && !v.O.ForceBroken {
				break
			}
		}
	}

	return nil
}

func (v syncCommand) checkoutEntries(entries *project.PathEntry) {
	p := entries.Project
	if p != nil {
		// TODO 1: mulple jobs using go routine
		// TODO 2: checkout project
		log.Notef("Start checkout %s", p.Path)
		p.Checkout(p.Revision, "")
	}
	for _, entry := range entries.Entries {
		v.checkoutEntries(entry)
	}
}

func (v syncCommand) LocalHalf(allProjects []*project.Project) error {
	entries := project.GroupByPath(allProjects)
	v.checkoutEntries(entries)

	return nil
}

// findObsoletePaths returns obsolete paths.
// Please note that the oldPaths and newPaths must be sorted.
func (v syncCommand) findObsoletePaths(oldPaths, newPaths []string) []string {
	result := []string{}

	i, j := 0, 0
	for i < len(oldPaths) && j < len(newPaths) {
		if oldPaths[i] < newPaths[j] {
			result = append(result, oldPaths[i])
			i++
		} else if oldPaths[i] > newPaths[j] {
			j++
		} else {
			i++
			j++
		}
	}
	for i < len(oldPaths) {
		result = append(result, oldPaths[i])
		i++
	}

	return result
}

func (v syncCommand) findGitWorktree(dir string) []string {
	result := []string{}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, err = os.Stat(filepath.Join(path, ".git")); err == nil {
				result = append(result, path)
				return filepath.SkipDir
			}
		}
		return nil
	})

	return result
}

func (v syncCommand) removeWorktree(dir string, gitTrees []string) error {
	var err error

	if len(gitTrees) == 0 {
		log.Printf("will remove %s", dir)
		return os.RemoveAll(dir)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Printf("will remove %s", path)
			err = os.Remove(path)
			if err != nil {
				return err
			}
			return nil
		}
		for _, p := range gitTrees {
			if path == p {
				return filepath.SkipDir
			}
			// TODO: seperator is /?
			if strings.HasPrefix(p, path+"/") {
				return nil
			}
		}
		log.Printf("will remove all %s", path)
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		return filepath.SkipDir
	})

	return err
}

func (v syncCommand) removeObsoletePaths(oldPaths, newPaths []string) error {
	sort.Strings(oldPaths)
	sort.Strings(newPaths)
	obsoletePaths := v.findObsoletePaths(oldPaths, newPaths)

	ws := v.WorkSpace()

	for _, p := range obsoletePaths {
		workdir := filepath.Clean(filepath.Join(ws.RootDir, p))
		gitdir := filepath.Join(workdir, ".git")
		workRepoPath := filepath.Clean(filepath.Join(ws.RootDir, config.DotRepo, p+".git"))

		if !strings.HasPrefix(workdir, ws.RootDir) {
			return fmt.Errorf("cannot delete project path '%s', which beyond repo root '%s'", workdir, ws.RootDir)
		}

		if _, err := os.Stat(gitdir); err != nil {
			continue
		}

		// Check if workdir is dirty or not
		r, err := git.PlainOpen(workdir)
		if err == git.ErrRepositoryNotExists {
			continue
		} else if err != nil {
			return fmt.Errorf("cannot open repository '%s': %s",
				p,
				err)
		}

		wt, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("fail to get worktree of '%s': %s",
				p,
				err)
		}

		status, err := wt.Status()
		if err != nil {
			return fmt.Errorf("fail to get worktree status of '%s': %s",
				p,
				err)
		}

		if !status.IsClean() {
			return fmt.Errorf(`Cannot remove project "%s": uncommitted changes are present.
Please commit changes, then run sync again`,
				p)
		}

		// Remove gitdir first
		err = os.RemoveAll(gitdir)
		if err != nil {
			return fmt.Errorf("fail to remove '%s': %s",
				gitdir,
				err)
		}

		// Remove worktree, except recursive git repositories
		ignoreRepos := v.findGitWorktree(workdir)
		err = v.removeWorktree(workdir, ignoreRepos)
		if err != nil {
			return fmt.Errorf("fail to remove '%s': %s", workdir, err)
		}

		// Remove project repository
		if _, err = os.Stat(workRepoPath); err != nil {
			if !strings.HasPrefix(workRepoPath, ws.RootDir) {
				return fmt.Errorf("cannot delete project repo '%s', which beyond repo root '%s'", workRepoPath, ws.RootDir)
			}
			err = os.RemoveAll(workRepoPath)
			if err != nil {
				return err
			}
		}

		// Repove object repository if has no other references
		// TODO: get object repository path, and drop it if no record in manifest
	}

	return nil
}

func (v syncCommand) UpdateProjectList() error {
	var (
		newPaths = []string{}
		oldPaths = []string{}
		ws       = v.WorkSpace()
	)

	allProjects, err := ws.GetProjects(&workspace.GetProjectsOptions{
		MissingOK:    true,
		SubmodulesOK: v.O.FetchSubmodules,
	})
	if err != nil {
		return err
	}

	for _, p := range allProjects {
		newPaths = append(newPaths, p.Path)
	}

	projectListFile := filepath.Join(ws.RootDir, config.DotRepo, "project.list")
	if _, err = os.Stat(projectListFile); err == nil {
		f, err := os.Open(projectListFile)
		defer f.Close()
		if err != nil {
			log.Fatalf("fail to open %s: %s", projectListFile, err)
		}
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString('\n')
			line = strings.TrimSpace(line)
			if line != "" {
				oldPaths = append(oldPaths, line)
			}
			if err != nil {
				break
			}
		}
	}

	err = v.removeObsoletePaths(oldPaths, newPaths)
	if err != nil {
		return err
	}

	projectListLockFile := projectListFile + ".lock"
	lockf, err := os.OpenFile(projectListLockFile,
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0644)
	if err != nil {
		return fmt.Errorf("fail to create lockfile '%s': %s", projectListLockFile, err)
	}
	defer lockf.Close()
	for _, p := range newPaths {
		_, err = lockf.WriteString(p + "\n")
		if err != nil {
			return fmt.Errorf("fail to save lockfile '%s': %s", projectListLockFile, err)
		}
	}
	lockf.Close()

	err = os.Rename(projectListLockFile, projectListFile)
	if err != nil {
		return fmt.Errorf("fail to rename lockfile to '%s': %s", projectListFile, err)
	}

	return nil
}

func (v syncCommand) runE(args []string) error {
	var (
		err error
	)

	ws := v.WorkSpace()

	if v.O.Jobs > 0 {
		v.O.Jobs = min(v.O.Jobs, v.defaultJobs())
	}
	if v.O.NetworkOnly && v.O.DetachHead {
		return newUserError("cannot combine -n and -d")
	}
	if v.O.NetworkOnly && v.O.LocalOnly {
		return newUserError("cannot combine -n and -l")
	}
	if v.O.ManifestName != "" && v.O.SmartSync {
		return newUserError("cannot combine -m and -s")
	}
	if v.O.ManifestName != "" && v.O.SmartTag != "" {
		return newUserError("cannot combine -m and -t")
	}
	if v.O.ManifestServerUsername != "" || v.O.ManifestServerPassword != "" {
		if !(v.O.SmartSync || v.O.SmartTag != "") {
			return newUserError("-u and -p may only be combined with -s or -t")
		}
		if v.O.ManifestServerUsername == "" || v.O.ManifestServerPassword == "" {
			return newUserError("both -u and -p must be given")
		}
	}

	if v.O.ManifestName != "" {
		ws.Override(v.O.ManifestName)
	}

	smartSyncManifestName := "smart_sync_override.xml"
	smartSyncManifestPath := filepath.Join(ws.ManifestProject.WorkDir, smartSyncManifestName)

	if v.O.SmartSync || v.O.SmartTag != "" {
		v.CallManifestServerRPC()
	} else {
		if _, err = os.Stat(smartSyncManifestPath); err == nil {
			err = os.Remove(smartSyncManifestPath)
			if err != nil {
				log.Fatalf("failed to remove existing smart sync override manifest: %s", smartSyncManifestPath)
			}
		}
	}

	err = v.updateManifestProject()
	if err != nil {
		return err
	}

	allProjects, err := ws.GetProjects(&workspace.GetProjectsOptions{
		MissingOK:    true,
		SubmodulesOK: v.O.FetchSubmodules,
	}, args...)

	if !v.O.LocalOnly {
		err = v.NetworkHalf(allProjects)
		if err != nil {
			return err
		}
	}

	if v.O.NetworkOnly ||
		ws.ManifestProject.MirrorEnabled() ||
		ws.ManifestProject.ArchiveEnabled() {
		return nil
	}

	// Remove obsolete projects
	if err = v.UpdateProjectList(); err != nil {
		log.Fatal(err)
	}

	err = v.LocalHalf(allProjects)
	if err != nil {
		return err
	}

	// If there's a notice that's supposed to print at the end of the sync,
	// print it now...
	if ws.Manifest.Notice != "" {
		log.Note(ws.Manifest.Notice)
	}

	return nil
}

var syncCmd = syncCommand{}

func init() {
	rootCmd.AddCommand(syncCmd.Command())
}