package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/armon/go-radix"
)

// TreeMap represents a parsed nginx config(s)
// It is the parent struct for parsed data
type TreeMap struct {
	Payload *Payload
	tree    *radix.Tree
	mu      sync.Mutex
	//serverCount   int
	//locationCount int
	// TODO: count derived from len of below?
	locations map[*Directive]struct{}
	servers   map[*Directive]struct{}
}

// NewTree creates a map wrapper around the payload
func NewTree(p *Payload) *TreeMap {
	tm := &TreeMap{Payload: p}
	tm.buildTree()
	return tm
}

// Inserter is the radix tree insert function type
type Inserter func(s string, v interface{}) (interface{}, bool)

// AType is our action type
type AType int

const (
	// ActionUnknown is an uninitialized Action
	ActionUnknown AType = iota

	// ActionInsert insert Directive(s)
	ActionInsert

	// ActionAppend append to Directive
	ActionAppend

	// ActionUpdate update directive
	ActionUpdate

	// ActionDelete deletes an action
	ActionDelete
)

// Changes may be ignored or deleted, I think.
// TODO: fish or cut bait with this
// NOTE: this would be a core of the "change list" to submit a group
//       of config changes as a transaction
type Changes struct {
	Act        AType
	Path       string
	Directives []*Directive
}

type ChangeSet struct {
	Name        string
	Description string
	Modules     []string
	Changes     []Changes
	Params      map[string]interface{}
}

// TODO: need way to disambiguate between path separation and location args in their name
const pathSep = ">" // "/"

// find a matching directive in the list
// and return its args
// TODO: make this a block method? and the name is not clear
func subVal(name string, blocks []*Directive) []string {
	for _, b := range blocks {
		if b.Directive == name {
			return b.Args
		}
	}
	return nil
}

// WalkBack ties a config path to its associated directives
// NOTE:  payload is implicit, as this will be created and used
//        within a specific payload instance
// TODO: ultimately use Directive pointer once debugged?
type WalkBack struct {
	Indicies  []int  // path to that block
	Index     int    // offset into final block slice -- NOTE could make it last entry of Indicies and check for end of slice
	Path      string // TODO: temp for now
	Directive *Directive
}

func (w WalkBack) String() string {
	return fmt.Sprintf("%v-%s = %q", w.Indicies, w.Directive.Directive, strings.Join(w.Directive.Args, " "))
}

// inject blocks into tree
// internal -- named to avoid "insert" for normal use
func (t *TreeMap) inject(Insert Inserter, path string, index []int, blocks []*Directive) {
	debugf("inject path: %s blocks: %d index: %v\n", path, len(blocks), index)
	for i, block := range blocks {
		if block.Directive == "#" {
			continue
		}
		if block.Directive == "include" {
			for _, include := range block.Includes {
				t.inject(Insert, path, []int{include}, t.Payload.Config[include].Parsed)
			}
			continue
		}
		// extend COPY of the index for any children
		sub := append([]int{}, index...)
		sub = append(sub, i)
		val := WalkBack{Indicies: index, Directive: block, Index: i, Path: path}

		// TODO: child is not used, what was I thinking?
		save := func(current, child string) {
			if child == "" {
				child = current
			}
			Insert(current, val)
			t.inject(Insert, child, sub, block.Block)
		}
		save(path+pathSep+block.Name(), "")
		if block.tag != "" {
			// no full path required as each tag unique
			save(block.Directive+"@"+block.tag, "")
		}
		switch block.Directive {
		case "server":
			if _, ok := t.servers[block]; !ok {
				t.servers[block] = struct{}{}
				// no full path required as each numbered server is unique
				into := fmt.Sprintf("%s#%d", block.Directive, len(t.servers))
				save(into, "")
			}
		case "location":
			if _, ok := t.locations[block]; !ok {
				t.locations[block] = struct{}{}
				// no full path required as each numbered location is unique
				into := fmt.Sprintf("%s#%d", block.Directive, len(t.locations))
				save(into, path)
			}
		}
	}
}

// BuildTree creates a radix tree into all the crossplane config nodes
func (t *TreeMap) buildTree() {
	t.tree = radix.New()
	config := t.Payload.Config[0]
	t.servers = make(map[*Directive]struct{})
	t.locations = make(map[*Directive]struct{})
	t.inject(t.tree.Insert, "", []int{0}, config.Parsed)
}

