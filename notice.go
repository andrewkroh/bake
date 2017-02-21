package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	pathSeparator  = string(os.PathSeparator)
	vendirDirMatch = "vendor" + pathSeparator
)

var noticeLog = logrus.WithField("package", "main").WithField("cmd", "notice")

func registerNoticeCommand(app *kingpin.Application) {
	cmd := getNoticeCommandDefaults()
	notice := app.Command("notice", "Create a NOTICE file containing the licenses of the project's vendored dependencies.").Action(cmd.Run)
	notice.Flag("beat", "Beat name").Short('b').Default(cmd.BeatName).StringVar(&cmd.BeatName)
	notice.Flag("copyright", "Copyright owner").Short('c').Default(cmd.Copyright).StringVar(&cmd.Copyright)
	notice.Flag("year", "Copyright begin year").Short('y').Default(strconv.Itoa(cmd.Year)).IntVar(&cmd.Year)
	notice.Flag("output", "Output file").Short('o').Default(cmd.Output).PlaceHolder(cmd.Output).StringVar(&cmd.Output)
	notice.Arg("dirs", "Directories to recursively search for vendored license files. Defaults to the project root.").Default(cmd.Dirs[0]).ExistingDirsVar(&cmd.Dirs)
}

type NoticeCommand struct {
	BeatName  string
	Copyright string
	Year      int // Copyright start year.
	Output    string
	Dirs      []string
}

func getNoticeCommandDefaults() *NoticeCommand {
	return &NoticeCommand{
		BeatName:  "Elastic Beats",
		Copyright: "Elasticsearch BV",
		Year:      2014,
		Output:    filepath.Join(ProjectRootRel, "NOTICE"),
		Dirs:      []string{filepath.Join(ProjectRootRel, ".")},
	}
}

func (c *NoticeCommand) Run(ctx *kingpin.ParseContext) error {
	noticeLog.WithField("output", c.Output).WithField("dirs", c.Dirs).Debug("Running notice")
	return generateNotice(c)
}

func generateNotice(cmd *NoticeCommand) error {
	licenses, err := searchForLicenses(cmd.Dirs)
	if err != nil {
		return err
	}

	noticeLog.WithField("licenses", licenses).Info("Found license files")

	var projects []*projectInfo
	for _, license := range licenses {
		p, err := getProjectInfo(license)
		if err != nil {
			return err
		}

		projects = append(projects, p)
	}

	p := noticeParams{
		BeatName:           cmd.BeatName,
		Copyright:          cmd.Copyright,
		CopyrightYearStart: cmd.Year,
		CopyrightYearEnd:   time.Now().Year(),
		Projects:           deduplicate(projects),
	}

	// Sort the libraries by name so that the list is stable.
	sortProjects(p.Projects)

	f, err := ioutil.TempFile("", "notice")
	if err != nil {
		return errors.Wrap(err, "failed to create temp file")
	}

	if err := noticeTemplate.Execute(f, p); err != nil {
		f.Close()
		os.Remove(f.Name())
		return errors.Wrap(err, "failed to populate template")
	}

	f.Sync()
	f.Close()
	os.Rename(f.Name(), cmd.Output)

	noticeLog.WithField("output", cmd.Output).Infof("Notice written")

	return nil
}

func getProjectInfo(license string) (*projectInfo, error) {
	contents, err := ioutil.ReadFile(license)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read license file")
	}

	// Dos2Unix
	contents = bytes.Replace(contents, []byte{'\r', '\n'}, []byte{'\n'}, -1)

	p := &projectInfo{Name: getLibraryName(license)}
	if bytes.Contains(contents, []byte("Apache License")) {
		p.License = "Apache License"
	} else {
		p.License = string(contents)
	}

	return p, nil
}

// getLibraryName returns project's Golang name by using the path after the
// last vendor directory.
func getLibraryName(path string) string {
	i := strings.LastIndex(path, vendirDirMatch)
	if i == -1 {
		return path
	}

	i += len(vendirDirMatch)
	if len(path) < i {
		return path
	}

	pkg := path[i:]
	pkg = filepath.Dir(pkg)
	strings.Replace(pkg, `\`, `/`, -1)
	return pkg
}

// searchForLicenses recursively search for LICENSE* files in vendor directories.
func searchForLicenses(dirs []string) ([]string, error) {
	var licenses []string

	findLicense := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if strings.Contains(path, ".git"+pathSeparator) {
			return nil
		}

		if !strings.Contains(path, vendirDirMatch) {
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), "LICENSE") {
			licenses = append(licenses, path)
		}

		return nil
	}

	for _, d := range dirs {
		if err := filepath.Walk(d, findLicense); err != nil {
			return nil, errors.Wrapf(err, "filepath walk failed for dir=%v", d)
		}
	}

	return licenses, nil
}

// Sorting

type Projects []*projectInfo

func (p Projects) Len() int      { return len(p) }
func (p Projects) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type ByName struct{ Projects }

func (p ByName) Less(i, j int) bool {
	nameI := strings.ToLower(p.Projects[i].Name)
	nameJ := strings.ToLower(p.Projects[j].Name)
	return nameI < nameJ
}

// sortProjects sorts the projectInfo objects by name.
func sortProjects(projects []*projectInfo) {
	sort.Sort(ByName{projects})
}

// deduplicate returns a deduplicated list of projects based on name.
func deduplicate(projects []*projectInfo) []*projectInfo {
	dedup := map[string]struct{}{}
	var out []*projectInfo
	for _, p := range projects {
		if _, found := dedup[p.Name]; !found {
			dedup[p.Name] = struct{}{}
			out = append(out, p)
		}
	}
	return out
}

// NOTICE Template

type noticeParams struct {
	BeatName           string
	Copyright          string
	CopyrightYearStart int
	CopyrightYearEnd   int
	Projects           []*projectInfo
}

type projectInfo struct {
	Name    string
	License string
}

var noticeTemplate = template.Must(template.New("notice").Parse(rawTemplate))

const rawTemplate = `{{.BeatName}}
Copyright {{.CopyrightYearStart}}-{{.CopyrightYearEnd}}

This product includes software developed by The Apache Software
Foundation (http://www.apache.org/).

==========================================================================
Third party libraries used by the {{.BeatName}}:
==========================================================================

{{range .Projects}}
--------------------------------------------------------------------
{{.Name}}
--------------------------------------------------------------------
{{.License}}
{{end}}`
