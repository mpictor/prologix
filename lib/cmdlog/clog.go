package cmdlog

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gotmc/prologix"
)

func isAscii(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		switch {
		case r < 7:
			return true
		case r > 6 && r < 14:
			return false
		case r > 13 && r < 32:
			return true
		case r > 127:
			return true
		}
		return false
	})
}

var (
	CmdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	R1Style  = lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	R2Style  = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
)

func PrettyFuncs(gpib *prologix.Controller) (
	query func(string) string,
	bquery func(string),
	cmd func(string),
) {
	query = func(q string) string {
		s, err := gpib.Query(q)
		if err != nil {
			q = CmdStyle.Render(q)
			log.Printf("query %q: error %s", q, err)
		}
		return s
	}
	bquery = func(q string) {
		a := query(q)
		q = CmdStyle.Render(q)

		a = strings.TrimSuffix(a, "\n") //appended by ar488
		if len(a) == 1 && a[0] == 0xff {
			// 7912 replies with 0xff if a response is expected
			// but the last command has no result
			a = ""
		}
		if len(a) == 0 {
			log.Print(R1Style.Render("<no response>"))
			return
		}

		if isAscii(a) {
			log.Printf("%s: [%d] %q", q, len(a), a)
		} else if len(a) < 32 {
			log.Printf("%s: [%d] %q (% 2x)", q, len(a), a, []byte(a))
		} else {
			log.Printf("%s: [%d] % 2x", q, len(a), []byte(a))
		}
	}

	cmd = func(c string) {
		if err := gpib.Command(c); err != nil {
			log.Printf("cmd %s: error %s", CmdStyle.Render(c), err)
		} else {
			log.Printf("%s()", CmdStyle.Render(c))
		}
	}
	return query, bquery, cmd
}
