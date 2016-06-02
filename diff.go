package librarianpuppetgo

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/tmtk75/cli"
)

func findModIn(mods []Mod, m Mod) (Mod, error) {
	for _, i := range mods {
		if i.name == m.name {
			return i, nil
		}
	}
	return Mod{}, fmt.Errorf("missing %s", m.name)
}

const releaseBranchPattern = `release/0.([0-9]+)`

func increment(s string) (string, error) {
	re := regexp.MustCompile(releaseBranchPattern).FindAllStringSubmatch(s, -1)
	if len(re) == 0 {
		return "", fmt.Errorf("%v didn't match '%v'", s, releaseBranchPattern)
	}
	minor := re[0][1]
	v, err := strconv.Atoi(minor)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("release/0.%d", v+1), nil
}

func Diff(c *cli.Context, a, b string) {
	diff(c, a, b, func(oldm, newm Mod, oldref, newref string) {
		fmt.Println(newm.Dest(), oldref, newref)
		run2(os.Stdout, newm.Dest(), "git", []string{"--no-pager", "diff", "-w", oldref, newref})
	})
}

func parse(f string) []Mod {
	ar := newReader(f)
	mods, err := parsePuppetfile(ar)
	if err != nil {
		log.Fatalln(err)
	}
	return mods
}

type DiffFunc func(oldm, newm Mod, oldref, newref string)

func diff(c *cli.Context, oldfile, newfile string, f DiffFunc) {
	modulepath = c.String("modulepath")

	oldmods := parse(oldfile)
	newmods := parse(newfile)

	for _, newm := range newmods {
		oldm, err := findModIn(oldmods, newm)
		if err != nil {
			log.Printf("INFO: %v in %s\n", err, oldfile)
			continue
		}
		newref := newm.opts["ref"]
		if newref == "" {
			logger.Printf("INFO: missing ref in %v of %v", newm.name, newfile)
			continue
		}
		oldref := oldm.opts["ref"]
		if oldref == "" {
			logger.Printf("INFO: missing ref in %v of %v", oldm.name, oldfile)
			continue
		}
		f(oldm, newm, oldref, newref)
	}
}
