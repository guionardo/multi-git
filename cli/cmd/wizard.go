package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spf13/cobra"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/guionardo/multi-git/internal/config"
	"github.com/guionardo/multi-git/internal/detect"
	"github.com/guionardo/multi-git/internal/ssh"
)

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Guided setup wizard (interactive)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWizard() },
}

type wizStep int

const (
	stepWelcome wizStep = iota
	stepPersonal
	stepCorporate
	stepChecking
	stepResults
	stepApply
	stepDone
)

type wizardModel struct {
	step    wizStep
	detect  detect.Result
	inputs  []textinput.Model
	focus   int
	spinner spinner.Model
	checks  []string
	err     error
	done    bool
	width   int
	height  int

	personalName, personalEmail, personalKey string
	corporateName, corporateEmail, corporateKey string

	checking  bool
	applying  bool
	applyDone bool
	written   []string
}

func runWizard() error {
	d := detect.Run()
	m := newWizardModel(d)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	fm, ok := final.(wizardModel)
	if !ok {
		return nil
	}
	if fm.err != nil {
		return fm.err
	}
	return nil
}

func newWizardModel(d detect.Result) wizardModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	personalName := d.GitIdentity.Name
	if personalName == "" && d.GHLogin != "" {
		personalName = d.GHLogin
	}
	personalEmail := d.GitIdentity.Email
	if personalEmail == "" && d.GHEmail != "" {
		personalEmail = d.GHEmail
	}
	corporateName := personalName
	if d.AZLogin != "" && !strings.Contains(d.AZLogin, "@") {
		corporateName = d.AZLogin
	}
	corporateEmail := "you@company.com"
	if d.AZEmail != "" {
		corporateEmail = d.AZEmail
	}

	personalKey := "id_ed25519"
	if !keyExists(d.SSHKeys, personalKey) && len(d.SSHKeys) > 0 {
		personalKey = d.SSHKeys[0].Name
	}
	corporateKey := "id_rsa"
	if !keyExists(d.SSHKeys, corporateKey) {
		if len(d.SSHKeys) > 1 {
			corporateKey = d.SSHKeys[1].Name
		} else if len(d.SSHKeys) > 0 {
			corporateKey = d.SSHKeys[0].Name
		}
	}

	mkInput := func(val, placeholder string) textinput.Model {
		ti := textinput.New()
		ti.SetValue(val)
		ti.Placeholder = placeholder
		ti.CharLimit = 128
		ti.Width = 40
		return ti
	}

	inputs := []textinput.Model{
		mkInput(personalName, "Your Name"),
		mkInput(personalEmail, "you@example.com"),
		mkInput(personalKey, "id_ed25519"),
		mkInput(corporateName, "Your Name"),
		mkInput(corporateEmail, "you@company.com"),
		mkInput(corporateKey, "id_rsa"),
	}
	inputs[0].Focus()

	return wizardModel{
		step:          stepWelcome,
		detect:        d,
		spinner:       sp,
		inputs:        inputs,
		personalName:  personalName,
		personalEmail: personalEmail,
		personalKey:   personalKey,
		corporateName:  corporateName,
		corporateEmail: corporateEmail,
		corporateKey:   corporateKey,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step != stepDone {
				return m, tea.Quit
			}
		case "enter":
			switch m.step {
			case stepWelcome:
				m.step = stepPersonal
				m.focus = 0
				m.inputs[0].Focus()
				return m, nil
			case stepPersonal:
				if m.focus < 2 {
					m.focus++
					blurAll(m.inputs)
					m.inputs[m.focus].Focus()
					return m, nil
				}
				// save personal
				m.personalName = m.inputs[0].Value()
				m.personalEmail = m.inputs[1].Value()
				m.personalKey = m.inputs[2].Value()
				m.step = stepCorporate
				m.focus = 3
				blurAll(m.inputs)
				m.inputs[3].Focus()
				return m, nil
			case stepCorporate:
				if m.focus < 5 {
					m.focus++
					blurAll(m.inputs)
					m.inputs[m.focus].Focus()
					return m, nil
				}
				m.corporateName = m.inputs[3].Value()
				m.corporateEmail = m.inputs[4].Value()
				m.corporateKey = m.inputs[5].Value()
				m.step = stepChecking
				m.checking = true
				return m, tea.Batch(m.spinner.Tick, doChecksCmd(m))
			case stepResults:
				m.step = stepApply
				return m, nil
			case stepApply:
				m.step = stepDone
				m.applying = true
				return m, tea.Batch(m.spinner.Tick, doApplyCmd(m))
			case stepDone:
				return m, tea.Quit
			}
		case "tab", "down":
			if m.step == stepPersonal || m.step == stepCorporate {
				start, end := 0, 3
				if m.step == stepCorporate {
					start, end = 3, 6
				}
				m.focus++
				if m.focus >= end {
					m.focus = start
				}
				blurAll(m.inputs)
				m.inputs[m.focus].Focus()
				return m, nil
			}
		case "shift+tab", "up":
			if m.step == stepPersonal || m.step == stepCorporate {
				start, end := 0, 3
				if m.step == stepCorporate {
					start, end = 3, 6
				}
				m.focus--
				if m.focus < start {
					m.focus = end - 1
				}
				blurAll(m.inputs)
				m.inputs[m.focus].Focus()
				return m, nil
			}
		case "esc":
			if m.step == stepResults {
				m.step = stepCorporate
				m.focus = 3
				blurAll(m.inputs)
				m.inputs[3].Focus()
				return m, nil
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case checksDoneMsg:
		m.checking = false
		m.checks = msg.results
		m.step = stepResults
		return m, nil
	case applyDoneMsg:
		m.applying = false
		m.applyDone = true
		m.written = msg.written
		m.err = msg.err
		m.step = stepDone
		return m, nil
	}

	// update focused input
	if m.step == stepPersonal || m.step == stepCorporate {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return m, cmd
	}
	return m, nil
}

