Spells is a CLI utilty to read, print, and filter spells from an RPG.

How To Use:
 1. Acquire/write some CSV files containing the spells you want the program to read. See example.csv for formatting.
 2. Run the program, type "help" for a list of commands. "Load 'yourfilename'" will load the specified CSV file.
 3. Optionally, you can create a folder holding your CSV files, and edit spells.cfg so that it points to the folder and the files within. These files will be loaded into program memory at start.
 4. Type the name or beginning of the name of the spell you want to see the details of, all matches should be printed to your screen.
 5. Try filtering the spells to see which meet specific criteria.

Features:
* Reads from specified csv files into program memory.
* Entering a spell name prints the relevant information about it, nicely formatted.
* Using the filter system allows the user to specify filters and print lists of spells that match.
* Autocompletion of names.
* Sorting functions to allow for sorting by level or name.
* User config file to specify CSV files for autoloading on program start.

TODO:
* Better error handling of malformed csv?
* Webapp implementation of CLI features. *** (CLI separated from main logic.) ***
