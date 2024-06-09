package calendar

import (
	"fmt"
	tb "gopkg.in/telebot.v3"
	"strconv"
	"time"
)

type StrategyKey int

const (
	FromConfigStrategy StrategyKey = iota + 1
	MonthStrategy
	FromExistense
)

type Strategy interface {
	GetKeyboard(Rows) Rows

	GetMonthRow() tb.Row
	GetWeekdaysRow() tb.Row
	GetDays() Rows
	GetControls() tb.Row
}

type FromConfig struct {
	BaseCalendar
}

func (fs FromConfig) GetMonthRow() tb.Row {
	var monthRow tb.Row

	btn := tb.Btn{
		Unique: genUniqueParam("month_year_btn"),
		Text:   fmt.Sprintf("%s %v", fs.getMonthDisplayName(fs.currMonth), fs.currYear),
	}

	monthRow = append(monthRow, btn)

	return monthRow
}

func (fs FromConfig) GetWeekdaysRow() tb.Row {
	var weekdaysRow tb.Row

	for i, wd := range fs.getWeekdaysDisplayArray() {
		btn := tb.Btn{Unique: genUniqueParam("weekday_" + fmt.Sprint(i)), Text: wd}
		// TODO add ignore handler
		weekdaysRow = append(weekdaysRow, btn)
	}

	return weekdaysRow
}

func (fs FromConfig) GetDays() Rows {
	beginningOfMonth := time.Date(fs.currYear, fs.currMonth, 1, 0, 0, 0, 0, time.UTC)
	amountOfDaysInMonth := beginningOfMonth.AddDate(0, 1, -1).Day()

	var daysRows Rows

	// Calculating the number of empty buttons that need to be inserted forward
	weekdayNumber := int(beginningOfMonth.Weekday())
	if weekdayNumber == 0 && fs.opt.Language == RussianLangAbbr { // russian Sunday exception
		weekdayNumber = 7
	}

	// The difference between English and Russian weekdays order
	// en: Sunday (0), Monday (1), Tuesday (3), ...
	// ru: Monday (1), Tuesday (2), ..., Sunday (7)
	if fs.opt.Language != RussianLangAbbr {
		weekdayNumber++
	}

	var row tb.Row
	// Inserting empty buttons forward
	for i := 1; i < weekdayNumber; i++ {
		cell := tb.Btn{Unique: genUniqueParam("empty_cell"), Text: " "}
		// TODO add empty cell handler
		row = append(row, cell)
	}

	// Inserting month's days' buttons
	for i := 1; i <= amountOfDaysInMonth; i++ {
		dayText := strconv.Itoa(i)
		cell := tb.Btn{
			Unique: genUniqueParam("day_" + fmt.Sprint(i)),
			Text:   dayText, Data: dayText,
		}

		// TODO add day cell handler

		row = append(row, cell)

		if len(row)%AmountOfDaysInWeek == 0 {
			daysRows = append(daysRows, row)
			row = tb.Row{} // empty row
		}
	}

	// Inserting empty buttons at the end
	if len(row) > 0 {
		for i := len(row); i < AmountOfDaysInWeek; i++ {
			cell := tb.Btn{Unique: genUniqueParam("empty_cell"), Text: " "}
			row = append(row, cell)
		}
		daysRows = append(daysRows, row)
	}

	return daysRows
}

func (fs FromConfig) GetControls() tb.Row {
	var row tb.Row

	prev := tb.Btn{Unique: genUniqueParam("prev_month"), Text: "＜"}

	// Hide "prev" button if it rests on the range
	if fs.currYear <= fs.opt.YearRange[0] && fs.currMonth == 1 {
		prev.Text = ""
	}

	next := tb.Btn{Unique: genUniqueParam("next_month"), Text: "＞"}

	// Hide "next" button if it rests on the range
	if fs.currYear >= fs.opt.YearRange[1] && fs.currMonth == 12 {
		next.Text = ""
	}

	row = append(row, prev, next)

	return row
}

func (fs FromConfig) GetKeyboard(_ Rows) Rows {
	month := fs.GetMonthRow()
	weekDays := fs.GetWeekdaysRow()
	days := fs.GetDays()
	controls := fs.GetControls()

	out := append(Rows{month, weekDays}, days...)
	out = append(out, controls)

	return out
}

type MonthBoard struct {
	BaseCalendar
}

func (m MonthBoard) GetKeyboard(_ Rows) Rows {
	var (
		rows Rows
		row  tb.Row
	)

	// Generating a list of months
	for i := 1; i <= 12; i++ {
		monthName := m.getMonthDisplayName(time.Month(i))
		monthBtn := tb.Btn{
			Unique: genUniqueParam("month_pick_" + fmt.Sprint(i)),
			Text:   monthName, Data: strconv.Itoa(i),
		}

		row = append(row, monthBtn)

		// Arranging the months in 2 columns
		if i%2 == 0 {
			rows = append(rows, row)
			row = tb.Row{} // empty row
		}
	}

	return rows
}

func (m MonthBoard) GetMonthRow() tb.Row {
	return tb.Row{}
}

func (m MonthBoard) GetWeekdaysRow() tb.Row {
	return tb.Row{}
}

func (m MonthBoard) GetDays() Rows {
	return Rows{}
}

func (m MonthBoard) GetControls() tb.Row {
	return tb.Row{}
}

type FromExistent struct {
}

func (f FromExistent) GetKeyboard(rows Rows) Rows {
	return rows
}

func (f FromExistent) GetMonthRow() tb.Row {
	return tb.Row{}
}

func (f FromExistent) GetWeekdaysRow() tb.Row {
	return tb.Row{}
}

func (f FromExistent) GetDays() Rows {
	return Rows{}
}
func (f FromExistent) GetControls() tb.Row {
	return tb.Row{}
}
