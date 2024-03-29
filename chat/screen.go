package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iteration-A/hanekawa/constants"
	"github.com/iteration-A/hanekawa/headings"
	"github.com/iteration-A/hanekawa/statusbar"
	"github.com/iteration-A/hanekawa/websockets"
)

type Model struct {
	content     string
	ready       bool
	viewport    viewport.Model
	input       textinput.Model
	typing      bool
	firstLetter bool
	chatName    string
	username    string
}

func initialModel() Model {
	i := textinput.New()
	i.CharLimit = 80
	i.Width = constants.TermWidth / 2
	i.Prompt = ""
	i.Placeholder = "Type something..."
	i.PlaceholderStyle = placeholder
	i.SetCursorMode(textinput.CursorStatic)

	return Model{
		input: i,
	}
}

func New() Model {
	return initialModel()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		height := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-height)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - height
		}

	case constants.RoomSelectedMsg:
		m.chatName = string(msg)
		return m, getLastMessagesCmd(m.chatName)

	case MessagesMsg:
		messages := make([]string, len(msg))
		for index, message := range msg {
			messages[index] = formatMessage(message)
		}
		m.content = fmt.Sprintf("%s\n%s", joinedFormat(m.username), joinMessages(messages))
		m.viewport.SetContent(m.content)
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case constants.TokenMsg:
		m.username = string(msg.Username)

	case websockets.UserJoinedMsg:
		m.content = fmt.Sprintf("%s\n%s", m.content, joinedFormat(msg.Username))
		m.viewport.SetContent(m.content)
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case websockets.UserLeftMsg:
		m.content = fmt.Sprintf("%s\n%s", m.content, leftFormat(msg.Username))
		m.viewport.SetContent(m.content)
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case websockets.NewMessageMsg:
		m.content = fmt.Sprintf("%s\n%s", m.content, newMessageFormat(msg.From, msg.Content))
		m.viewport.SetContent(m.content)
		m.viewport.SetYOffset(m.calcExcess() + lipgloss.Height(m.content))
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "h", "H":
			if !m.typing {
				return m, goToRoomsCmd
			}

		case "i":
			if !m.typing {
				m.typing = true
				m.input.Focus()
				m.firstLetter = true
			} else {
				m.firstLetter = false
			}

		case "g":
			if m.typing {
				break
			}
			m.viewport.YOffset = 0

		case "G":
			if m.typing {
				break
			}
			m.viewport.YOffset = 0
			m.viewport.YOffset = m.calcExcess()

		case "esc":
			m.typing = false
			m.input.Blur()

		case "enter":
			websockets.ChatroomChanIn <- websockets.SendMessage{
				Room:    m.chatName,
				Content: m.input.Value(),
			}
			m.input.SetValue("")

		default:
			m.firstLetter = false
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	if m.typing {
		if !m.firstLetter {
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m Model) headerView() string {
	return headings.Title(m.chatName)
}
func (m Model) footerView() string {
	var msg string
	if m.typing {
		msg = "INSERT (esc)"
	} else {
		msg = "j↓ k↑ i(type)"
	}

	return statusbar.StatusLine(msg, m.input.View(), "Hanekawa🍙")
}

func (m Model) calcExcess() int {
	h := lipgloss.Height
	totalHeight := constants.TermHeight - h(m.headerView()) - h(m.footerView())
	excess := h(m.content) - totalHeight

	return excess
}

func formatMessage(msg message) string {
	return fmt.Sprintf("[%s] %s", msg.User.Username, msg.Content)
}

func newMessageFormat(username, content string) string {
	return fmt.Sprintf("[%s] %s", username, content)
}

func joinedFormat(username string) string {
	return joinedMessage.Render(fmt.Sprintf("%v just joined!", username))
}

func leftFormat(username string) string {
	return joinedMessage.Render(fmt.Sprintf("%v just left!", username))
}

func joinMessages(messages []string) string {
	return strings.Join(messages, "\n")
}
