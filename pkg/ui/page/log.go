package page

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/slok/brigadeterm/pkg/controller"
)

const (
	// JobLogPageName is the name that identifies the JobLogPage page.
	JobLogPageName = "joblog"
)

const (
	jobInfoFMT = `
%[1]sJob: [white]%s
%[1]sID: [white]%s
%[1]sStarted: [white]%s
%[1]sDuration: [white]%v`
	jobLogUsage = `[yellow](F5) [white]Reload	[yellow](ESC) [white]Back`
)

// JobLog is the page where a build job log will be available.
type JobLog struct {
	controller controller.Controller
	router     *Router

	// page layout.
	layout tview.Primitive

	// components.
	jobsInfo *tview.TextView
	logBox   *tview.TextView
	usage    *tview.TextView

	registerPageOnce sync.Once
}

// NewJobLog returns a new JobLog.
func NewJobLog(controller controller.Controller, router *Router) *JobLog {
	j := &JobLog{
		controller: controller,
		router:     router,
	}
	j.createComponents()
	return j
}

// createComponents will create all the layout components.
func (j *JobLog) createComponents() {
	j.jobsInfo = tview.NewTextView().
		SetDynamicColors(true)
	j.jobsInfo.SetBorder(true).
		SetBorderColor(tcell.ColorYellow)

	j.logBox = tview.NewTextView().
		SetDynamicColors(true)
	j.logBox.SetBorder(true).
		SetTitle("Log")

	j.usage = tview.NewTextView().
		SetDynamicColors(true)

	// Create the layout.
	j.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(j.jobsInfo, 0, 1, false).
		AddItem(j.logBox, 0, 5, true).
		AddItem(j.usage, 1, 1, false)
}

// Register satisfies Page interface.
func (j *JobLog) Register(pages *tview.Pages) {
	j.registerPageOnce.Do(func() {
		pages.AddPage(JobLogPageName, j.layout, true, false)
	})
}

// BeforeLoad satisfies Page interface.
func (j *JobLog) BeforeLoad() {
}

// Refresh will refresh all the page data.
func (j *JobLog) Refresh(projectID, buildID, jobID string) {
	// TODO: Get context.
	ctx := j.controller.JobLogPageContext(jobID)
	j.fill(ctx)

	// Set key handlers.
	j.logBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5:
			// Reload.
			j.router.LoadJobLog(projectID, buildID, jobID)
		case tcell.KeyEsc:
			// Back.
			j.router.LoadBuildJobList(projectID, buildID)
		}
		return event
	})
}

func (j *JobLog) fill(ctx *controller.JobLogPageContext) {
	j.fillUsage()
	j.fillBuildInfo(ctx)
	j.fillLog(ctx.Log)
}

func (j *JobLog) fillUsage() {
	j.usage.Clear()
	j.usage.SetText(jobLogUsage)
}

func (j *JobLog) fillBuildInfo(ctx *controller.JobLogPageContext) {
	color := "[red]"
	j.jobsInfo.SetBorderColor(tcell.ColorRed)
	if !ctx.Job.FinishedOK {
		j.jobsInfo.SetBorderColor(tcell.ColorGreen)
		color = "[green]"
	}

	j.jobsInfo.Clear()
	info := fmt.Sprintf(jobInfoFMT,
		color,
		ctx.Job.Name,
		ctx.Job.ID,
		ctx.Job.Started,
		ctx.Job.Ended.Sub(ctx.Job.Started),
	)
	j.jobsInfo.SetText(info)
}

func (j *JobLog) fillLog(data []byte) {
	j.logBox.Clear()
	j.logBox.SetText(string(data))
}
