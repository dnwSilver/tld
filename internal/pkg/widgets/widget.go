package widgets

type RenderResult string

type Widget interface {
	Render() RenderResult
}

const lastSymbolLen = 1
const lastSymbolVal string = "#"
const utfLen = 5

func (r RenderResult) Crop(length int) RenderResult {
	runes := []rune(r)
	if length >= len(runes) {
		return r
	}
	substr := string(runes[0 : length-lastSymbolLen])
	return RenderResult(substr + lastSymbolVal)
}
