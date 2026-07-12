// internal/ui/overlay/result.go
package overlay

import "github.com/EnJulian/shadowbox/internal/ui/style"

// Result is the task-completion toast, dismissed by any key.
type Result struct {
	Summary string
	Err     error
}

func (r Result) View(st style.Styles) string {
	var body string
	if r.Err != nil {
		body = st.Danger.Render("x "+r.Summary) + "\n" + st.Item.Render(r.Err.Error())
	} else {
		body = st.Selected.Render("+ " + r.Summary)
	}
	return body + "\n\n" + st.Help.Render("press any key to dismiss")
}
