package tui

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/Aayush9029/netscanner-tool/internal/discovery"
	"github.com/Aayush9029/netscanner-tool/internal/scanner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phaseSelect phase = iota
	phaseManual
	phaseScan
	phaseDone
)

type scanDoneMsg struct {
	result scanner.Result
	err    error
}

type scanEventMsg scanner.Event

type model struct {
	suggestions []discovery.Suggestion
	cursor      int
	phase       phase
	input       textinput.Model
	target      string
	opts        scanner.Options

	cancel context.CancelFunc
	events <-chan scanner.Event
	done   <-chan scanDoneMsg

	started         time.Time
	completedChecks int
	totalChecks     int
	openHosts       map[string]map[int]string
	lastOpen        string
	result          scanner.Result
	err             error
	width           int
	height          int
}

var (
	screenStyle = lipgloss.NewStyle().Padding(1, 2)
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

func Run(suggestions []discovery.Suggestion, opts scanner.Options) error {
	input := textinput.New()
	input.Placeholder = "192.168.1.0/24, router.local, 10.0.0.5"
	input.CharLimit = 256
	input.Width = 56

	m := model{
		suggestions: suggestions,
		input:       input,
		opts:        opts,
		openHosts:   map[string]map[int]string{},
	}
	_, err := tea.NewProgram(m).Run()
	return err
}

func RunScan(target string, opts scanner.Options) error {
	m := model{
		phase:     phaseScan,
		target:    target,
		opts:      opts,
		openHosts: map[string]map[int]string{},
	}
	program := tea.NewProgram(m)
	_, err := program.Run()
	return err
}

func (m model) Init() tea.Cmd {
	if m.phase == phaseScan {
		return m.startScan()
	}
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
	}

	switch m.phase {
	case phaseSelect:
		return m.updateSelect(msg)
	case phaseManual:
		return m.updateManual(msg)
	case phaseScan:
		return m.updateScan(msg)
	case phaseDone:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter", "q":
				return m, tea.Quit
			}
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.suggestions)-1 {
			m.cursor++
		}
	case "enter":
		selected := m.suggestions[m.cursor]
		if selected.Kind == "manual" {
			m.phase = phaseManual
			m.input.Focus()
			return m, textinput.Blink
		}
		m.target = selected.Target
		m.phase = phaseScan
		return m, m.startScan()
	}
	return m, nil
}

func (m model) updateManual(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			target := strings.TrimSpace(m.input.Value())
			if target == "" {
				m.err = fmt.Errorf("target cannot be empty")
				return m, nil
			}
			m.err = nil
			m.target = target
			m.phase = phaseScan
			return m, m.startScan()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateScan(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanEventMsg:
		event := scanner.Event(msg)
		if event.TotalChecks > 0 {
			m.totalChecks = event.TotalChecks
		}
		if event.CompletedChecks > m.completedChecks {
			m.completedChecks = event.CompletedChecks
		}
		if event.Open {
			if _, ok := m.openHosts[event.Host]; !ok {
				m.openHosts[event.Host] = map[int]string{}
			}
			m.openHosts[event.Host][event.Port] = event.Service
			m.lastOpen = fmt.Sprintf("%s:%d %s", event.Host, event.Port, event.Service)
		}
		return m, waitForScan(m.events, m.done)
	case scanDoneMsg:
		m.result = msg.result
		m.err = msg.err
		m.phase = phaseDone
		return m, nil
	}
	return m, nil
}

func (m *model) startScan() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.started = time.Now()
	m.completedChecks = 0
	m.totalChecks = 0
	m.openHosts = map[string]map[int]string{}
	m.lastOpen = ""

	events := make(chan scanner.Event, 256)
	done := make(chan scanDoneMsg, 1)
	m.events = events
	m.done = done

	go func() {
		result, err := scanner.Scan(ctx, m.target, m.opts, events)
		close(events)
		done <- scanDoneMsg{result: result, err: err}
		close(done)
	}()

	return waitForScan(events, done)
}

func waitForScan(events <-chan scanner.Event, done <-chan scanDoneMsg) tea.Cmd {
	return func() tea.Msg {
		if event, ok := <-events; ok {
			return scanEventMsg(event)
		}
		if doneMsg, ok := <-done; ok {
			return doneMsg
		}
		return scanDoneMsg{}
	}
}

func (m model) View() string {
	switch m.phase {
	case phaseManual:
		return screenStyle.Render(m.viewManual())
	case phaseScan:
		return screenStyle.Render(m.viewScan())
	case phaseDone:
		return screenStyle.Render(m.viewDone())
	default:
		return screenStyle.Render(m.viewSelect())
	}
}

