package theme

import "fyne.io/fyne/v2"

type CustomTheme struct {
	fyne.Theme
}

func (m CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		return resourceSourceHanSansCNBoldOtf
	}
	return resourceSourceHanSansCNRegularOtf
}