func blurAll(inputs []textinput.Model) {
	for i := range inputs {
		inputs[i].Blur()
	}
}

type checksDoneMsg struct {
	results []string
}

func doChecksCmd(m wizardModel) tea.Cmd {
	return func() tea.Msg {
		home := m.detect.Home
		var results []string
		// key pair match
		personalPriv := filepath.Join(home, ".ssh", m.personalKey)
		personalPub := personalPriv + ".pub"
		if ssh.KeyPairMatch(personalPriv, personalPub) {
			results = append(results, "✓ personal key pair matches")
		} else {
			results = append(results, "✗ personal key pair mismatch or missing")
		}
		corpPriv := filepath.Join(home, ".ssh", m.corporateKey)
		corpPub := corpPriv + ".pub"
		if ssh.KeyPairMatch(corpPriv, corpPub) {
			results = append(results, "✓ corporate key pair matches")
		} else {
			results = append(results, "✗ corporate key pair mismatch or missing")
		}
		// permissions
		if ok, msg := ssh.KeyPermissions(personalPriv); ok {
			results = append(results, "✓ personal private key permissions ok")
		} else {
			results = append(results, "✗ personal key permissions: "+msg)
		}
		if ok, msg := ssh.KeyPermissions(corpPriv); ok {
			results = append(results, "✓ corporate private key permissions ok")
		} else {
			results = append(results, "✗ corporate key permissions: "+msg)
		}
		// host checks
		ok, msg := ssh.HostCheck("github.com", personalPriv, 15*time.Second)
		if ok {
			results = append(results, "✓ github.com reachable/authenticated")
		} else {
			results = append(results, "✗ github.com: "+truncate(msg, 80))
		}
		ok, msg = ssh.HostCheck("vs-ssh.visualstudio.com", corpPriv, 15*time.Second)
		if ok {
			results = append(results, "✓ vs-ssh.visualstudio.com reachable/authenticated")
		} else {
			results = append(results, "✗ vs-ssh.visualstudio.com: "+truncate(msg, 80))
		}
		return checksDoneMsg{results: results}
	}
}