func (m model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("netscanner"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("choose a suggested target"))
	b.WriteString("\n\n")

	for i, suggestion := range m.suggestions {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = "> "
			style = accentStyle.Copy().Bold(true)
		}
		line := fmt.Sprintf("%s%s", cursor, suggestion.Label)
		if suggestion.Target != "" && suggestion.Label != suggestion.Target {
			line += mutedStyle.Render("  " + suggestion.Target)
		}
		b.WriteString(style.Render(line))
		if suggestion.Detail != "" {
			b.WriteString("\n")
			b.WriteString("  " + mutedStyle.Render(suggestion.Detail))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("enter scan  •  j/k move  •  esc quit"))
	return b.String()
}

func (m model) viewManual() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("manual target"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("enter an IP, hostname, CIDR, or comma-separated list"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("enter scan  •  esc quit"))
	return b.String()
}

func (m model) viewScan() string {
	var b strings.Builder
	elapsed := time.Since(m.started).Round(time.Second)
	percent := 0.0
	if m.totalChecks > 0 {
		percent = float64(m.completedChecks) / float64(m.totalChecks) * 100
	}

	b.WriteString(titleStyle.Render("scanning " + m.target))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d/%d checks  %.0f%%  %s", m.completedChecks, m.totalChecks, percent, elapsed)))
	b.WriteString("\n\n")
	b.WriteString(progressBar(percent, 42))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("open hosts: %s\n", accentStyle.Render(fmt.Sprintf("%d", len(m.openHosts)))))
	if m.lastOpen != "" {
		b.WriteString(fmt.Sprintf("latest: %s\n", accentStyle.Render(m.lastOpen)))
	}
	b.WriteString("\n")
	b.WriteString(renderOpenHosts(m.openHosts, 8))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("ctrl+c cancel"))
	return b.String()
}

func (m model) viewDone() string {
	var b strings.Builder
	if m.err != nil && len(m.result.Hosts) == 0 {
		b.WriteString(errorStyle.Render(m.err.Error()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(warnStyle.Render(m.err.Error()))
		b.WriteString("\n\n")
	}
	b.WriteString(titleStyle.Render("scan complete"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d hosts scanned  %d open hosts  %d open ports  %dms",
		m.result.HostsScanned,
		m.result.OpenHostCount,
		m.result.OpenPortCount,
		m.result.DurationMS,
	)))
	b.WriteString("\n\n")

	if len(m.result.Hosts) == 0 {
		b.WriteString(warnStyle.Render("No open ports found."))
	} else {
		b.WriteString(renderResultTable(m.result))
	}

	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("enter/q close"))
	return b.String()
}

func progressBar(percent float64, width int) string {
	if width < 10 {
		width = 10
	}
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return accentStyle.Render(strings.Repeat("=", filled)) + mutedStyle.Render(strings.Repeat("-", width-filled))
}

func renderOpenHosts(hosts map[string]map[int]string, limit int) string {
	if len(hosts) == 0 {
		return mutedStyle.Render("waiting for open ports")
	}
	lines := sortedOpenHostLines(hosts)
	if len(lines) > limit {
		lines = append(lines[:limit], mutedStyle.Render(fmt.Sprintf("... %d more", len(lines)-limit)))
	}
	return strings.Join(lines, "\n")
}

func sortedOpenHostLines(hosts map[string]map[int]string) []string {
	keys := make([]string, 0, len(hosts))
	for host := range hosts {
		keys = append(keys, host)
	}
	sortHosts(keys)

	lines := make([]string, 0, len(keys))
	for _, host := range keys {
		var ports []int
		for port := range hosts[host] {
			ports = append(ports, port)
		}
		sortInts(ports)
		var values []string
		for _, port := range ports {
			values = append(values, fmt.Sprintf("%d/%s", port, hosts[host][port]))
		}
		lines = append(lines, fmt.Sprintf("%s  %s", accentStyle.Render(host), strings.Join(values, ", ")))
	}
	return lines
}

func sortHosts(hosts []string) {
	sort.Slice(hosts, func(i, j int) bool {
		ipA := net.ParseIP(hosts[i]).To4()
		ipB := net.ParseIP(hosts[j]).To4()
		if ipA == nil || ipB == nil {
			return hosts[i] < hosts[j]
		}
		for index := 0; index < net.IPv4len; index++ {
			if ipA[index] == ipB[index] {
				continue
			}
			return ipA[index] < ipB[index]
		}
		return false
	})
}

func sortInts(values []int) {
	sort.Ints(values)
}

func renderResultTable(result scanner.Result) string {
	var b strings.Builder
	b.WriteString(mutedStyle.Render(pad("HOST", 18) + pad("NAME", 26) + "OPEN PORTS"))
	b.WriteString("\n")
	for _, host := range result.Hosts {
		var ports []string
		for _, port := range host.OpenPorts {
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Service))
		}
		name := host.Hostname
		if name == "" {
			name = "-"
		}
		b.WriteString(pad(host.Host, 18))
		b.WriteString(pad(name, 26))
		b.WriteString(strings.Join(ports, ", "))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func pad(value string, width int) string {
	if len(value) >= width {
		return value[:width-1] + " "
	}
	return value + strings.Repeat(" ", width-len(value))
}
