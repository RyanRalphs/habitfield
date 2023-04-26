package main

import (
	"fmt"
	"log"
	"os"
	"time"

	habit "github.com/RyanRalphs/habitfield"
)

func main() {
	userInput := os.Args[1:]
	writer := os.Stdout
	if len(userInput) == 0 {
		habit.PrintHelp(writer)
		log.Fatal("Exiting Program. Please try again after reading the above help message!")
	}
	input, err := habit.ProcessUserInput(userInput, writer)

	userHabit := habit.Habit{Name: input, LastRecordedEntry: time.Now(), Streak: 1}

	if err != nil {
		log.Fatal(err)
	}

	db, err := habit.OpenDatabase("habits")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}

	tracker := habit.NewTracker(db)

	record, err := tracker.GetHabit(userHabit)

	if err != nil {
		fmt.Printf("Habit %s does not exist. Creating habit...\n", userHabit.Name)
		record, err = tracker.AddHabit(userHabit)
		fmt.Printf("%+v", record)
		os.Exit(0)
	}

	if err != nil {
		fmt.Println(err)
	}

	record, err = tracker.UpdateHabit(record)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v", record)
}
