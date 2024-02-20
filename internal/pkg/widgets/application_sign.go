package widgets

import "strconv"

type ApplicationSignWidget struct {
	name string
	year int16
}

func (widget *ApplicationSignWidget) Render() RenderResult {
	return RenderResult(Icons.CopyRight + " " + widget.name + ", " + strconv.Itoa(int(widget.year)))
}

func ApplicationSign(name string, year int16) Widget {
	return &ApplicationSignWidget{name, year}
}
