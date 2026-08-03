package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Calculator struct {
	locale Locale
}

type CalculationInput struct {
	StartTime     string
	WorkTime      string
	Worked        string
	Plan          string
	AddTZ         bool   // конвертировать результат в целевой TZ
	InputTimezone string // зона ввода (из конфига, напр. "Europe/Moscow")
	Timezone      string // целевая IANA timezone
	Breaks        []BreakTime
}

type BreakTime struct {
	From string
	To   string
}

func NewCalculator(locale Locale) *Calculator {
	return &Calculator{locale: locale}
}

func (c *Calculator) Calculate(input CalculationInput) (string, error) {
	startTime, err := c.parseTime(input.StartTime)
	if err != nil {
		return "", fmt.Errorf("%s", c.locale.ErrInvalidStartTime)
	}

	remainingTime, err := c.calculateRemainingTime(input)
	if err != nil {
		return "", err
	}

	totalBreakDuration, err := c.calculateBreaksDuration(input.Breaks, startTime, startTime.Add(remainingTime))
	if err != nil {
		return "", err
	}

	endTime := startTime.Add(remainingTime).Add(totalBreakDuration)

	result, err := c.formatResult(endTime, input.AddTZ, input.InputTimezone, input.Timezone)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (c *Calculator) calculateRemainingTime(input CalculationInput) (time.Duration, error) {
	if input.WorkTime != "" {
		return c.ParseDuration(input.WorkTime)
	}

	if input.Worked != "" && input.Plan != "" {
		worked, err := c.ParseDuration(input.Worked)
		if err != nil {
			return 0, fmt.Errorf("%s", c.locale.ErrInvalidWorked)
		}

		plan, err := c.ParseDuration(input.Plan)
		if err != nil {
			return 0, fmt.Errorf("%s", c.locale.ErrInvalidPlan)
		}

		if worked > plan {
			return 0, fmt.Errorf(c.locale.ErrWorkedExceedsPlan, input.Worked, input.Plan)
		}

		return plan - worked, nil
	}

	return 0, fmt.Errorf("%s", c.locale.ErrNoTimeInput)
}

func (c *Calculator) calculateBreaksDuration(breaks []BreakTime, workStart, _ time.Time) (time.Duration, error) {
	total := time.Duration(0)

	for i, br := range breaks {
		if br.From == "" || br.To == "" {
			continue
		}

		from, err := c.parseTime(br.From)
		if err != nil {
			return 0, fmt.Errorf(c.locale.ErrBreakInvalidFrom, i+1)
		}

		to, err := c.parseTime(br.To)
		if err != nil {
			return 0, fmt.Errorf(c.locale.ErrBreakInvalidTo, i+1)
		}

		if to.Before(from) || to.Equal(from) {
			return 0, fmt.Errorf(c.locale.ErrBreakEndBeforeStart, i+1)
		}

		if !workStart.IsZero() && from.Before(workStart) {
			return 0, fmt.Errorf(c.locale.ErrBreakBeforeWorkday, i+1, br.From, workStart.Format("15:04"))
		}

		total += to.Sub(from)
	}

	return total, nil
}

// BreaksDuration вычисляет суммарную длительность перерывов.
// СInvalid перерывы (пустые или неверный формат) пропускаются.
func (c *Calculator) BreaksDuration(breaks []BreakTime) time.Duration {
	total := time.Duration(0)
	for _, br := range breaks {
		if br.From == "" || br.To == "" {
			continue
		}
		from, err := c.parseTime(br.From)
		if err != nil {
			continue
		}
		to, err := c.parseTime(br.To)
		if err != nil {
			continue
		}
		if to.Before(from) || to.Equal(from) {
			continue
		}
		total += to.Sub(from)
	}
	return total
}

func (c *Calculator) formatResult(endTime time.Time, addTZ bool, inputTimezone, timezone string) (string, error) {
	if !addTZ {
		return endTime.Format("15:04"), nil
	}

	if timezone == "" {
		return "", fmt.Errorf("%s", c.locale.ErrTimezoneNotSet)
	}

	targetLoc := TimezoneLocation(timezone)
	if targetLoc == nil {
		return "", fmt.Errorf(c.locale.ErrTimezoneInvalid, timezone)
	}

	// Определяем зону ввода — если не задана или невалидна, используем UTC
	inputLoc := TimezoneLocation(inputTimezone)
	if inputLoc == nil {
		inputLoc = time.UTC
	}

	// Строим endTime в зоне ввода с реальной датой, затем конвертируем
	now := time.Now()
	localEnd := time.Date(
		now.Year(), now.Month(), now.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0,
		inputLoc,
	)

	converted := localEnd.In(targetLoc)
	label, _ := converted.Zone()

	return converted.Format("15:04") + " " + label, nil
}

func (c *Calculator) parseTime(value string) (time.Time, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("%s", c.locale.ErrInvalidTimeFormat)
	}

	hours, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hours < 0 || hours > 23 {
		return time.Time{}, fmt.Errorf("%s", c.locale.ErrInvalidHours)
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minutes < 0 || minutes > 59 {
		return time.Time{}, fmt.Errorf("%s", c.locale.ErrInvalidMinutes)
	}

	return time.Date(0, 1, 1, hours, minutes, 0, 0, time.UTC), nil
}

func (c *Calculator) ParseDuration(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("%s", c.locale.ErrInvalidDuration)
	}

	hours, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hours < 0 {
		return 0, fmt.Errorf("%s", c.locale.ErrInvalidHours)
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("%s", c.locale.ErrInvalidMinutes)
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, nil
}
