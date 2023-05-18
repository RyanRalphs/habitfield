package habitfield_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/RyanRalphs/habitfield"
)

func TestProcessUserInput(t *testing.T) {
	scenarios := []struct {
		name  string
		input []string
		want  string
	}{
		{
			name:  "Prints habit command",
			input: []string{"habit", "test"},
			want:  "test",
		},
		{
			name:  "Prints not a habit command",
			input: []string{"test"},
			want:  "test is not a habit command",
		},
		{
			name:  "Directs user to help if no habit provided",
			input: []string{"habit"},
			want:  "Habit is a command line tool for tracking habits. To get started, type 'habit help'",
		},
	}
	for _, test := range scenarios {
		t.Run(test.name, func(t *testing.T) {
			fakeOutput := &bytes.Buffer{}
			got, err := habitfield.ProcessUserInput(test.input, fakeOutput)

			if err != nil {
				got = err.Error()
			}

			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

var ht *habitfield.Tracker
var dbName = "test.db"
var testHabit = habitfield.Habit{
	Name:              "test",
	LastRecordedEntry: time.Now(),
	Streak:            1,
}

func setupTest() func() {
	if _, err := os.Stat(dbName); err == nil {
		os.Remove(dbName)
	}
	db, err := habitfield.OpenDatabase(dbName)

	if err != nil {
		panic(err)
	}

	ht = habitfield.NewTracker(db)

	return func() {
		defer db.Close()
	}
}

func TestAddingANewHabit(t *testing.T) {
	defer setupTest()()
	_, err := ht.AddHabit(testHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestRetrievingAStoredHabit(t *testing.T) {
	defer setupTest()()
	_, err := ht.AddHabit(testHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	habit, err := ht.GetHabit(testHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if habit.Name != "test" {
		t.Errorf("got %v, want %v", habit.Name, "test")
	}
}

func TestRetrievingAHabitThatDoesntExist(t *testing.T) {
	defer setupTest()()
	fakeHabit := habitfield.Habit{
		Name:              "fake",
		LastRecordedEntry: time.Now(),
		Streak:            0,
	}

	_, err := ht.GetHabit(fakeHabit)
	if err == nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestStreakUpdatingOfHabits(t *testing.T) {
	defer setupTest()()
	yesterdaysHabit := habitfield.Habit{
		Name:              "test",
		LastRecordedEntry: time.Now().AddDate(0, 0, -1),
		Streak:            1,
	}
	addedHabit, err := ht.AddHabit(yesterdaysHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	updatedHabit, err := ht.UpdateHabit(addedHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if updatedHabit.Streak != 2 {
		t.Errorf("got %v, want %v", updatedHabit.Streak, 2)
	}
}

func TestStreakDoesntIncrementIfAddedTwiceOnSameDay(t *testing.T) {
	defer setupTest()()
	addedHabit, err := ht.AddHabit(testHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	updatedHabit, err := ht.UpdateHabit(addedHabit)
	if err == nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if updatedHabit.Streak != 1 {
		t.Errorf("got %v, want %v", updatedHabit.Streak, 1)
	}
}

func TestStreakResetsIfNotAddedDaily(t *testing.T) {
	defer setupTest()()
	lastWeeksHabit := habitfield.Habit{
		Name:              "test",
		LastRecordedEntry: time.Now().AddDate(0, 0, -7),
		Streak:            7,
	}

	_, err := ht.AddHabit(lastWeeksHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	updatedHabit, err := ht.UpdateHabit(testHabit)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if updatedHabit.Streak != 1 {
		t.Errorf("got %v, want %v", updatedHabit.Streak, 1)
	}
}
