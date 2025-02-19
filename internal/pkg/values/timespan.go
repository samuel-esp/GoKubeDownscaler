package values

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/caas-team/gokubedownscaler/internal/pkg/util"
)

var (
	errInvalidWeekday          = errors.New("error: specified weekday is invalid")
	errRelativeTimespanInvalid = errors.New("error: specified relative timespan is invalid")
	errTimeOfDayOutOfRange     = errors.New("error: the time of day has fields that are out of rane")
)

// rfc339Regex is a regex that matches an rfc339 timestamp.
const rfc3339Regex = `(.+Z|.+[+-]\d{2}:\d{2})`

// absoluteTimeSpanRegex matches a absolute timespan. It's groups are the two rfc3339 timestamps.
var absoluteTimeSpanRegex = regexp.MustCompile(fmt.Sprintf(`^%s *- *%s$`, rfc3339Regex, rfc3339Regex))

type TimeSpan interface {
	// isTimeInSpan checks if time is in the timespan or not
	isTimeInSpan(time time.Time) bool
}

type timeSpans []TimeSpan

// inTimeSpans checks if current time is in one of the timespans or not.
func (t *timeSpans) inTimeSpans() bool {
	for _, timespan := range *t {
		if !timespan.isTimeInSpan(time.Now()) {
			continue
		}

		return true
	}

	return false
}

// String implementation for timeSpans.
func (t *timeSpans) String() string {
	if *t != nil {
		return util.UndefinedString
	}

	return fmt.Sprint(*t)
}

func (t *timeSpans) Set(value string) error {
	spans := strings.Split(value, ",")
	timespans := make([]TimeSpan, 0, len(spans))

	for _, timespanText := range spans {
		timespanText = strings.TrimSpace(timespanText)

		if isAbsoluteTimestamp(timespanText) {
			// parse as absolute timestamp
			timespan, err := parseAbsoluteTimeSpan(timespanText)
			if err != nil {
				return fmt.Errorf("failed to parse absolute timespan: %w", err)
			}

			timespans = append(timespans, timespan)

			continue
		}

		// parse as relative timestamp
		timespan, err := parseRelativeTimeSpan(timespanText)
		if err != nil {
			return fmt.Errorf("failed to parse relative timespan: %w", err)
		}

		timespans = append(timespans, timespan)
	}

	*t = timeSpans(timespans)

	return nil
}

// parseAbsoluteTimespans parses an absolute timespan. will panic if timespan is not an absolute timespan.
func parseAbsoluteTimeSpan(timespan string) (absoluteTimeSpan, error) {
	timestamps := absoluteTimeSpanRegex.FindStringSubmatch(timespan)[1:]

	fromTime, err := time.Parse(time.RFC3339, strings.TrimSpace(timestamps[0]))
	if err != nil {
		return absoluteTimeSpan{}, fmt.Errorf("failed to parse rfc3339 timestamp: %w", err)
	}

	toTime, err := time.Parse(time.RFC3339, strings.TrimSpace(timestamps[1]))
	if err != nil {
		return absoluteTimeSpan{}, fmt.Errorf("failed to parse rfc3339 timestamp: %w", err)
	}

	return absoluteTimeSpan{
		from: fromTime,
		to:   toTime,
	}, nil
}

