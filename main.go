package main

import (
	"log"

	lib "github.com/chakrit/smoke/smokelib"
	"github.com/chakrit/smoke/specs"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "smoke [filename]",
	Short: "Runs smoke tests defined in the given filename.",
	Run:   runRootCmd,
}

func runRootCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalln("requires 1 filename")
		return
	}

	if args[0] == "help" {
		if err := cmd.Help(); err != nil {
			log.Fatalln(err)
		}
		return
	}

	defer log.Println("done.")
	for _, filename := range args {
		file, err := specs.Load(filename)
		if err != nil {
			log.Println(err)
		}

		tests, err := file.Tests()
		if err != nil {
			log.Println(err)
		}

		results, err := lib.RunTests(tests)
		if err != nil {
			log.Println(err)
		}

		// TODO: Report/print result
		_ = results
	}
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
