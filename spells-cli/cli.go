package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/frenata/spells"
)

// player input loop
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

var defFiles = []string{
	"csv/bard.csv",
	"csv/cleric.csv",
	"csv/druid.csv",
	"csv/paladin.csv",
	"csv/ranger.csv",
	"csv/sorcerer.csv",
	"csv/warlock.csv",
	"csv/wizard.csv",
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
	s.SetDefaults(defFiles)
	cliInput(s)
}