func parseRelativeTimeSpan(timespanString string) (*relativeTimeSpan, error) {
	var err error
	timespan := relativeTimeSpan{}

	parts := strings.Split(timespanString, " ")
	if len(parts) != 3 {
		return nil, errRelativeTimespanInvalid
	}

	weekdaySpan := strings.Split(parts[0], "-")
	if len(weekdaySpan) != 2 {
		return nil, errRelativeTimespanInvalid
	}

	timeSpan := strings.Split(parts[1], "-")
	if len(timeSpan) != 2 {
		return nil, errRelativeTimespanInvalid
	}

	timezone := parts[2]

	timespan.timezone, err = time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timezone: %w", err)
	}

	timespan.timeFrom, err = parseDayTime(timeSpan[0], timespan.timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time of day from: %w", err)
	}

	timespan.timeTo, err = parseDayTime(timeSpan[1], timespan.timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time of day to: %w", err)
	}

	timespan.weekdayFrom, err = getWeekday(weekdaySpan[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'weekdayFrom': %w", err)
	}

	timespan.weekdayTo, err = getWeekday(weekdaySpan[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'weekdayTo': %w", err)
	}

	return &timespan, nil
}

type relativeTimeSpan struct {
	timezone    *time.Location
	weekdayFrom time.Weekday
	weekdayTo   time.Weekday
	timeFrom    time.Time
	timeTo      time.Time
}

// isWeekdayInRange checks if the weekday falls into the weekday range.
func (t relativeTimeSpan) isWeekdayInRange(weekday time.Weekday) bool {
	if t.weekdayFrom <= t.weekdayTo { // check if range wraps across weeks
		return weekday >= t.weekdayFrom && weekday <= t.weekdayTo
	}

	return weekday >= t.weekdayFrom || weekday <= t.weekdayTo
}

// isTimeOfDayInRange checks if the time falls into the time of day range.
func (t relativeTimeSpan) isTimeOfDayInRange(timeOfDay time.Time) bool {
	if t.timeFrom.After(t.timeTo) { // check if range wraps across days
		return timeOfDay.After(t.timeFrom) || timeOfDay.Equal(t.timeFrom) || timeOfDay.Before(t.timeTo)
	}

	return (t.timeFrom.Before(timeOfDay) || t.timeFrom.Equal(timeOfDay)) && t.timeTo.After(timeOfDay)
}

// isTimeInSpan check if the time is in the span.
func (t relativeTimeSpan) isTimeInSpan(targetTime time.Time) bool {
	targetTime = targetTime.In(t.timezone)
	timeOfDay := getTimeOfDay(targetTime)
	weekday := targetTime.Weekday()

	return t.isTimeOfDayInRange(timeOfDay) && t.isWeekdayInRange(weekday)
}

// String implementation for relativeTimeSpan.
func (t relativeTimeSpan) String() string {
	return fmt.Sprintf(
		"relativeTimeSpan(%.3s-%.3s %s-%s %s)",
		t.weekdayFrom,
		t.weekdayTo,
		t.timeFrom.Format("15:04"),
		t.timeTo.Format("15:04"),
		t.timezone,
	)
}

type absoluteTimeSpan struct {
	from time.Time
	to   time.Time
}

// isTimeInSpan check if the time is in the span.
func (t absoluteTimeSpan) isTimeInSpan(targetTime time.Time) bool {
	return (t.from.Before(targetTime) || t.from.Equal(targetTime)) && t.to.After(targetTime)
}

// String implementation for absoluteTimeSpan.
func (t absoluteTimeSpan) String() string {
	return fmt.Sprintf(
		"absoluteTimeSpan(%s - %s)",
		t.from.Format(time.RFC3339),
		t.to.Format(time.RFC3339),
	)
}

// isAbsoluteTimestamp checks if timestamp string is absolute.
func isAbsoluteTimestamp(timestamp string) bool {
	return absoluteTimeSpanRegex.MatchString(timestamp)
}

// getWeekday gets the weekday from the given string.
func getWeekday(weekday string) (time.Weekday, error) {
	weekdays := map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}

	if day, ok := weekdays[strings.ToLower(weekday)]; ok {
		return day, nil
	}

	return 0, errInvalidWeekday
}

// parseDayTime parses the given time of day string to a zero date time.
func parseDayTime(daytime string, timezone *time.Location) (time.Time, error) {
	parts := strings.Split(daytime, ":")

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse hour of daytime: %w", err)
	}

	if hour < 0 || hour > 24 {
		return time.Time{}, errTimeOfDayOutOfRange
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse minute of daytime: %w", err)
	}

	if minute < 0 || minute >= 60 {
		return time.Time{}, errTimeOfDayOutOfRange
	}

	return time.Date(0, time.January, 1, hour, minute, 0, 0, timezone), nil
}

// getTimeOfDay gets the time of day of the given time.
func getTimeOfDay(targetTime time.Time) time.Time {
	return time.Date(0, time.January, 1,
		targetTime.Hour(),
		targetTime.Minute(),
		targetTime.Second(),
		targetTime.Nanosecond(),
		targetTime.Location(),
	)
}
