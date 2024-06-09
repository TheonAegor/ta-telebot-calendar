package calendar

import (
	"context"
	fsm "github.com/TheonAegor/ta-fsm-telebot"
	"github.com/TheonAegor/ta-fsm-telebot/fsmopt"
	"log"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	stateSetSecondDate fsm.State = "stateSetSecondDate"
)

type StateHandler func(cal ICalender, btn tele.Btn)

type ExtraHandler func(cal ICalender)

type ICalender interface {
	GetKeyboard() Rows
	GetStateKey() string
	GetButtons() UniqueButtonsMap
	GetCurrYear() int
	GetYearRange() [2]int

	SetStrategy(strategy StrategyKey)
	SetCurrMonth(month time.Month)
	SetCurrYear(y int)
	GetCurrMonth() time.Month
}

// NewStateCalendar builds and returns a StateCalendar
func NewStateCalendar(m *fsm.Manager, d fsm.Dispatcher, stateKey string, opt ...Option) *StateCalendar {
	so := DefaultStateOptions(m, d)

	for _, o := range opt {
		o(&so)
	}

	err := so.validate()
	if err != nil {
		panic(err)
	}
	bc := BaseCalendar{
		currYear:  so.InitialYear,
		currMonth: so.InitialMonth,
		opt:       &so,
	}
	calStrategies := map[StrategyKey]Strategy{
		FromConfigStrategy: FromConfig{BaseCalendar: bc},
		MonthStrategy:      MonthBoard{BaseCalendar: bc},
		FromExistense:      FromExistent{},
	}

	cal := &StateCalendar{
		Manager:    m,
		Dispatcher: d,
		kb:         make([][]tele.InlineButton, 0),
		opt:        &so,
		currYear:   so.InitialYear,
		currMonth:  so.InitialMonth,
		strategy:   calStrategies[FromConfigStrategy],
		strategies: calStrategies,
	}

	so.ExtraHandlers = append(so.ExtraHandlers, func(cal ICalender) {
		m.Bind(
			d,
			fsmopt.On(tele.OnCallback),
			fsmopt.OnStates(stateSetSecondDate),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				state.SetState(context.TODO(), fsm.AnyState)
				// find button edit, add dot
				dayInt, err := strconv.Atoi(c.Data())
				if err != nil {
					return err
				}

				var data [2]int
				state.Data(context.TODO(), cal.GetStateKey(), &data)
				state.Update(context.TODO(), cal.GetStateKey(), [2]int{data[0], dayInt})

				return c.Edit(c.Message().Text, cal.GetKeyboard())
			}),
		)
	})

	return cal
}

type StateBtn struct {
	b        *tele.Btn
	row, idx int

	seed int64
}

type UniqueButtonsMap map[string]StateBtn

func (ubm UniqueButtonsMap) GetKeyboard() Rows {
	rows := make(Rows, 0)
	rowNumMap := make(map[int]tele.Row)

	for _, stateBtn := range ubm {
		rowNumMap[stateBtn.row] = append(rowNumMap[stateBtn.row], *stateBtn.b)
	}

	for _, row := range rowNumMap {
		rows = append(rows, row)
	}

	return rows
}

type BaseCalendar struct {
	currYear  int
	currMonth time.Month

	opt *StateOptions
}

// StateCalendar represents the main object
type StateCalendar struct {
	Manager    *fsm.Manager
	Dispatcher fsm.Dispatcher

	opt      *StateOptions
	kb       [][]tele.InlineButton
	stateKey string

	currYear  int
	currMonth time.Month

	uniqueButtons UniqueButtonsMap

	strategies map[StrategyKey]Strategy
	strategy   Strategy
}

type Option func(options *StateOptions)

// StateOptions represents a struct for passing optional
// properties for customizing a calendar keyboard
type StateOptions struct {
	// The year that will be initially active in the calendar.
	// Default value - today's year
	InitialYear int

	// The month that will be initially active in the calendar
	// Default value - today's month
	InitialMonth time.Month

	// The range of displayed years
	// Default value - {1970, 292277026596} (time.Unix years range)
	YearRange [2]int

	// The language of all designations.
	// If equals "ru" the designations would be Russian,
	// otherwise - English
	Language string

	Seed int64

	MonthBtnHandler StateHandler
	// handler for months menu
	MonthsBtnHandler  StateHandler
	WeekDayBtnHandler StateHandler
	DayBtnHandler     StateHandler

	EmptyBtnHandler   StateHandler
	YearBtnHandler    StateHandler
	ControlBtnHandler StateHandler

	ExtraHandlers []ExtraHandler
}

