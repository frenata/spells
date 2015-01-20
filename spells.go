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
	"strconv"
	"strings"
)

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

func Read(reader *csv.Reader) (s Spell, e error) {
	record, err := reader.Read()
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

	reg := regexp.MustCompile("^\\(.*\\)")
	mats := reg.FindString(record[7])
	s.Material = mats
	s.Description = strings.TrimSpace(strings.TrimPrefix(record[7], mats))

	return s, nil
}

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

	for {
		class := strings.TrimSuffix(path.Base(filename), ".csv")

		s, err := Read(reader)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
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
	var filters []string
	cliReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Please enter a command.")
		input, err := cliReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		input = strings.TrimSpace(input)

		switch {
		case input == "help":
			fmt.Println("exit - exits program")
			fmt.Println("load filename - loads csv file into memory")
			fmt.Println("filter request - filters the list according to request")
			fmt.Println("list - prints current filter list")
			fmt.Println("spellname - prints spell information")
		case input == "exit":
			fmt.Println("Exiting program, sir.")
			return
		case strings.HasPrefix(input, "load "):
			input = strings.TrimPrefix(input, "load ")
			fmt.Printf("Loading... %v\n", input)
			if input == "pre" {
				loadSpells(spellMap, "", true)
			} else {
				loadSpells(spellMap, input, false)
			}
		case strings.HasPrefix(input, "filter"):
			input = strings.TrimPrefix(input, "filter ")
			if input == "filter" {
				fmt.Printf("Filters: %v\n", filters)
			} else if input == "clear" {
				filters = []string{}
				fmt.Println("Clearing filtered list.")
			} else {
				filters = append(filters, input)
				fmt.Printf("Filtering... %v\n", filters)
			}
		case input == "list":
			fmt.Printf("Filters: %v\n", filters)
			fmt.Println(filterList(spellMap, filters))
		default:
			s, ok := spellMap[input]
			if ok {
				fmt.Println(s)
			} else {
				fmt.Println("Spell not found. Please try again.")
			}

		}
	}
}

func filterList(spellMap map[string]Spell, filters []string) (output string) {
	fMap := make(map[string]Spell)

	var test bool
	for n, s := range spellMap {
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
			case strings.HasPrefix(f, "school "):
				f = strings.TrimPrefix(f, "school ")
				if s.School != f {
					test = false
					break spellmap
				}
			case strings.HasPrefix(f, "class "):
				f = strings.TrimPrefix(f, "class ")
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
			fMap[n] = s
		}
	}

	for _, f := range fMap {
		output += fmt.Sprintf("%v\n\n", f)
	}
	return output
}

// TODO:
// DONE 1. create data struct
// DONE 2. Read in from csv files, populate slice of structs
// DONE 3. CLI utility to simply type name element and return the details.
// 4. Filtering commands, show all cantrips, or all rituals, or all wizards...
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