// ShowTree shows the payload config tree
// NOTE: this is effectively a dev tool and may end up being removed
// TODO: What should be exposed via API?
func (t *TreeMap) ShowTree(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(t.Payload.Config) > 0 {
		fmt.Fprintf(w, "\nincluded files:\n")
		for i, conf := range t.Payload.Config {
			fmt.Fprintf(w, "%2d: %s\n", i, conf.File)
		}
		fmt.Fprintln(w)
	}
	if t.tree == nil {
		t.buildTree()
	}
	debugf("tree entries: %d\n", t.tree.Len())
	debugf("tree memory: %d\n", unsafe.Sizeof(t.tree))
	walker := func(s string, v interface{}) bool {
		wb := v.(WalkBack)
		fmt.Fprintf(w, "=> K: %-60s -- V: %s\n", s, wb)
		return false
	}
	t.tree.Walk(walker)
}

// Append blocks to the given path
func (t *TreeMap) Append(path string, blocks ...*Directive) error {
	debugf("append path: %s -- %q\n", path, dirs(blocks...))
	wb, err := t.walkback(path)
	if err != nil {
		return err
	}
	var b *[]*Directive
	for i, x := range wb.Indicies {
		if i == 0 {
			b = &(t.Payload.Config[x].Parsed)
			continue
		}
		b = &((*b)[x].Block)
	}
	block := (*b)[wb.Index]
	(*block).Block = append((*block).Block, blocks...)
	// refresh the tree for the new/shifted blocks
	debugf("\n\n APPEND PATH:%s", path)
	t.inject(t.tree.Insert, path, wb.Indicies, block.Block)
	return nil
}

// get path meta info
func (t *TreeMap) walkback(path string) (WalkBack, error) {
	b, ok := t.tree.Get(path)
	if !ok {
		return WalkBack{}, fmt.Errorf("bad path: %q", path)
	}
	wb, ok := b.(WalkBack)
	if !ok {
		return WalkBack{}, fmt.Errorf("wanted %T - got %T", WalkBack{}, wb)
	}
	return wb, nil
}

func dirs(dd ...*Directive) string {
	s := make([]string, len(dd))
	for i, d := range dd {
		s[i] = d.Directive
	}
	return strings.Join(s, ", ")
}

// Insert before the directive indicated by the path
func (t *TreeMap) Insert(path string, inserts ...*Directive) error {
	debugf("insert path: %s -- %q\n", path, dirs(inserts...))
	wb, err := t.walkback(path)
	if err != nil {
		return err
	}
	debugf("WALKBACK: %+v\n", wb)
	debugf("HELLO inserts: %v", inserts)

	var block *Directive
	blocks := &(t.Payload.Config[0].Parsed)
	idx := wb.Indicies
	for i, x := range idx {
		switch i {
		case 0:
			// first round is the config in question
			blocks = &(t.Payload.Config[x].Parsed)
		case 1:
			block = ((*blocks)[x])
			debugf("block %d: %s %v\n", i, block.Directive, block.Block)
		default:
			block = ((*block).Block[x])
			debugf("block %d: %s %v\n", i, block.Directive, block.Block)
		}
	}
	into := block.Directive
	if len(block.Args) > 0 {
		into += " " + strings.Join(block.Args, " ")
	}
	debugf("Inserting into %q, at %d/%d\n", into, wb.Index+1, len(block.Block))
	if err := block.Insert(wb.Index, inserts...); err != nil {
		return fmt.Errorf("path: %s index: %d err:%w", path, wb.Index, err)
	}
	t.inject(t.tree.Insert, path, idx, block.Block)
	return nil
}

// Delete removes the directive (and children, if any) at the given path
func (t *TreeMap) Delete(path string) error {
	debugf("deleting path: %s\n", path)
	wb, err := t.walkback(path)
	if err != nil {
		return err
	}
	kill := wb.Directive
	var b *[]*Directive
	for i, x := range wb.Indicies {
		if i == 0 {
			b = &(t.Payload.Config[x].Parsed)
			continue
		}
		b = &((*b)[x].Block)
	}
	debugf("deleting from %v\n", b)
	for i, b2 := range *b {
		if b2.Equal(kill) {
			//debugf("KILLING: %v\n", kill)
			*b = append((*b)[:i], (*b)[i+1:]...)
			t.tree.Delete(path)
			return nil
		}
	}
	return errors.New("could not delete from path")

}

