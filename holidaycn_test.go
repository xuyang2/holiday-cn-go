package holidaycn

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xuyang2/holiday-cn-go/pkg/holiday"
)

// IsNowHoliday checks if current time in China is a holiday
func IsNowHoliday() (bool, string, error) {
	now := time.Now().In(cnLocation)
	return IsRestDay(now)
}

// IsNowRestDay checks if current time in China is a rest day (holiday or weekend)
func IsNowRestDay() (bool, error) {
	now := time.Now().In(cnLocation)
	isRest, _, err := IsRestDay(now)
	return isRest, err
}

// IsRestDay checks if a given date is either a holiday or weekend
func IsRestDay(date time.Time) (bool, string, error) {
	year := date.Year()

	data := holiday.GetYearData(year)
	if data == nil {
		return false, "", fmt.Errorf("no holiday data for year %d", year)
	}

	dateStr := date.Format("2006-01-02")
	if day, exists := data[dateStr]; exists {
		return day.IsOffDay, day.Name, nil
	}

	// If not found in holiday data, use weekend logic
	return date.Weekday() == time.Saturday || date.Weekday() == time.Sunday, "", nil
}

// IsWorkday checks if a given date is a workday
func IsWorkday(date time.Time) (bool, error) {
	isRestDay, _, err := IsRestDay(date)
	if err != nil {
		return false, err
	}
	return !isRestDay, nil
}

// AfterWorkdays returns the next workday after counting the specified number of workdays from the start date
// For example:
// - If workdays=1, it counts one workday and returns the next workday after that
// - If workdays=2, it counts two workdays and returns the next workday after that
// Rest days (holidays and weekends) are skipped in the counting
func AfterWorkdays(startDate time.Time, workdays int) (time.Time, error) {
	if workdays < 0 {
		return time.Time{}, fmt.Errorf("workdays must be non-negative, got %d", workdays)
	}

	currentDate := startDate
	remainingWorkdays := workdays

	for remainingWorkdays >= 0 {
		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)

		// Check if it's a workday
		isWorkday, err := IsWorkday(currentDate)
		if err != nil {
			return time.Time{}, err
		}

		// Only decrement remaining workdays if it's a workday
		if isWorkday {
			remainingWorkdays--
		}
	}

	return currentDate, nil
}

func TestIsRestDay(t *testing.T) {
	tests := []struct {
		name            string
		date            time.Time
		wantIsRestDay   bool
		wantHolidayName string
		wantErr         bool
	}{
		{
			name:            "New Year 2025",
			date:            time.Date(2025, 1, 1, 0, 0, 0, 0, cnLocation),
			wantIsRestDay:   true,
			wantHolidayName: "元旦",
			wantErr:         false,
		},
		{
			name:            "Regular Workday",
			date:            time.Date(2025, 1, 2, 0, 0, 0, 0, cnLocation),
			wantIsRestDay:   false,
			wantHolidayName: "",
			wantErr:         false,
		},
		{
			name:            "Regular Weekend",
			date:            time.Date(2025, 1, 4, 0, 0, 0, 0, cnLocation), // Saturday
			wantIsRestDay:   true,
			wantHolidayName: "",
			wantErr:         false,
		},
		{
			name:            "Future Year",
			date:            time.Date(2030, 1, 1, 0, 0, 0, 0, cnLocation),
			wantIsRestDay:   false,
			wantHolidayName: "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsRestDay, gotHolidayName, err := IsRestDay(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsRestDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsRestDay != tt.wantIsRestDay {
				t.Errorf("IsRestDay() gotIsRestDay = %v, want %v", gotIsRestDay, tt.wantIsRestDay)
			}
			if gotHolidayName != tt.wantHolidayName {
				t.Errorf("IsRestDay() gotHolidayName = %v, want %v", gotHolidayName, tt.wantHolidayName)
			}
		})
	}
}

func TestIsWorkday(t *testing.T) {
	tests := []struct {
		name          string
		date          time.Time
		wantIsWorkday bool
		wantErr       bool
	}{
		{
			name:          "Regular Workday",
			date:          time.Date(2025, 1, 2, 0, 0, 0, 0, cnLocation),
			wantIsWorkday: true,
			wantErr:       false,
		},
		{
			name:          "Holiday",
			date:          time.Date(2025, 1, 1, 0, 0, 0, 0, cnLocation),
			wantIsWorkday: false,
			wantErr:       false,
		},
		{
			name:          "Weekend",
			date:          time.Date(2025, 1, 4, 0, 0, 0, 0, cnLocation), // Saturday
			wantIsWorkday: false,
			wantErr:       false,
		},
		{
			name:          "Future Year",
			date:          time.Date(2030, 1, 1, 0, 0, 0, 0, cnLocation),
			wantIsWorkday: false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsWorkday, err := IsWorkday(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsWorkday() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsWorkday != tt.wantIsWorkday {
				t.Errorf("IsWorkday() = %v, want %v", gotIsWorkday, tt.wantIsWorkday)
			}
		})
	}
}

func TestAfterWorkdays(t *testing.T) {
	tests := []struct {
		name        string
		startDate   time.Time
		workdays    int
		wantDate    time.Time
		wantErr     bool
		errContains string
	}{
		{
			name:      "One workday, next day is workday",
			startDate: time.Date(2025, 1, 2, 0, 0, 0, 0, cnLocation), // Thursday
			workdays:  1,
			wantDate:  time.Date(2025, 1, 6, 0, 0, 0, 0, cnLocation), // Next workday after counting 1 workday (Friday)
			wantErr:   false,
		},
		{
			name:      "One workday, starting Friday",
			startDate: time.Date(2025, 1, 3, 0, 0, 0, 0, cnLocation), // Friday
			workdays:  1,
			wantDate:  time.Date(2025, 1, 7, 0, 0, 0, 0, cnLocation), // Next workday after counting 1 workday (Monday)
			wantErr:   false,
		},
		{
			name:      "Two workdays starting Friday",
			startDate: time.Date(2025, 1, 3, 0, 0, 0, 0, cnLocation), // Friday
			workdays:  2,
			wantDate:  time.Date(2025, 1, 8, 0, 0, 0, 0, cnLocation), // Next workday after counting 2 workdays (Tue)
			wantErr:   false,
		},
		{
			name:      "One workday during Chinese New Year",
			startDate: time.Date(2025, 1, 27, 0, 0, 0, 0, cnLocation), // Monday before CNY
			workdays:  1,
			wantDate:  time.Date(2025, 2, 6, 0, 0, 0, 0, cnLocation), // Next workday after counting 1 workday during CNY
			wantErr:   false,
		},
		{
			name:        "Negative workdays",
			startDate:   time.Date(2025, 1, 2, 0, 0, 0, 0, cnLocation),
			workdays:    -1,
			wantErr:     true,
			errContains: "must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, err := AfterWorkdays(tt.startDate, tt.workdays)
			if (err != nil) != tt.wantErr {
				t.Errorf("AfterWorkdays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("AfterWorkdays() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if !gotDate.Equal(tt.wantDate) {
				t.Errorf("AfterWorkdays() = %v, want %v", gotDate, tt.wantDate)
			}
		})
	}
}
