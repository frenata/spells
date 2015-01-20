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

// TODO:
// DONE 1. create data struct
// DONE 2. Read in from csv files, populate slice of structs
// DONE 3. CLI utility to simply type name element and return the details.
// 4. Filtering commands, show all cantrips, or all rituals, or all wizards...
// 5. Webapp - 3 column layout? filtering on the right, list of names on the left, data in the middle
func main() {
	fmt.Println("Welcome to Spells!")
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

	spellMap := make(map[string]Spell)

	for _, f := range files {
		err := ReadAll(f, spellMap)
		if err != nil {
			fmt.Println(err)
		}
	}

	cliReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Please enter a spell.")
		input, err := cliReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		input = strings.TrimSpace(input)

		switch input {
		case "exit":
			fmt.Println("Exiting program, sir.")
			return
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
