package custom_gogit

import (
	gitv5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
	"os"
)

func CloneSubModule(url string, refname string) error {
	
	os.RemoveAll("/tmp/platform")
	publicKeys, err := ssh.NewPublicKeysFromFile("git", "/tmp/repo.pem", "")

	if err != nil {
		log.Info(err)
		return err
	}
	ref := plumbing.NewBranchReferenceName(refname)
	r, err := gitv5.PlainClone("/tmp/platform", false, &gitv5.CloneOptions{
		Auth:              publicKeys,
		URL:               url,
		RecurseSubmodules: gitv5.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
		ReferenceName:     ref,
	})
	if err != nil {
		log.Info(err)
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		log.Info(err)
		return err
	}
	sub, err := w.Submodule("manifest")
	if err != nil {
		log.Info(err)
		return err
	}
	_, err = sub.Repository()
	if err != nil {
		log.Info(err)
		return err
	}
	return nil

}
