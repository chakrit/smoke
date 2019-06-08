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

	for _, filename := range args {
		file, err := Load(filename)
		if err != nil {
			log.Println(err)
		}

		err = file.Run()
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
