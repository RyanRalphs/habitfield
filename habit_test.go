package habitfield_test

import (
	"bytes"
	"os"
	"testing"

	habit "github.com/RyanRalphs/habitfield"
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
			got, err := habit.ProcessUserInput(test.input, fakeOutput)

			if err != nil {
				got = err.Error()
			}

			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

var ht *habit.Tracker
var dbName = "test.db"

func setupTest() func() {
	if _, err := os.Stat(dbName); err == nil {
		os.Remove(dbName)
	}
	db, err := habit.OpenDatabase(dbName)

	if err != nil {
		panic(err)
	}

	ht = habit.NewTracker(db)

	return func() {
		defer db.Close()
	}
}

func TestAddingANewHabit(t *testing.T) {
	defer setupTest()()
	err := ht.AddHabit("test")
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestRetrievingAStoredHabit(t *testing.T) {
	defer setupTest()()
	err := ht.AddHabit("test")
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	habit, err := ht.GetHabit("test")
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if habit.Name != "test" {
		t.Errorf("got %v, want %v", habit.Name, "test")
	}
}

func TestRetrievingAHabitThatDoesntExist(t *testing.T) {
	defer setupTest()()
	_, err := ht.GetHabit("test2")
	if err == nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestStreakUpdatingOfHabits(t *testing.T) {
	defer setupTest()()
	err := ht.AddHabit("test")
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	testHabit, err := ht.UpdateHabit("test")
	if err == nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if testHabit.Streak != 1 {
		t.Errorf("got %v, want %v", testHabit.Streak, 1)
	}
}
