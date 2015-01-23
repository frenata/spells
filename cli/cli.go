package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/frenata/spells"
)

// user input loop
func cliInput(sm *spells.SpellMap) {
	var sorter string = "level"
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
			fmt.Println("  sort              - directs the program how to sort Spells")
			fmt.Println("  filter            - filters the list according to request")
			fmt.Println("  list              - prints the current filtered list of spells")
			fmt.Println("  help              - prints this help")
		case strings.HasPrefix(input, "sort"):
			if input == "sort" {
				fmt.Println("  Enter 'sort level' to sort spells by level. (default)")
				fmt.Println("  Enter 'sort name' to sort spells by name.")
			} else if input == "sort name" {
				sorter = "name"
				fmt.Println("Now sorting by name.")
			} else if input == "sort level" {
				sorter = "level"
				fmt.Println("Now sorting by level.")
			}
		case input == "exit" || input == "quit" || input == "q" || input == "x":
			fmt.Println("Exiting program, sir.")
			return
		case strings.HasPrefix(input, "load "):
			input = strings.TrimPrefix(input, "load ")
			fmt.Printf("Loading... %v\n", input)
			if input == "def" {
				sm.LoadSpells("", true)
			} else {
				sm.LoadSpells(input, false)
			}
		case strings.HasPrefix(input, "filter"):
			input = strings.TrimPrefix(input, "filter ")
			if input == "filter" {
				fmt.Printf("Current filters: %v\n", sm.Filters)
				fmt.Println("Options:")
				fmt.Println("  filter clear                    - clears the filter list")
				fmt.Println("  filter list                     - prints the current filtered list")
				fmt.Println("  filter 0-9                      - only the specified level of spell")
				fmt.Println("  filter ritual                   - only ritual spells")
				fmt.Println("  filter concentration            - only spells that require Concentration")
				fmt.Println("  filter school=<NameOfSchool>    - only spells of the given school")
				fmt.Println("  filter class=<NameOfClass>      - only spells castable by the given class")
			} else if input == "clear" {
				sm.Filters = []string{}
				fmt.Println("Clearing filtered list.")
			} else if input == "list" {
				fmt.Printf("Filters: %v\n", sm.Filters)
				fmt.Println(PrintSorted(sm.Filter(), sorter))
			} else {
				sm.Filters = append(sm.Filters, input)
				fmt.Printf("Filtering... %v\n", sm.Filters)
			}
		case input == "list" || input == "ls":
			fmt.Printf("Filters: %v\n", sm.Filters)
			fmt.Println(PrintSorted(sm.Filter(), sorter))
		default:
			list := sm.KeySearch(input)
			if len(list) == 0 {
				fmt.Println("No spell or command not recognized. Please try again.")
			} else {
				fmt.Println(PrintSorted(list, sorter))
			}
		}
	}
}

func PrintSorted(list []spells.Spell, sorter string) (output string) {
	if sorter == "name" {
		sort.Sort(spells.ByName(list))
	} else if sorter == "level" {
		sort.Sort(spells.ByLevel(list))
	}

	for _, s := range list {
		reg := regexp.MustCompile("^.*?\\,")
		str := reg.ReplaceAllStringFunc(s.String(), bold)
		output += fmt.Sprintf("%v\n\n", str)
	}
	return output
}

func bold(s string) string {
	return "\033[31m" + s[:len(s)-1] + "\033[0m" + ","
}

func readConfig(file string) (list []string, err error) {
	dir, _ := os.Getwd()

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("No config file found: %v\nNo default CSV files will be loaded.\n", path.Join(dir, "spells.cfg"))
	}
	s := string(b)

	lines := strings.Split(s, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "dir=") {
			dir = strings.TrimPrefix(l, "dir=")
		}
		if strings.HasPrefix(l, "list=") {
			files := strings.Split(strings.TrimPrefix(l, "list="), ";")
			for _, f := range files {
				list = append(list, path.Join(dir, f))
			}
		}
	}

	if len(list) == 0 { // No files read.
		return nil, errors.New("No valid CSV files found in spells.cfg file.")
	}

	return list, nil
}

// TODO:
// DONE 1. create data struct
// DONE 2. Read in from csv files, populate slice of structs
// DONE 3. CLI utility to simply type name element and return the details.
// DONE 4. Filtering commands, show all cantrips, or all rituals, or all wizards...
// 5. Webapp - 3 column layout? filtering on the right, list of names on the left, data in the middle
func main() {
	fmt.Println("Welcome to Spells!")
	s := spells.NewSpellMap()

	def, err := readConfig("spells.cfg")
	if err != nil {
		fmt.Println(err)
	} else {
		s.SetDefaults(def)
		s.LoadSpells("", true)
		fmt.Println("Default CSV files loaded.")
	}

	cliInput(s)
}
