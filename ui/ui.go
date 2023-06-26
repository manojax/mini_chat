package ui

import (
	"bytes"
	"chat_tool/connection"
	"chat_tool/entity"
	"chat_tool/utils"
	"context"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	defaultPort        = "25042"
	defaultBroadcastIP = "224.0.0.1"
	reprintFrequency   = 50 * time.Millisecond
	timeFormat         = time.RFC3339
	maxMessagesInView  = 10000
	CHAT_PAGE          = "CHAT_PAGE"
	LOG_PAGE           = "LOG_PAGE"
)

type App struct {
	broadcastIP string
	owner       *entity.Owner
	loggerView  *LoggerView
	textInput   *TextInput
	textView    *TextView
	sidebar     *Sidebar
	view        *tview.Flex
	ui          *tview.Application
	pages       *tview.Pages
	currentRoom *entity.Room
	currentView int
}

func checkInputPort(textToCheck string, lastChar rune) bool {
	if _, err := strconv.ParseInt(textToCheck, 10, 32); err != nil {
		return false
	}
	return true
}

func NewApp(broadcastChanBuffer int) *App {
	title := ""
	yourName := ""
	localPort := defaultPort
	broadcastIP := defaultBroadcastIP

	appInfo := tview.NewApplication()
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	background := tview.NewImage()
	background.SetColors(tview.TrueColor)

	b, _ := base64.StdEncoding.DecodeString(background_image)
	photo, _ := jpeg.Decode(bytes.NewReader(b))
	background.SetImage(photo)

	form := tview.NewForm().
		AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs."}, 0, func(option string, optionIndex int) {
			title = option
		}).
		AddInputField("Your name", "", 20, nil, func(text string) {
			yourName = strings.TrimSpace(text)
		}).
		AddInputField("Broadcast IP", broadcastIP, 20, nil, func(text string) {
			if net.ParseIP(strings.TrimSpace(text)) != nil {
				broadcastIP = strings.TrimSpace(text)
			} else {
				broadcastIP = defaultBroadcastIP
			}
		}).
		AddInputField("Port", localPort, 20, checkInputPort, func(text string) {
			localPort = strings.TrimSpace(text)
			if localPort == "" {
				localPort = defaultPort
			}
		}).
		AddButton("☻ FANTASY REALM ☻", func() {
			if yourName != "" && title != "" {
				appInfo.Stop()
			}
		})
	form.SetButtonBackgroundColor(tcell.ColorRed).
		SetButtonsAlign(tview.AlignCenter)
	form.SetBorder(true)
	form.SetTitle("Information")

	pages := tview.NewPages().
		AddPage("background", background, true, true).
		AddPage("modal", modal(form, 55, 13), true, true)

	if err := appInfo.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}

	if yourName == "" || title == "" {
		panic("exits")
	}
	p := entity.NewOwner(fmt.Sprintf("%s (%s)", yourName, title), localPort, broadcastChanBuffer)
	appChat := &App{
		owner:       p,
		loggerView:  NewLoggerView(),
		textInput:   NewTextInput(),
		textView:    NewTextView(),
		sidebar:     NewSidebar(p.Repo),
		view:        tview.NewFlex(),
		ui:          tview.NewApplication(),
		pages:       tview.NewPages(),
		currentRoom: nil,
		currentView: 0,
		broadcastIP: broadcastIP,
	}
	appChat.initView()
	appChat.initBindings()
	appChat.run()
	return appChat
}

