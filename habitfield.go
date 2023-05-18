package habitfield

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
)

type Tracker struct {
	db   *storm.DB
	add  bool
	list bool
	help bool
}

type Habit struct {
	ID                int    `storm:"id,increment"`
	Name              string `storm:"unique"`
	LastRecordedEntry time.Time
	Streak            int32
}

type option func(*Tracker) error

func NewTracker(db *storm.DB) *Tracker {
	return &Tracker{db: db}
}

func ProcessUserInput(userInput []string, writer io.Writer) (string, error) {
	if userInput[0] == "habit" {
		if len(userInput) > 1 {
			if userInput[1] == "help" {
				PrintHelp(writer)
				return "", fmt.Errorf("Hope this is helpful!")
			}

			if userInput[1] == "list" {
				ListHabits()
				return "", fmt.Errorf("Your habits are listed above")
			}
			habit := strings.Join(userInput[1:], " ")
			return habit, nil

		}
		return "", fmt.Errorf("Habit is a command line tool for tracking habits. To get started, type 'habit help'")
	}
	return "", fmt.Errorf("%s is not a habit command", userInput[0])
}

func (t *Tracker) AddHabit(habit Habit) (Habit, error) {
	if err := t.db.Save(&habit); err != nil {
		if errors.Is(err, storm.ErrAlreadyExists) {
			return habit, fmt.Errorf("habit already exists: %s", habit.Name)
		}
		return habit, fmt.Errorf("failed to save habit: %v", err)
	}

	return habit, nil
}

func (t *Tracker) GetHabit(habit Habit) (Habit, error) {
	if err := t.db.One("Name", habit.Name, &habit); err != nil {
		if err == storm.ErrNotFound {
			return habit, fmt.Errorf("habit not found: %s", habit.Name)
		}
		return habit, fmt.Errorf("failed to get habit: %v", err)
	}
	return habit, nil
}

func (t *Tracker) UpdateHabit(habit Habit) (Habit, error) {
	habit, err := t.GetHabit(habit)
	if err != nil {
		return habit, err
	}

	now := time.Now()
	if now.Day() == habit.LastRecordedEntry.Day() {
		return habit, fmt.Errorf("habit already recorded for today")
	}

	if now.Sub(habit.LastRecordedEntry).Hours() > 48 {
		habit.Streak = 0
	}

	habit.LastRecordedEntry = now
	habit.Streak++

	if err := t.db.Update(&habit); err != nil {
		return habit, fmt.Errorf("failed to update habit: %v", err)
	}

	return habit, nil
}

func (t *Tracker) ListHabits(writer io.Writer) error {
	var habits []Habit
	if err := t.db.All(&habits); err != nil {
		return fmt.Errorf("failed to list habits: %v", err)
	}

	if len(habits) == 0 {
		fmt.Fprintln(writer, "No habits found.")
		return nil
	}

	fmt.Fprintln(writer, "Habit streaks:")
	for _, habit := range habits {
		fmt.Fprintf(writer, "Habit %d: '%s' | Last Recorded On: %s with a streak of %d\n", habit.ID, habit.Name, habit.LastRecordedEntry.Format("02-01-2006"), habit.Streak)
	}

	return nil
}

func ListHabits() {
	db, err := OpenDatabase("habits")
	if err != nil {
		fmt.Println(err)
		return
	}
	NewTracker(db).ListHabits(os.Stdout)
}

func (t *Tracker) Close() error {
	return t.db.Close()
}

func PrintHelp(writer io.Writer) {
	fmt.Fprintf(writer, "Welcome to your personal habit tracker!!\n\n"+
		"To add a habit, run `-add <habitName>`.\n"+
		"To list all habits, run `-list`.\n\n")
}

func OpenDatabase(databaseName string) (*storm.DB, error) {
	return storm.Open(databaseName)
}

func Run() {
	if len(os.Args) == 1 {
		PrintHelp(os.Stdout)
	}

	db, err := OpenDatabase("habits")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	t := NewTracker(db)
	habit, err := t.FromArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	if t.add {
		record, err := t.GetHabit(habit)

		if err != nil {
			fmt.Printf("Adding habit: %s\n", habit.Name)
			record, err = t.AddHabit(habit)
			os.Exit(0)
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		fmt.Printf("Updating habit: %s\n", habit.Name)

		_, err = t.UpdateHabit(record)

		if err != nil {
			fmt.Println(err)
		}
	}
	if t.list {
		err := t.ListHabits(os.Stdout)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
	if t.help {
		PrintHelp(os.Stdout)
	}
}

func (t *Tracker) FromArgs(args []string) (Habit, error) {
	habit := Habit{LastRecordedEntry: time.Now(), Streak: 1}
	if len(args) == 0 {
		return habit, nil
	}

	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	add := fset.Bool("add", false, "Add a Habit")

	list := fset.Bool("list", false, "List all habits")

	help := fset.Bool("help", false, "Print help")

	err := fset.Parse(args)
	if err != nil {
		return habit, err
	}
	t.add = *add
	t.list = *list
	t.help = *help
	args = fset.Args()
	if len(args) < 1 {
		return habit, nil
	}

	habit.Name = strings.Join(args, " ")

	return habit, nil
}