// ChangeConfig allows a transaction of multiple configuration changes
// TODO: ensure it is truly atomic (any failure results in rollback)
func (t *TreeMap) ChangeConfig(changes ...Changes) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, change := range changes {
		debugf("\n\n\nCHANGING %d/%d: %q (%d) -> %q\n", i+1, len(changes), change.Path, change.Act, dirs(change.Directives...))
		switch change.Act {
		case ActionInsert:
			if err := t.Insert(change.Path, change.Directives...); err != nil {
				return err
			}
		case ActionAppend:
			if err := t.Append(change.Path, change.Directives...); err != nil {
				return err
			}
		case ActionDelete:
			if err := t.Delete(change.Path); err != nil {
				return err
			}
		default:
			return fmt.Errorf("action #%d -- %+v is not supported at this time", i, change.Act)
		}
	}
	/*
		if Debugging {
			t.ShowTree(os.Stderr)
		}
	*/
	return nil
}

// ChangesLoad unmarshals a json collection of changes
func ChangesLoad(r io.Reader) ([]Changes, error) {
	var c []Changes
	return c, json.NewDecoder(r).Decode(&c)
}

// ChangesFromFile unmarshals a json collection of changes from the given file
func ChangesFromFile(filename string) ([]Changes, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ChangesLoad(f)
}

// WalkFunc represents a radix walk func
type WalkFunc func(string, interface{}) bool

// Matcher evaluates a Directive and returns true if it matches conditions
type Matcher func(*Directive) bool

// Apply is a general purpose function to modify the tree
func (t *TreeMap) Apply(path string, m Matcher, blocks ...*Directive) error {
	wlk := func(path string, obj interface{}) bool {
		debugf("APPLY WALK PATH: %s\n", path)
		wb := obj.(WalkBack)
		if m(wb.Directive) {
			debugf("MATCH APPLYING: %v\n", blocks)
			wb.Directive.Block = append(wb.Directive.Block, blocks...)
			return true
		}
		return false
	}
	t.tree.WalkPrefix(path, wlk)

	return nil
}

// Get returns the value by the give path
func (t *TreeMap) Get(s string) (interface{}, error) {
	if t.tree == nil {
		t.buildTree()
	}
	v, ok := t.tree.Get(s)
	if !ok {
		prefix, what, found := t.tree.LongestPrefix(s)
		fmt.Printf("\nPREFIX: %s (%t)\n", prefix, found)
		wb, proper := what.(WalkBack)
		if !proper {
			return nil, fmt.Errorf("entry not found: %q", s)
		}
		exprLeft := strings.Index(s, "[")
		if exprLeft < 0 {
			return nil, fmt.Errorf("missing path expression")
		}
		exprRight := strings.Index(s, "]")
		if exprRight < 0 {
			return nil, fmt.Errorf(`missing close of expression ']"`)
		}
		if false {
			from := len(prefix) + len(pathSep)
			fmt.Printf("FROM: %v  WB: %+v\n", from, wb)
		}
		//remains := s[from:exprLeft]
		//fmt.Printf("REMAINS: %q\n", remains)
		//fmt.Printf("\nDIRECTIVE: %+v\n", wb.Directive)
		expr := s[exprLeft+1 : exprRight]
		children := strings.Split(expr, ",")
		debugf("EXPR: %q\n", children)
		return nil, fmt.Errorf("entry not found: %q", s)
	}
	wb, ok := v.(WalkBack)
	if !ok {
		return nil, fmt.Errorf("not a WalkBack: %T", v)
	}
	b := wb.Directive
	if len(b.Args) > 0 {
		return strings.Join(b.Args, "-"), nil
	}
	return b, nil
}

// ChangeMe modifies a config with a changeset
func ChangeMe(conf, edit string) (*TreeMap, error) {
	debugf("MODIFY: %s\n", conf)
	debugf("USING : %s\n", edit)
	// TODO: identify where to put mutex protection (here?)
	var catchErrors, single, comment bool
	var ignore []string
	p, err := ParseFile(conf, ignore, catchErrors, single, comment)
	if err != nil {
		log.Printf("Whelp, parsing file %s: %v", conf, err)
		return nil, err
	}

	f, err := os.Open(edit)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %q -- %w", edit, err)
	}
	defer f.Close()

	var changes []Changes
	if err = json.NewDecoder(f).Decode(&changes); err != nil {
		return nil, fmt.Errorf("json decode fail: %w", err)
	}
	tm := &TreeMap{Payload: p}
	tm.buildTree()

	if err = tm.ChangeConfig(changes...); err != nil {
		return nil, err
	}

	return tm, tm.Payload.Render(os.Stdout)
}