type applyDoneMsg struct {
	written []string
	err     error
}

func doApplyCmd(m wizardModel) tea.Cmd {
	return func() tea.Msg {
		data := config.Data{
			PersonalName:   m.personalName,
			PersonalEmail:  m.personalEmail,
			PersonalKey:    m.personalKey,
			CorporateName:  m.corporateName,
			CorporateEmail: m.corporateEmail,
			CorporateKey:   m.corporateKey,
		}
		home := m.detect.Home
		written, err := config.Apply(data, home, Flags.DryRun, Flags.Yes, func(label string) bool {
			return true
		})
		return applyDoneMsg{written: written, err: err}
	}
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func (m wizardModel) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("multi-git wizard")
	b.WriteString(title + "\n\n")

	switch m.step {
	case stepWelcome:
		b.WriteString("Welcome! This wizard will configure multiple Git identities.\n")
		b.WriteString("Detected home: " + m.detect.Home + "\n")
		b.WriteString(fmt.Sprintf("SSH keys found: %d\n", len(m.detect.SSHKeys)))
		b.WriteString("\nPress Enter to begin, q to quit.\n")
	case stepPersonal:
		b.WriteString("Personal identity (GitHub)\n\n")
		labels := []string{"Name", "Email", "SSH key name"}
		for i := 0; i < 3; i++ {
			b.WriteString(labels[i] + ": " + m.inputs[i].View() + "\n")
		}
		b.WriteString("\nTab/Enter to move, Enter on last field to continue.\n")
	case stepCorporate:
		b.WriteString("Corporate identity (Azure DevOps)\n\n")
		labels := []string{"Name", "Email", "SSH key name"}
		for i := 0; i < 3; i++ {
			b.WriteString(labels[i] + ": " + m.inputs[i+3].View() + "\n")
		}
		b.WriteString("\nTab/Enter to move, Enter on last field to continue.\n")
	case stepChecking:
		b.WriteString("Checking keys and hosts...\n\n")
		b.WriteString(m.spinner.View() + " working\n")
	case stepResults:
		b.WriteString("Validation results:\n\n")
		for _, c := range m.checks {
			b.WriteString(c + "\n")
		}
		b.WriteString("\nPress Enter to apply, Esc to go back.\n")
	case stepApply:
		if m.applying {
			b.WriteString("Applying configuration...\n\n")
			b.WriteString(m.spinner.View() + " writing files\n")
		}
	case stepDone:
		if m.err != nil {
			b.WriteString("Error: " + m.err.Error() + "\n")
		} else {
			b.WriteString("Configuration applied.\n\n")
			b.WriteString("Written files:\n")
			for _, w := range m.written {
				b.WriteString("  - " + w + "\n")
			}
			b.WriteString("\nNext steps (manual):\n")
			b.WriteString(tutorialText(m))
		}
		b.WriteString("\nPress Enter to exit.\n")
	}
	return b.String()
}

func tutorialText(m wizardModel) string {
	var b strings.Builder
	b.WriteString("1. Ensure your public keys are registered:\n")
	b.WriteString("   - GitHub: Settings -> SSH and GPG keys -> add " + m.personalKey + ".pub\n")
	b.WriteString("   - Azure DevOps: User settings -> SSH Public Keys -> add " + m.corporateKey + ".pub\n")
	b.WriteString("2. Move repositories under ~/personal and ~/work.\n")
	b.WriteString("3. Run 'multi-git verify' to validate.\n")
	b.WriteString("4. Test SSH: ssh -T git@github.com ; ssh -T git@vs-ssh.visualstudio.com\n")
	return b.String()
}