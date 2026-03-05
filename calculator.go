package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Calculator struct{}

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

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(input CalculationInput) (string, error) {
	startTime, err := parseTime(input.StartTime)
	if err != nil {
		return "", fmt.Errorf("неверное время начала")
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
		return parseDuration(input.WorkTime)
	}

	if input.Worked != "" && input.Plan != "" {
		worked, err := parseDuration(input.Worked)
		if err != nil {
			return 0, fmt.Errorf("неверный формат 'отработано'")
		}

		plan, err := parseDuration(input.Plan)
		if err != nil {
			return 0, fmt.Errorf("неверный формат 'план'")
		}

		if worked > plan {
			return 0, fmt.Errorf("отработано (%s) больше плана (%s)", input.Worked, input.Plan)
		}

		return plan - worked, nil
	}

	return 0, fmt.Errorf("введите либо оставшееся время, либо отработано/план")
}

func (c *Calculator) calculateBreaksDuration(breaks []BreakTime, workStart, _ time.Time) (time.Duration, error) {
	total := time.Duration(0)

	for i, br := range breaks {
		if br.From == "" || br.To == "" {
			continue
		}

		from, err := parseTime(br.From)
		if err != nil {
			return 0, fmt.Errorf("перерыв %d: неверный формат времени начала", i+1)
		}

		to, err := parseTime(br.To)
		if err != nil {
			return 0, fmt.Errorf("перерыв %d: неверный формат времени конца", i+1)
		}

		if to.Before(from) || to.Equal(from) {
			return 0, fmt.Errorf("перерыв %d: время конца раньше или равно началу", i+1)
		}

		if !workStart.IsZero() && from.Before(workStart) {
			return 0, fmt.Errorf("перерыв %d: начало (%s) раньше начала рабочего дня (%s)",
				i+1, br.From, workStart.Format("15:04"))
		}

		total += to.Sub(from)
	}

	return total, nil
}

func (c *Calculator) formatResult(endTime time.Time, addTZ bool, inputTimezone, timezone string) (string, error) {
	if !addTZ {
		return endTime.Format("15:04"), nil
	}

	if timezone == "" {
		return "", fmt.Errorf("timezone не задан в конфиге (~/.config/work_timer/config.toml)")
	}

	targetLoc := TimezoneLocation(timezone)
	if targetLoc == nil {
		return "", fmt.Errorf("неверный timezone в конфиге: %q", timezone)
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

func parseTime(value string) (time.Time, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("неверный формат времени")
	}

	hours, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hours < 0 || hours > 23 {
		return time.Time{}, fmt.Errorf("неверные часы")
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minutes < 0 || minutes > 59 {
		return time.Time{}, fmt.Errorf("неверные минуты")
	}

	return time.Date(0, 1, 1, hours, minutes, 0, 0, time.UTC), nil
}

func parseDuration(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("неверный формат длительности")
	}

	hours, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hours < 0 {
		return 0, fmt.Errorf("неверные часы")
	}

	minutes, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("неверные минуты")
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, nil
}
