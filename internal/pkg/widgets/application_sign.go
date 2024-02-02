package widgets

import "strconv"

type ApplicationSignWidget struct {
	name string
	year int16
}

func ApplicationSign(name string, year int16) Widget {
	return ApplicationSignWidget{name, year}
}

func (widget ApplicationSignWidget) Render() string {
	return `󰗦 ` + widget.name + ", " + strconv.Itoa(int(widget.year))
}
