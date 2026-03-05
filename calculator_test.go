package main

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid time with leading zeros",
			input:   "09:30",
			want:    "09:30",
			wantErr: false,
		},
		{
			name:    "valid time without leading zeros",
			input:   "9:30",
			want:    "09:30",
			wantErr: false,
		},
		{
			name:    "midnight",
			input:   "00:00",
			want:    "00:00",
			wantErr: false,
		},
		{
			name:    "end of day",
			input:   "23:59",
			want:    "23:59",
			wantErr: false,
		},
		{
			name:    "invalid hours",
			input:   "25:00",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid minutes",
			input:   "12:60",
			want:    "",
			wantErr: true,
		},
		{
			name:    "negative hours",
			input:   "-1:30",
			want:    "",
			wantErr: true,
		},
		{
			name:    "missing colon",
			input:   "0930",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "letters instead of numbers",
			input:   "ab:cd",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Format("15:04") != tt.want {
				t.Errorf("parseTime() = %v, want %v", got.Format("15:04"), tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "8 hours",
			input:   "08:00",
			want:    8 * time.Hour,
			wantErr: false,
		},
		{
			name:    "8 hours without leading zero",
			input:   "8:00",
			want:    8 * time.Hour,
			wantErr: false,
		},
		{
			name:    "30 minutes",
			input:   "00:30",
			want:    30 * time.Minute,
			wantErr: false,
		},
		{
			name:    "8.5 hours",
			input:   "08:30",
			want:    8*time.Hour + 30*time.Minute,
			wantErr: false,
		},
		{
			name:    "zero duration",
			input:   "00:00",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid minutes",
			input:   "08:60",
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative hours",
			input:   "-5:30",
			want:    0,
			wantErr: true,
		},
		{
			name:    "missing colon",
			input:   "0830",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculatorCalculate(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name    string
		input   CalculationInput
		want    string
		wantErr bool
	}{
		{
			name: "simple calculation with remaining time",
			input: CalculationInput{
				StartTime: "09:00",
				WorkTime:  "08:00",
				AddTZ:     false,
			},
			want:    "17:00",
			wantErr: false,
		},
		{
			name: "calculation with breaks",
			input: CalculationInput{
				StartTime: "09:00",
				WorkTime:  "08:00",
				Breaks: []BreakTime{
					{From: "12:00", To: "13:00"},
				},
				AddTZ: false,
			},
			want:    "18:00",
			wantErr: false,
		},
		{
			name: "calculation with multiple breaks",
			input: CalculationInput{
				StartTime: "09:00",
				WorkTime:  "08:00",
				Breaks: []BreakTime{
					{From: "12:00", To: "12:30"},
					{From: "15:00", To: "15:15"},
				},
				AddTZ: false,
			},
			want:    "17:45",
			wantErr: false,
		},
		{
			name: "calculation with timezone conversion",
			input: CalculationInput{
				StartTime:     "09:00",
				WorkTime:      "08:00",
				AddTZ:         true,
				InputTimezone: "Europe/Moscow",    // UTC+3
				Timezone:      "Asia/Krasnoyarsk", // UTC+7, diff +4h
			},
			want:    "21:00 +07",
			wantErr: false,
		},
		{
			name: "calculation with worked/plan",
			input: CalculationInput{
				StartTime: "09:00",
				Worked:    "05:00",
				Plan:      "08:00",
				AddTZ:     false,
			},
			want:    "12:00",
			wantErr: false,
		},
		{
			name: "worked exceeds plan",
			input: CalculationInput{
				StartTime: "09:00",
				Worked:    "10:00",
				Plan:      "08:00",
				AddTZ:     false,
			},
			want:    "09:00",
			wantErr: false,
		},
		{
			name: "invalid start time",
			input: CalculationInput{
				StartTime: "25:00",
				WorkTime:  "08:00",
			},
			wantErr: true,
		},
		{
			name: "missing work time and plan",
			input: CalculationInput{
				StartTime: "09:00",
			},
			wantErr: true,
		},
		{
			name: "invalid break time",
			input: CalculationInput{
				StartTime: "09:00",
				WorkTime:  "08:00",
				Breaks: []BreakTime{
					{From: "25:00", To: "13:00"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Calculate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateBreaksDuration(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name    string
		breaks  []BreakTime
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "no breaks",
			breaks:  []BreakTime{},
			want:    0,
			wantErr: false,
		},
		{
			name: "single break",
			breaks: []BreakTime{
				{From: "12:00", To: "13:00"},
			},
			want:    time.Hour,
			wantErr: false,
		},
		{
			name: "multiple breaks",
			breaks: []BreakTime{
				{From: "12:00", To: "12:30"},
				{From: "15:00", To: "15:45"},
			},
			want:    75 * time.Minute,
			wantErr: false,
		},
		{
			name: "empty break",
			breaks: []BreakTime{
				{From: "", To: ""},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "invalid break end before start",
			breaks: []BreakTime{
				{From: "13:00", To: "12:00"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.calculateBreaksDuration(tt.breaks)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateBreaksDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateBreaksDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
