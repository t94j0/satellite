package path

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/satellite/geoip"
)

// Paths is the compilation of parsed paths
type Paths struct {
	base                 string
	pathsList            string
	dbRoot               string
	globalConditionsPath string

	state   *State
	geoipDB geoip.DB
	list    []*Path
}

// New creates a new Paths variable from the specified base path
func New(serverRoot, pathsList, dbPath, gcp string) (*Paths, error) {
	list := make([]*Path, 0)

	state, err := NewState(path.Join(serverRoot, dbPath))
	if err != nil {
		return nil, err
	}

	ret := &Paths{
		base:                 serverRoot,
		pathsList:            path.Join(serverRoot, pathsList),
		dbRoot:               dbPath,
		globalConditionsPath: gcp,

		list:  list,
		state: state,
	}

	if err := ret.Reload(); err != nil {
		return ret, err
	}

	return ret, nil
}

// NewDefault instantiates a Paths object with default configuration
func NewDefault(serverRoot, gcp string) (*Paths, error) {
	return New(serverRoot, "pathList.yml", ".db", gcp)
}

// NewDefaultTest For many of the tests, we don't need to apply the global conditionals, so this helper function is for test cases
func NewDefaultTest(serverRoot string) (*Paths, error) {
	return New(serverRoot, "pathList.yml", ".db", "")
}

// AddGeoIP adds the GeoIP path to this location
func (paths *Paths) AddGeoIP(path string) error {
	db, err := geoip.New(path)
	if err != nil {
		return err
	}
	paths.geoipDB = db

	return nil
}

// Len gets the number of paths
func (paths *Paths) Len() int {
	return len(paths.list)
}

// Match matches a page given a URI. It returns the specified Path and a boolean
// value to determine if there was a page that matched the URI
func (paths *Paths) Match(uri string) (*Path, bool) {
	for _, v := range paths.list {
		g := glob.MustCompile(v.Path, '/')
		if g.Match(uri) {
			if _, err := os.Stat(path.Join(paths.base, v.Path)); err != nil {
				v.HostedFile = v.Path
			} else {
				v.HostedFile = uri
			}
			return v, true
		}
	}
	info, err := os.Stat(path.Join(paths.base, uri))
	if err == nil && !info.IsDir() {
		return &Path{Path: uri, HostedFile: uri}, true
	}
	return nil, false
}

// ingestPathList adds the proxy from target path if it exists
func (paths *Paths) ingestPathList() ([]*Path, error) {
	pathsList := paths.pathsList
	if _, err := os.Stat(pathsList); os.IsNotExist(err) {
		// Do not fail if the pathList does not exist
		return []*Path{}, nil
	}

	pathArr, err := NewPathArray(pathsList)
	if err != nil {
		return []*Path{}, err
	}

	return pathArr, nil
}

func (paths *Paths) collectConditionalsDirectory(targetPath string) (RequestConditions, error) {
	condsResult := make([]RequestConditions, 0)

	collectWalkFunc := func(oPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		condData, err := ioutil.ReadFile(oPath)
		if err != nil {
			return err
		}

		conds, err := NewRequestConditions(condData)
		if err != nil {
			return err
		}

		condsResult = append(condsResult, conds)
		return nil
	}

	if err := filepath.Walk(targetPath, collectWalkFunc); err != nil {
		return RequestConditions{}, err
	}

	mergedConds, err := MergeRequestConditions(condsResult...)
	if err != nil {
		return RequestConditions{}, err
	}

	return mergedConds, nil
}

func (paths *Paths) applyGlobalConditionals(p *Path) error {
	gcp := paths.globalConditionsPath
	if gcp == "" {
		return nil
	}

	if f, err := os.Stat(gcp); err != nil || !f.IsDir() {
		return nil
	}

	globalConditions, err := paths.collectConditionalsDirectory(gcp)
	if err != nil {
		return err
	}

	newP, err := MergeRequestConditions(globalConditions, p.Conditions)
	if err != nil {
		return err
	}
	p.Conditions = newP

	return nil
}

func (paths *Paths) validate(pathList []*Path) error {
	for _, v := range pathList {
		// Ensure all path URI globbing compiles
		if _, err := glob.Compile(v.Path); err != nil {
			return errors.Wrap(err, "unable to compile glob: "+v.Path)
		}

		// Ensure paths are backed up by a file
		// fmt.Println(v.Path)
	}

	return nil
}

// Reload refreshes the list of paths internally to Paths
func (paths *Paths) Reload() error {
	pathsList, err := paths.ingestPathList()
	if err != nil {
		return err
	}

	if err := paths.validate(pathsList); err != nil {
		return err
	}

	paths.list = pathsList

	return nil
}

// Serve serves a page without checking conditionals
func (paths *Paths) Serve(w http.ResponseWriter, req *http.Request) error {
	uri := req.URL.Path
	targetPath, exists := paths.Match(uri)
	if !exists {
		return errors.New("not_found render page not found")
	}

	if err := targetPath.ServeHTTP(w, req, paths.base); err != nil {
		return err
	}
	return nil
}

// MatchAndServe matches a path, determines if the path should be served, and serves the file based on an HTTP request. If a failure occurs, this function will serve failed pages.
//
// This is a helper function which combines already-exposed functions to make file serving easy.
//
// Returns true when the file was served and false when a 404 page should be returned
func (paths *Paths) MatchAndServe(w http.ResponseWriter, req *http.Request) (bool, error) {
	uri := req.URL.Path

	matchedPath, exists := paths.Match(uri)
	if !exists {
		return false, nil
	}

	paths.applyGlobalConditionals(matchedPath)

	shouldHost := matchedPath.ShouldHost(req, paths.state, paths.geoipDB)
	if shouldHost {
		if err := matchedPath.ServeHTTP(w, req, paths.base); err != nil {
			return false, err
		}
		return true, nil
	}

	if matchedPath.FailRedirect(w, req) {
		return true, nil
	}

	matched, err := matchedPath.FailRender(w, req, func(uri string) *Path {
		newPath, found := paths.Match(matchedPath.OnFailure.Render)
		if !found {
			return nil
		}
		return newPath
	}, paths.base)
	if err != nil {
		return false, err
	}
	return matched, nil
}