func (app *App) Run(ctx context.Context, version string) error {
	if c := connection.NewBroker(app.owner, app.broadcastIP); c != nil {
		c.Start(ctx)
	}

	go func() {
		for {
			removeid, ok := <-app.owner.Repo.Updated
			if !ok {
				return
			}
			if removeid == app.currentRoom.Id {
				app.currentRoom = nil
			}
		}
	}()

	go utils.LL.Exec(ctx, func(s string) {
		app.loggerView.RenderMessages(s)
	})

	frameChat := tview.NewFrame(app.view).
		SetBorders(1, 1, 1, 1, 2, 2).
		AddText(fmt.Sprintf("Hello: %s - %s", app.owner.Name, app.owner.Id), true, tview.AlignLeft, tcell.ColorGreen).
		AddText(fmt.Sprintf("Version: %s", version), false, tview.AlignRight, tcell.ColorRed).
		AddText("CreatedBy: Duy Nguyễn (duy.nguyen7)", false, tview.AlignLeft, tcell.ColorRed).
		AddText("⬅ Logs", false, tview.AlignCenter, tcell.ColorYellow)
	app.pages.AddPage(CHAT_PAGE, frameChat, true, true)

	frameLog := tview.NewFrame(app.loggerView.View).
		SetBorders(1, 1, 1, 1, 2, 2).
		AddText(fmt.Sprintf("Hello: %s - %s", app.owner.Name, app.owner.Id), true, tview.AlignLeft, tcell.ColorGreen).
		AddText(fmt.Sprintf("Version: %s", version), false, tview.AlignRight, tcell.ColorRed).
		AddText("CreatedBy: Duy Nguyễn (duy.nguyen7)", false, tview.AlignLeft, tcell.ColorRed).
		AddText("⮕ Chat", false, tview.AlignCenter, tcell.ColorYellow)
	app.pages.AddPage(LOG_PAGE, frameLog, true, false)

	return app.ui.SetRoot(app.pages, true).EnableMouse(true).SetFocus(app.pages).Run()
}

func (app *App) initView() {
	app.view.
		AddItem(app.sidebar.View, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(app.textView.View, 0, 3, false).
			AddItem(app.textInput.View, 5, 1, false),
			0, 2, false)
}

func (app *App) initBindings() {
	app.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft {
			app.pages.SwitchToPage(LOG_PAGE)
		}
		return event
	})
	app.loggerView.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight {
			app.pages.SwitchToPage(CHAT_PAGE)
		}
		return event
	})

	app.sidebar.View.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftDoubleClick {
			if app.sidebar.View.GetItemCount() > 0 {
				app.currentRoom = app.getCurrentRoom()
			}
		}
		return action, event
	})

	app.sidebar.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if app.sidebar.View.GetItemCount() > 0 {
				app.currentRoom = app.getCurrentRoom()
			}
		}
		return event
	})

	app.textInput.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if app.currentRoom == nil || app.textInput.View.GetText() == "" {
				return event
			}
			message := app.textInput.View.GetText()
			peer := app.currentRoom
			app.currentRoom.AddMessage(message, app.owner.Name)
			if err := peer.SendMessage(app.owner.Id, message, app.owner.DH); err != nil {
				utils.LL.Error("SendMessage: %s", err.Error())
				app.owner.Repo.Delete(peer.Id)
				app.textView.View.SetText("")
				app.currentRoom = nil
				app.ui.SetFocus(app.sidebar.View)
			}
			app.textInput.View.SetText("")
		}
		return event
	})
}

func (app *App) renderMessages() {
	timeStr := time.Now().Format("Current time is 15:04:05")
	app.textView.View.SetTitle(timeStr).
		SetTitleColor(tcell.ColorGreen)
	if app.currentRoom != nil {
		app.textInput.View.SetDisabled(false)
		app.textView.RenderMessages(app.currentRoom.Messages, app.owner.Name)
		app.textView.View.SetTitle(fmt.Sprintf("%s | Chatting with %s", timeStr, app.currentRoom.Name))
	} else {
		app.textInput.View.SetDisabled(true)
	}
}

func (app *App) getCurrentRoom() *entity.Room {
	_, id := app.sidebar.View.GetItemText(app.sidebar.View.GetCurrentItem())
	if id == "" {
		return nil
	}
	room, found := app.owner.Repo.Get(id)
	if !found {
		return nil
	}
	return room
}

func (app *App) run() {
	ticker := time.NewTicker(reprintFrequency)
	go func() {
		for {
			<-ticker.C
			app.ui.QueueUpdateDraw(app.sidebar.Reprint)
			app.ui.QueueUpdateDraw(app.renderMessages)
		}
	}()
}
