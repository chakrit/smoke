package main

import (
	"log"

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
		file, err := Load(filename)
		if err != nil {
			log.Println(err)
		}

		result, err := file.Run()
		if err != nil {
			log.Println(err)
		}

		err = Report(result)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
