package template

import "time"

func hour() int {
	return time.Now().Hour()
}

func minute() int {
	return time.Now().Minute()
}

func second() int {
	return time.Now().Second()
}

func month() int {
	return int(time.Now().Month())
}

func year() int {
	return time.Now().Year()
}

func day() int {
	return time.Now().Day()
}

func yearday() int {
	return time.Now().YearDay()
}

func weekday() int {
	// 星期天返回0，手动改成7
	t := time.Now().Weekday()
	if t == 0 {
		t = 7
	}
	return int(t)
}
