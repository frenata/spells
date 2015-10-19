package spells

// Spells reads from one or more csv files, creates a struct for each entry, and runs a web app
// that allows for filtering on the various elements of the struct.

// Perhaps this could be generalized later, but for now it's specific to DND 5E spell lists.

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Spell is a basic data struct for storing all the relevent information about an RPG spell.
type Spell struct {
	Level         int
	Name          string
	Ritual        bool
	School        string
	Time          string
	Range         string
	Components    string
	Duration      string
	Concentration bool
	Description   string
	Material      string
	Class         []string
}

// String pretty prints the Spell struct.
func (s Spell) String() string {
	var output string

	output += s.Name
	if s.Level == 0 {
		output += fmt.Sprintf(", %v cantrip for %v\n", s.School, s.Class)
	} else {
		output += fmt.Sprintf(", Level %d %v spell for %v\n", s.Level, s.School, s.Class)
	}
	if s.Ritual {
		output += "Ritual\n"
	}
	output += fmt.Sprintf("Casting time: %v, Range: %v, Duration: %v\n", s.Time, s.Range, s.Duration)
	if s.Concentration {
		output += "Concentration\n"
	}
	output += fmt.Sprintf("Components: %v\n", s.Components)
	if s.Material != "" {
		output += fmt.Sprintf("Materials: %v\n", s.Material)
	}
	output += fmt.Sprint(s.Description)

	return output
}

// ByName implements sort.Interface for []Spell based on alphabetical sort on the Name field
type ByName []Spell

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// ByLevel implements sort.Interface for []Spell based on level, then alphabetical name.
type ByLevel []Spell

func (a ByLevel) Len() int      { return len(a) }
func (a ByLevel) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLevel) Less(i, j int) bool {
	if a[i].Level == a[j].Level {
		return a[i].Name < a[j].Name
	} else {
		return a[i].Level < a[j].Level
	}
}

// Reads a single line of csv
func newSpell(reader *csv.Reader) (s Spell, e error) {
	record, err := reader.Read()

	// For both these errors, add better context information for returning, plus check possible error conditions
	// for other fields
	if err != nil {
		return s, err
	}

	s.Level, err = strconv.Atoi(strings.TrimSpace(record[0]))
	if err != nil {
		return s, err
	}

	if strings.HasSuffix(record[1], " (Ritual)") {
		s.Ritual = true
	}
	s.Name = strings.TrimSpace(strings.TrimSuffix(record[1], " (Ritual)"))

	school := strings.Split(record[2], "level ")
	if len(school) > 1 {
		s.School = strings.TrimSpace(school[1])
	} else { // cantrip?
		school = strings.Split(record[2], " Cantrip")
		s.School = strings.TrimSpace(school[0])
	}

	if strings.HasPrefix(record[6], "Concentration, up to ") {
		s.Concentration = true
	}
	s.Duration = strings.TrimSpace(strings.TrimPrefix(record[6], "Concentration, up to "))

	s.Time = strings.TrimSpace(record[3])
	s.Range = strings.TrimSpace(record[4])
	s.Components = strings.TrimSpace(record[5])

	descrip := strings.TrimSpace(record[7])
	reg := regexp.MustCompile("^\\(.*?\\)")
	mats := reg.FindString(descrip)
	s.Material = mats
	s.Description = strings.TrimSpace(strings.TrimPrefix(descrip, mats))

	return s, nil
}

// struct for handling the map and filters
type SpellMap struct {
	list     map[string]Spell
	Filters  []string
	defaults []string
	Sorter   sort.Interface
}

// NewSpellMap returns a new SpellMap ready to use and with sorting by level.
func NewSpellMap() *SpellMap {
	sm := new(SpellMap)
	sm.list = make(map[string]Spell)
	sm.Filters = make([]string, 0)
	sm.defaults = make([]string, 0)

	return sm
}

// SetDefaults sets a list of filenames as the default CSV files to read.
func (sm *SpellMap) SetDefaults(files []string) {
	sm.defaults = files
}

// ReadAll takes a csv file and a map of spell names to Spell structs and adds all lines in the csv into the map.
func (sm *SpellMap) readAll(filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	reader.FieldsPerRecord = 9 //8

	for {
		class := strings.TrimSuffix(path.Base(filename), ".csv")

		s, err := newSpell(reader)
		if err == io.EOF {
			break
		} else if err != nil {
			// read() should return more context, and ReadAll should add filename context.
			// Print enough information that user can see exactly where the csv is malformed.
			fmt.Println(err)
			continue //Don't quit reading the file just because of one error.
		} else {
			s.Class = append(s.Class, strings.Title(class))
		}

		if v, ok := sm.list[s.Name]; ok {
			v.Class = append(v.Class, strings.Title(class))
			s = v
		}
		sm.list[s.Name] = s

	}

	return nil
}

// read a file, then load up all the csv entries in it. Optionally takes bool to designate defaults.
func (sm *SpellMap) LoadSpells(filename string, pre bool) error {
	if !pre {
		err := sm.readAll(filename)
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		for _, f := range sm.defaults {
			err := sm.readAll(f)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

// KeySearch, given a string, searches for partial matches to keys in the map, returns a list of spells.
func (sm *SpellMap) KeySearch(partial string) (spells []Spell) {
	for k := range sm.list {
		if strings.HasPrefix(k, partial) {
			spells = append(spells, sm.list[k])
		}
	}

	return spells
}

// Filter filters the Spell map based on user-input filter and returns a list of spells for printing.
func (sm *SpellMap) Filter() (spells []Spell) {
	var test bool
	for _, s := range sm.list {
		test = true
	spellmap:
		for _, f := range sm.Filters {
			num, err := strconv.Atoi(f)
			if err != nil {
				num = -1
			}
			switch {
			case f == "bonus":
				if !strings.Contains(s.Time, "bonus") {
					test = false
					break spellmap
				}
			case f == "reaction":
				if !strings.Contains(s.Time, "reaction") {
					test = false
					break spellmap
				}
			case f == "ritual":
				if !s.Ritual {
					test = false
					break spellmap
				}
			case f == "concentration":
				if !s.Concentration {
					test = false
					break spellmap
				}
			case strings.HasPrefix(f, "school="):
				f = strings.TrimPrefix(f, "school=")
				if s.School != f {
					test = false
					break spellmap
				}
			case strings.HasPrefix(f, "class="):
				f = strings.TrimPrefix(f, "class=")
				classMatch := false
				for _, c := range s.Class {
					if c == f {
						classMatch = true
						break
					}
				}
				if !classMatch {
					test = false
					break spellmap
				}
			case num >= 0 && num <= 9:
				if s.Level != num {
					test = false
					break spellmap
				}
			}
		}
		if test {
			spells = append(spells, s)
		}
	}
	return spells
}
