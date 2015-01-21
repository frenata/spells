package main

// Spells reads from one or more csv files, creates a struct for each entry, and runs a web app
// that allows for filtering on the various elements of the struct.

// Perhaps this could be generalized later, but for now it's specific to DND 5E spell lists.

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
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
	if runtime.GOOS == "linux" {
		output += bold(s.Name)
	} else {
		output += s.Name
	}
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

// Reads a single line of csv
func read(reader *csv.Reader) (s Spell, e error) {
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

// ReadAll takes a csv file and a map of spell names to Spell structs and adds all lines in the csv into the map.
func ReadAll(filename string, spellMap map[string]Spell) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	reader.FieldsPerRecord = 8

	for {
		class := strings.TrimSuffix(path.Base(filename), ".csv")

		s, err := read(reader)
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

		if v, ok := spellMap[s.Name]; ok {
			v.Class = append(v.Class, strings.Title(class))
			s = v
		}
		spellMap[s.Name] = s

	}

	return nil
}

func loadSpells(spellMap map[string]Spell, filename string, pre bool) error {
	if !pre {
		err := ReadAll(filename, spellMap)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if pre {
		var files = []string{
			"csv/ranger.csv",
			"csv/druid.csv",
			"csv/bard.csv",
			"csv/cleric.csv",
			"csv/paladin.csv",
			"csv/sorcerer.csv",
			"csv/wizard.csv",
			"csv/warlock.csv",
		}

		for _, f := range files {
			err := ReadAll(f, spellMap)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

func cliInput(spellMap map[string]Spell) {
	var sortF string = "name"
	var filters []string
	cliReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nPlease enter a command.")
		input, err := cliReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		input = strings.TrimSpace(input)

		switch {
		case input == "help" || input == "h":
			fmt.Println("  exit              - exits program")
			fmt.Println("  load 'filename'   - loads csv file into memory")
			fmt.Println("  'spellname'       - prints spell information")
			fmt.Println("  list              - prints current filter list")
			fmt.Println("  sort              - directs the program how to sort Spells")
			fmt.Println("  filter            - filters the list according to request")
			fmt.Println("  help              - prints this help")
		case strings.HasPrefix(input, "sort"):
			if input == "sort" {
				fmt.Println("  Enter 'sort name' to sort spells by name. (default)")
				fmt.Println("  Enter 'sort level' to sort spells first by level, then by name.")
			} else if input == "sort name" {
				sortF = "name"
				fmt.Println("Now sorting by name.")
			} else if input == "sort level" {
				sortF = "level"
				fmt.Println("Now sorting by level.")
			}
		case input == "exit" || input == "quit" || input == "q" || input == "x":
			fmt.Println("Exiting program, sir.")
			return
		case strings.HasPrefix(input, "load "):
			input = strings.TrimPrefix(input, "load ")
			fmt.Printf("Loading... %v\n", input)
			if input == "dnd" {
				loadSpells(spellMap, "", true)
			} else {
				loadSpells(spellMap, input, false)
			}
		case strings.HasPrefix(input, "filter"):
			input = strings.TrimPrefix(input, "filter ")
			if input == "filter" {
				fmt.Printf("Current filters: %v\n", filters)
				fmt.Println("Options:")
				fmt.Println("  filter clear                    - clears the filter list")
				fmt.Println("  filter 0-9                      - only the specified level of spell")
				fmt.Println("  filter ritual                   - only ritual spells")
				fmt.Println("  filter concentration            - only spells that require Concentration")
				fmt.Println("  filter school=<NameOfSchool>    - only spells of the given school")
				fmt.Println("  filter class=<NameOfClass>      - only spells castable by the given class")
			} else if input == "clear" {
				filters = []string{}
				fmt.Println("Clearing filtered list.")
			} else {
				filters = append(filters, input)
				fmt.Printf("Filtering... %v\n", filters)
			}
		case input == "list":
			fmt.Printf("Filters: %v\n", filters)
			fmt.Println(filterList(spellMap, filters, sortF))
		default:
			// Need another function to check for spellname, or start of spellname, then return list.
			s, ok := spellMap[input]
			if ok {
				fmt.Println(s)
			} else {
				fmt.Println("Command not recognized. Please try again.")
			}

		}
	}
}

func filterList(spellMap map[string]Spell, filters []string, sortFunction string) (output string) {
	spells := make([]Spell, 0)

	var test bool
	for _, s := range spellMap {
		test = true
	spellmap:
		for _, f := range filters {
			//f = strings.TrimSpace(f)
			num, err := strconv.Atoi(f)
			if err != nil {
				num = -1
			}
			switch {
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
				//fmt.Println("test")
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
	if sortFunction == "name" {
		sort.Sort(ByName(spells))
	} else if sortFunction == "level" {
		sort.Sort(ByLevel(spells))

	}

	for _, s := range spells {
		output += fmt.Sprintf("%v\n\n", s)
	}
	return output
}

func bold(s string) string {
	return "\033[31m" + s + "\033[0m"
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

// TODO:
// DONE 1. create data struct
// DONE 2. Read in from csv files, populate slice of structs
// DONE 3. CLI utility to simply type name element and return the details.
// DONE 4. Filtering commands, show all cantrips, or all rituals, or all wizards...
// 5. Webapp - 3 column layout? filtering on the right, list of names on the left, data in the middle
func main() {
	fmt.Println("Welcome to Spells!")
	spellMap := make(map[string]Spell)

	/*if err := loadSpells(spellMap, "", true); err != nil {
		fmt.Println(err)
		return
	}*/

	cliInput(spellMap)
}
