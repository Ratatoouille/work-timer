package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Calculator struct{}

type CalculationInput struct {
	StartTime string
	WorkTime  string
	Worked    string
	Plan      string
	AddFour   bool
	Breaks    []BreakTime
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

	totalBreakDuration, err := c.calculateBreaksDuration(input.Breaks)
	if err != nil {
		return "", err
	}

	endTime := startTime.Add(remainingTime).Add(totalBreakDuration)

	return c.formatResult(endTime, input.AddFour), nil
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

		remaining := plan - worked
		if remaining < 0 {
			remaining = 0
		}
		return remaining, nil
	}

	return 0, fmt.Errorf("введите либо оставшееся время, либо отработано/план")
}

func (c *Calculator) calculateBreaksDuration(breaks []BreakTime) (time.Duration, error) {
	total := time.Duration(0)

	for _, br := range breaks {
		if br.From == "" || br.To == "" {
			continue
		}

		from, err := parseTime(br.From)
		if err != nil {
			return 0, fmt.Errorf("неверный формат времени перерыва (начало)")
		}

		to, err := parseTime(br.To)
		if err != nil {
			return 0, fmt.Errorf("неверный формат времени перерыва (конец)")
		}

		duration := to.Sub(from)
		if duration < 0 {
			return 0, fmt.Errorf("время окончания перерыва раньше начала")
		}

		total += duration
	}

	return total, nil
}

func (c *Calculator) formatResult(endTime time.Time, addFour bool) string {
	if addFour {
		endTime = endTime.Add(4 * time.Hour)

		return endTime.Format("15:04") + " KRSK"
	}

	return endTime.Format("15:04")
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
