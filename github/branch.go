package github

import (
	"fmt"
	"github.com/jingweno/gh/git"
	"regexp"
	"strings"
)

type Branch struct {
	Name string
}

func (b *Branch) ShortName() string {
	reg := regexp.MustCompile("^refs/(remotes/)?.+?/")
	return reg.ReplaceAllString(b.Name, "")
}

func (b *Branch) LongName() string {
	reg := regexp.MustCompile("^refs/(remotes/)?")
	return reg.ReplaceAllString(b.Name, "")
}

func (b *Branch) RemoteName() string {
	reg := regexp.MustCompile("^refs/remotes/([^/]+)")
	if reg.MatchString(b.Name) {
		return reg.FindStringSubmatch(b.Name)[1]
	}

	return ""
}

func (b *Branch) Upstream() (u *Branch, err error) {
	name, err := git.SymbolicFullName(fmt.Sprintf("%s@{upstream}", b.ShortName()))
	if err != nil {
		return
	}

	u = &Branch{name}

	return
}

func (b *Branch) IsRemote() bool {
	return strings.HasPrefix(b.Name, "refs/remotes")
}
