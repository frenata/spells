package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// pipeLess takes a string and prints it using the "less" program, returning any errors
func pipeLess(s string) error {
	r, w := io.Pipe()

	go func() {
		fmt.Fprint(w, s)
		w.Close()
	}()

	less := exec.Command("less", "-R")
	less.Stdin = r
	less.Stdout = os.Stdout
	if err := less.Start(); err != nil {
		return err
	}
	return less.Wait()
}