func DefaultStateOptions(m *fsm.Manager, d fsm.Dispatcher) StateOptions {
	monthsBtnHandler := func(cal ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				monthNum, err := strconv.Atoi(c.Data())
				if err != nil {
					log.Fatal(err)
				}
				cal.SetCurrMonth(time.Month(monthNum))
				cal.SetStrategy(FromConfigStrategy)

				// Show the calendar keyboard with the active selected month back
				return c.Edit(c.Message().Text, cal.GetKeyboard())
			}),
		)
	}
	monthBtnHandler := func(cal ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				cal.SetStrategy(MonthStrategy)
				kb := cal.GetKeyboard()
				for _, row := range kb {
					for _, cell := range row {
						monthsBtnHandler(cal, cell)
					}
				}

				return c.Edit(c.Message().Text, cal.GetKeyboard())
			}),
		)
	}
	weekBtnHandler := func(cal ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				return ignoreQuery(c)
			}),
		)
	}
	dayBtnHandler := func(cal ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				// find button edit, add dot
				dayInt, err := strconv.Atoi(c.Data())
				if err != nil {
					return err
				}
				ubm := cal.GetButtons()
				ubm[btn.Unique].b.Text = btn.Text + "*"
				cal.SetStrategy(FromExistense)

				state.Update(context.TODO(), cal.GetStateKey(), [2]int{dayInt})

				state.SetState(context.TODO(), stateSetSecondDate)

				return c.Edit(c.Message().Text, cal.GetKeyboard())
			}),
		)
	}
	emptyBtnHandler := func(_ ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				return ignoreQuery(c)
			}),
		)
	}
	yearBtnHandler := func(cal ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				return ignoreQuery(c)
			}),
		)
	}
	controlBtnHandler := func(cal ICalender, btn tele.Btn) {
		m.Bind(
			d,
			fsmopt.On(&btn),
			fsmopt.OnStates(fsm.AnyState),
			fsmopt.Do(func(c tele.Context, state fsm.Context) error {
				if btn.Text == "<" {
					// Additional protection against entering the years ranges
					if cal.GetCurrMonth() > 1 {
						cal.SetCurrMonth(cal.GetCurrMonth() - 1)
					} else {
						cal.SetCurrMonth(12)
						if cal.GetCurrYear() > cal.GetYearRange()[0] {
							cal.SetCurrYear(cal.GetCurrYear() - 1)
						}
					}
				} else {
					// Additional protection against entering the years ranges
					if cal.GetCurrMonth() < 12 {
						cal.SetCurrMonth(cal.GetCurrMonth() + 1)
					} else {
						if cal.GetCurrYear() < cal.GetYearRange()[1] {
							cal.SetCurrYear(cal.GetCurrYear() + 1)
						}
						cal.SetCurrMonth(1)
					}

				}
				return c.Edit(c.Message().Text, cal.GetKeyboard())
			}),
		)
	}

	so := StateOptions{
		InitialYear:       time.Now().Year(),
		InitialMonth:      time.Now().Month(),
		YearRange:         [2]int{MinYearLimit, MaxYearLimit},
		Language:          "ru",
		Seed:              time.Now().UnixNano(),
		MonthBtnHandler:   monthBtnHandler,
		WeekDayBtnHandler: weekBtnHandler,
		DayBtnHandler:     dayBtnHandler,
		EmptyBtnHandler:   emptyBtnHandler,
		YearBtnHandler:    yearBtnHandler,
		ControlBtnHandler: controlBtnHandler,
	}

	return so
}

type Rows []tele.Row

// GetKeyboard builds the calendar inline-keyboard
func (cal *StateCalendar) GetKeyboard() Rows {
	// Move exRows to BaseCalendar
	var exRows Rows
	if len(cal.uniqueButtons) > 0 {
		exRows = cal.uniqueButtons.GetKeyboard()
	}

	rows := cal.strategy.GetKeyboard(exRows)
	ubm := make(UniqueButtonsMap, len(rows))
	for j, r := range rows {
		for k, c := range r {
			ubm[c.Unique] = StateBtn{
				b:   &c,
				row: j,
				idx: k,
			}
		}
	}
	cal.uniqueButtons = ubm

	monthRows := cal.strategy.GetMonthRow()
	for _, c := range monthRows {
		cal.opt.MonthBtnHandler(cal, c)
	}

	weekdaysRows := cal.strategy.GetWeekdaysRow()
	for _, c := range weekdaysRows {
		cal.opt.WeekDayBtnHandler(cal, c)
	}

	daysRows := cal.strategy.GetDays()
	for _, r := range daysRows {
		for _, c := range r {
			if c.Text == " " {
				cal.opt.EmptyBtnHandler(cal, c)
			} else {
				cal.opt.DayBtnHandler(cal, c)
			}
		}
	}

	controlRows := cal.strategy.GetControls()
	for _, c := range controlRows {
		cal.opt.ControlBtnHandler(cal, c)
	}

	return rows
}

func (cal *StateCalendar) GetStateKey() string {
	return cal.stateKey
}

func (cal *StateCalendar) GetButtons() UniqueButtonsMap {
	return cal.uniqueButtons
}

func (cal *StateCalendar) SetStrategy(strategy StrategyKey) {
	cal.strategy = cal.strategies[strategy]
}

func (cal *StateCalendar) SetCurrMonth(month time.Month) {
	cal.currMonth = month
}

func (cal *StateCalendar) SetCurrYear(y int) {
	cal.currYear = y
}

func (cal *StateCalendar) GetCurrYear() int {
	return cal.currYear
}

func (cal *StateCalendar) GetCurrMonth() time.Month {
	return cal.currMonth
}

func (cal *StateCalendar) GetYearRange() [2]int {
	return cal.opt.YearRange
}
