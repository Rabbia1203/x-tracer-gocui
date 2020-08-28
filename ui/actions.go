package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/Sheenam3/x-tracer-gocui/events"

)

//var DEBUG_DISPLAYED bool = false
var NAMESPACES_DISPLAYED bool = false




// Global action: Quit

func actionGlobalQuit(g *gocui.Gui, v *gocui.View) error {

	if err := deletePod("x-agent"); err != nil {
		return err
	}

	return gocui.ErrQuit
}

/*// Global action: Toggle debug
func actionGlobalToggleViewDebug(g *gocui.Gui, v *gocui.View) error {
	vn := "debug"

	if !DEBUG_DISPLAYED {
		debug(g, "Action: Display debug popup")
		g.SetViewOnTop(vn)
		g.SetCurrentView(vn)
	} else {
		debug(g, "Action: Hide debug popup")
		g.SetViewOnBottom(vn)
		g.SetCurrentView("pods")
	}

	DEBUG_DISPLAYED = !DEBUG_DISPLAYED

	return nil
}
*/

// View namespaces: Toggle display
func actionGlobalToggleViewNamespaces(g *gocui.Gui, v *gocui.View) error {
	vn := "namespaces"

	if !NAMESPACES_DISPLAYED {
//		debug(g, "Action: Display namespaces popup")
		g.SetViewOnTop(vn)
		g.SetCurrentView(vn)
		changeStatusContext(g, "SE")
	} else {
//		debug(g, "Action: Hide namespaces popup")
		g.SetViewOnBottom(vn)
		g.SetCurrentView("pods")
		changeStatusContext(g, "D")
	}

	NAMESPACES_DISPLAYED = !NAMESPACES_DISPLAYED

	return nil
}

// View pods: Up
func actionViewPodsUp(g *gocui.Gui, v *gocui.View) error {
	moveViewCursorUp(g, v, 2)
//	debug(g, "Select up in pods view")
	return nil
}

// View pods: Down
func actionViewPodsDown(g *gocui.Gui, v *gocui.View) error {
	moveViewCursorDown(g, v, false)
//	debug(g, "Select down in pods view")
	return nil
}


//Display Probe Tools after Pod select
func actionViewPodsSelect(g *gocui.Gui, v *gocui.View) error {
        line,err  := getViewLine(g,v)
        if err != nil {
                return err
        }
//      maxX, maxY := g.Size()
        LOG_MOD = "pod"
        errr := showSelectProbe(g)

        changeStatusContext(g, "SL")
//      viewLogs(g, maxX, maxY)
        displayConfirmation(g, line+" Pod selected")
        return errr

}

// View pods: Delete
func actionViewPodsDelete(g *gocui.Gui, v *gocui.View) error {
	p, err := getSelectedPod(g)
	if err != nil {
		return err
	}

	if err := deletePod(p); err != nil {
		return err
	}

	//debug(g, "Delete pod: "+p)
	displayConfirmation(g, p+" pod deleted")

	go viewPodsRefreshList(g)

	return nil
}

/*// View pods: Logs
func actionViewPodsLogs(g *gocui.Gui, v *gocui.View) error {
	LOG_MOD = "pod"
	err := showViewPodsLogs(g)

	changeStatusContext(g, "SL")

	return err
}*/

// View pod logs: Up
func actionViewPodsLogsUp(g *gocui.Gui, v *gocui.View) error {
	/*vLc, err := g.View("logs-containers")
	if err != nil {
		return err
	}*/
	moveViewCursorUp(g, v, 0)
	events.PublishEvent("logs:refresh", events.EmptyMessage{})
	//debug(g, "Select up in logs view")
	return nil
}

// View pod logs: Down
func actionViewPodsLogsDown(g *gocui.Gui, v *gocui.View) error {
	/*vLc, err := g.View("logs-containers")
	if err != nil {
		return err
	}*/
	moveViewCursorDown(g, v, false)
	events.PublishEvent("logs:refresh", events.EmptyMessage{})
	//debug(g, "Select down in logs view")
	return nil
}

// View logs: Hide
func actionViewPodsLogsHide(g *gocui.Gui, v *gocui.View) error {
	g.SetViewOnBottom("logs")
	g.SetViewOnBottom("logs-containers")
	g.SetCurrentView("pods")

	v.Clear()

	changeStatusContext(g, "D")

	//debug(g, "Action: Hide view logs)")

	return nil
}

// View namespaces: Up
func actionViewNamespacesUp(g *gocui.Gui, v *gocui.View) error {
	moveViewCursorUp(g, v, 0)
	//debug(g, "Select up in namespaces view")
	return nil
}

// View namespaces: Down
func actionViewNamespacesDown(g *gocui.Gui, v *gocui.View) error {
	moveViewCursorDown(g, v, false)
	//debug(g, "Select down in namespaces view")
	return nil
}

// Namespace: Choose
func actionViewNamespacesSelect(g *gocui.Gui, v *gocui.View) error {
	line, err := getViewLine(g, v)
	//debug(g, "Select namespace: "+line)
	NAMESPACE = line
	go viewPodsRefreshList(g)
	actionGlobalToggleViewNamespaces(g, v)
	displayConfirmation(g, line+" namespace selected")
	return err
}




// Probes:  Choose
func actionViewProbesSelect(g *gocui.Gui, v *gocui.View) error {
	line, err := getViewLine(g, v)
	LOG_MOD = "probe"

        G,p,lv := showViewPodsLogs(g)
	displayConfirmation(g, line+" probe selected")
	startAgent(G,p,lv,line)
	G.SetViewOnTop("logs")
	G.SetCurrentView("logs")
	//debug(g, "Select namespace: "+line)
	//NAMESPACE = line
	//go viewPodsRefreshList(g)
	//actionGlobalToggleViewNamespaces(g, v)

	return err
}


func actionViewProbesList(g *gocui.Gui, v *gocui.View) error {

	if err := deletePod("x-agent"); err != nil {
		return err
	}

	LOG_MOD = "pod"
	errr := showSelectProbe(g)
        changeStatusContext(g, "SL")

	g.DeleteView("logs")

	return errr
}
