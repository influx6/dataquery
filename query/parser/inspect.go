package parser

import (
	"fmt"
	"strconv"
	"strings"
)

//DefaultInspectionFactory provides a singleton for inspections
var DefaultInspectionFactory = NewInspectionFactory()

// AddDefaultInspections adds default inspection handlers to supplied inspectionfactory
func AddDefaultInspections(inspect *InspectionFactory) {
	inspect.Register("gt", func(data string) (Collector, error) {

		cond := NewCondition("gt")
		num, err := strconv.Atoi(data)

		if err != nil {
			return nil, err
		}

		cond.Set("value", num)

		return cond, nil
	})

	inspect.Register("gte", func(data string) (Collector, error) {

		cond := NewCondition("gte")
		num, err := strconv.Atoi(data)

		if err != nil {
			return nil, err
		}

		cond.Set("value", num)

		return cond, nil
	})

	inspect.Register("lt", func(data string) (Collector, error) {

		cond := NewCondition("lt")
		num, err := strconv.Atoi(data)

		if err != nil {
			return nil, err
		}

		cond.Set("value", num)

		return cond, nil
	})

	inspect.Register("lte", func(data string) (Collector, error) {

		cond := NewCondition("lte")
		num, err := strconv.Atoi(data)

		if err != nil {
			return nil, err
		}

		cond.Set("value", num)

		return cond, nil
	})

	inspect.Register("id", func(data string) (Collector, error) {

		cond := NewCondition("is")

		data = strings.TrimSpace(data)

		//confirm its a number
		num, err := strconv.Atoi(data)

		if err != nil {
			return nil, err
		}

		cond.Set("value", fmt.Sprintf("%d", num))

		return cond, nil
	})

	inspect.Register("in", func(data string) (Collector, error) {

		cond := NewCondition("in")

		val := strings.TrimSpace(data)
		val = strings.TrimPrefix(val, "[")
		val = strings.TrimSuffix(val, "]")

		options := strings.Split(val, " ")

		cond.Set("range", options)

		return cond, nil
	})

	inspect.Register("with", func(data string) (Collector, error) {

		cond := NewCondition("with")

		val := strings.TrimSpace(data)
		val = strings.TrimPrefix(val, "[")
		val = strings.TrimSuffix(val, "]")

		options := strings.Split(val, " ")

		cond.Set("value", options)

		return cond, nil
	})

	inspect.Register("is", func(data string) (Collector, error) {

		cond := NewCondition("is")
		cond.Set("value", strings.TrimSpace(data))

		return cond, nil
	})

	inspect.Register("isnot", func(data string) (Collector, error) {

		cond := NewCondition("isnot")
		cond.Set("value", strings.TrimSpace(data))

		return cond, nil
	})

	inspect.Register("range", func(data string) (Collector, error) {
		props := strings.Split(data, "..")

		if len(props) != 2 {
			return nil, fmt.Errorf("Invalid string %s does not match 'min..max' rule ", data)
		}

		smin, smax := strings.TrimSpace(props[0]), strings.TrimSpace(props[1])

		//log.Debug("range: Will Parse Min: %s and Max %s", smin, smax)

		min, err := strconv.Atoi(smin)

		if err != nil {
			return nil, fmt.Errorf("Invalid numeric value %d for min with error %+s", min, err)
		}

		max, err := strconv.Atoi(smax)

		if err != nil {
			return nil, fmt.Errorf("Invalid numeric value %d for max with error %+s", max, err)
		}

		cond := NewCondition("range")
		cond.Set("max", max)
		cond.Set("min", min)

		return cond, nil
	})
}

func init() {
	AddDefaultInspections(DefaultInspectionFactory)
}
