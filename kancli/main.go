package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type status int

const divisor = 4

const (
	todo status = iota
	inProgress
	done
)

/* MODEL MANAGEMENT */
var models []tea.Model

const (
	model status = iota
	form
)

/* STYLING */
var (
	columnStyle = lipgloss.NewStyle().
			Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

/* CUSTOM ITEM */

type Task struct {
	status      status
	title       string
	description string
}

func NewTask(status status, title, description string) Task {
	return Task{status: status, title: title, description: description}
}

func (t *Task) Next() {
	if t.status == done {
		t.status = todo
	} else {
		t.status++
	}
}

// func (t *Task) Prev() {
// 	if t.status == todo {
// 		t.status = done
// 	} else {
// 		t.status--
// 	}
// }

// implement the list.Item interface
func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

/* MAIN MODEL */

type Model struct {
	focused  status
	lists    []list.Model
	err      error
	loaded   bool
	quitting bool
}

func New() *Model {
	return &Model{}
}

func (m *Model) Next() {
	if m.focused == done {
		m.focused = todo
	} else {
		m.focused++
	}
}

func (m *Model) Prev() {
	if m.focused == todo {
		m.focused = done
	} else {
		m.focused--
	}
}

// func (m *Model) MoveToNext() tea.Msg {
// 	selectedItem := m.lists[m.focused].SelectedItem()
// 	selectedTask := selectedItem.(Task)
// 	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
// 	selectedTask.Next()
// 	m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items())-1, list.Item(selectedTask))
// 	return nil
// }

func (m *Model) MoveToNext() tea.Msg {
	if len(m.lists) == 0 {
		return nil
	}

	selectedItem := m.lists[m.focused].SelectedItem()
	if selectedItem == nil {
		return nil
	}

	selectedTask, ok := selectedItem.(Task)
	if !ok {
		return nil
	}

	from := m.focused
	to := selectedTask.status

	m.lists[from].RemoveItem(m.lists[from].Index())

	selectedTask.Next()

	dest := &m.lists[to]
	dest.InsertItem(len(dest.Items()), selectedTask)

	return nil
}

func (m *Model) InitLists(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height/2)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}

	// Init To Do
	m.lists[todo].Title = "To Do"
	m.lists[todo].SetItems([]list.Item{
		Task{status: todo, title: "buy milk", description: "strawberry milk"},
		Task{status: todo, title: "eat sushi", description: "negitoro roll"},
		Task{status: todo, title: "fold laundry", description: "or wear wrinkly t-shirts"},
	})
	// Init in progress
	m.lists[inProgress].Title = "In Progress"
	m.lists[inProgress].SetItems([]list.Item{
		Task{status: todo, title: "write code", description: "don't worry it's Go!"},
	})
	// Init done
	m.lists[done].Title = "Done"
	m.lists[done].SetItems([]list.Item{
		Task{status: todo, title: "stay cool", description: "as a cucumber"},
	})
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			columnStyle.Width(msg.Width / divisor)
			focusedStyle.Width(msg.Width / divisor)
			columnStyle.Height(msg.Height - divisor)
			focusedStyle.Height(msg.Height - divisor)
			m.InitLists(msg.Width, msg.Height)
			m.loaded = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			m.Prev()
		case "right", "l":
			m.Next()
		// case "enter":
		// 	return m, m.MoveToNext
		case "enter":
			if selectedItem := m.lists[m.focused].SelectedItem(); selectedItem != nil {
				task, ok := selectedItem.(Task)
				if ok {
					from := m.focused
					to := task.status

					m.lists[from].RemoveItem(m.lists[from].Index())
					task.Next()
					m.lists[to].InsertItem(len(m.lists[to].Items()), task)
				}
			}
			return m, nil
		case "n":
			models[model] = m
			models[form] = NewForm(m.focused)
			return models[form].Update(nil)
		}
	case Task:
		task := msg
		return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
	}

	if !m.loaded || len(m.lists) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.loaded {
		todoView := m.lists[todo].View()
		inProgView := m.lists[inProgress].View()
		doneView := m.lists[done].View()
		switch m.focused {
		case inProgress:
			return tea.NewView(lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				focusedStyle.Render(inProgView),
				columnStyle.Render(doneView),
			))
		case done:
			return tea.NewView(lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				columnStyle.Render(inProgView),
				focusedStyle.Render(doneView),
			))
		default:
			return tea.NewView(lipgloss.JoinHorizontal(
				lipgloss.Left,
				focusedStyle.Render(todoView),
				columnStyle.Render(inProgView),
				columnStyle.Render(doneView),
			))
		}
	} else {
		return tea.NewView("loading...")
	}
}

/* FORM MODEL */
type Form struct {
	focused     status
	title       textinput.Model
	description textarea.Model
}

func NewForm(focused status) *Form {
	form := &Form{}
	form.focused = focused
	form.title = textinput.New()
	form.title.Focus()
	form.description = textarea.New()
	return form
}

func (m Form) CreateTask() tea.Msg {
	task := NewTask(m.focused, m.title.Value(), m.description.Value())
	return task
}

func (m Form) Init() tea.Cmd {
	return nil
}

func (m Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.title.Focused() {
				m.title.Blur()
				m.description.Focus()
				return m, textinput.Blink
			} else {
				models[form] = m
				return models[model], m.CreateTask
			}
		}
	}
	if m.title.Focused() {
		m.title, cmd = m.title.Update(msg)
		return m, cmd
	} else {
		m.description, cmd = m.description.Update(msg)
		return m, cmd
	}
}

func (m Form) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, m.title.View(), m.description.View()))
}

func main() {
	models = []tea.Model{New(), NewForm(todo)}
	m := models[model]
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
