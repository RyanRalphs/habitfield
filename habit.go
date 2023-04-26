package habitfield

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
)

type Tracker struct {
	db *storm.DB
}

type Habit struct {
	ID                int    `storm:"id,increment"`
	Name              string `storm:"unique"`
	LastRecordedEntry time.Time
	Streak            int32
}

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

func (t *Tracker) AddHabit(name string) error {
	habit := Habit{Name: name, LastRecordedEntry: time.Now(), Streak: 1}

	if err := t.db.Save(&habit); err != nil {
		if errors.Is(err, storm.ErrAlreadyExists) {
			return fmt.Errorf("habit already exists: %s", name)
		}
		return fmt.Errorf("failed to save habit: %v", err)
	}

	return nil
}

func (t *Tracker) GetHabit(name string) (Habit, error) {
	var habit Habit
	if err := t.db.One("Name", name, &habit); err != nil {
		if err == storm.ErrNotFound {
			return habit, fmt.Errorf("habit not found: %s", name)
		}
		return habit, fmt.Errorf("failed to get habit: %v", err)
	}
	return habit, nil
}

func (t *Tracker) UpdateHabit(name string) (Habit, error) {
	habit, err := t.GetHabit(name)
	if err != nil {
		return habit, err
	}

	now := time.Now()
	if now.Day() == habit.LastRecordedEntry.Day() {
		return habit, fmt.Errorf("habit already recorded for today")
	}

	if now.Sub(habit.LastRecordedEntry).Hours() > 24 {
		fmt.Printf("Your streak for %s has been reset! Your previous streak was %d days - Try to beat it!\n", habit.Name, habit.Streak)

		habit.Streak = 0
	}

	habit.LastRecordedEntry = now
	habit.Streak++

	if err := t.db.Update(&habit); err != nil {
		return habit, fmt.Errorf("failed to update habit: %v", err)
	}

	fmt.Println("Habit updated!")
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
		"To add a habit, run `habit <habit>`.\n"+
		"To list all habits, run `habit list`.\n\n")
}

func OpenDatabase(databaseName string) (*storm.DB, error) {
	return storm.Open(databaseName)
}
