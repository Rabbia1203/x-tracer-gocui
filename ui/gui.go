package ui


import (
	"fmt"
	"context"
	"log"
	"os"
	"io"
	"strings"
	"time"
	pb "github.com/sheenam3/x-tracer/api"
//	"github.com/sheenam3/x-tracer/pkg/streamserver"
	"github.com/sheenam3/x-tracer/internal/agentmanager"
	"github.com/sheenam3/x-tracer/k8s"
	"github.com/jroimartin/gocui"

)

type server struct{}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}





var version = "master"
var LOG_MOD string = "pod"
var NAMESPACE string = "default"
var G *gocui.Gui

// Configure globale keys
var keys []Key = []Key{
	Key{"", gocui.KeyCtrlC, actionGlobalQuit},
//	Key{"", gocui.KeyCtrlD, actionGlobalToggleViewDebug},
	Key{"pods", gocui.KeyCtrlN, actionGlobalToggleViewNamespaces},
	Key{"pods", gocui.KeyArrowUp, actionViewPodsUp},
	Key{"pods", gocui.KeyArrowDown, actionViewPodsDown},
	Key{"pods", 'd', actionViewPodsDelete},
	Key{"pods", gocui.KeyEnter, actionViewPodsSelect},
	//Key{"logs", l, actionStreamLogs},
//	Key{"logs", 'l', actionViewPodsLogsHide},
//	Key{"logs", gocui.KeyArrowUp, actionViewPodsLogsUp},
//	Key{"logs", gocui.KeyArrowDown, actionViewPodsLogsDown},
	Key{"namespaces", gocui.KeyArrowUp, actionViewNamespacesUp},
	Key{"namespaces", gocui.KeyArrowDown, actionViewNamespacesDown},
	Key{"namespaces", gocui.KeyEnter, actionViewNamespacesSelect},
}

// Entry Point of the x-tracer
func InitGui() {
	c := getConfig()

	// Ask version
	if c.askVersion {
		fmt.Println(versionFull())
		os.Exit(0)
	}

	// Ask Help
	if c.askHelp {
		fmt.Println(versionFull())
		fmt.Println(HELP)
		os.Exit(0)
	}

	// Only used to check errors
	getClientSet()

	G, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer G.Close()

	G.Highlight = true
	G.SelFgColor = gocui.ColorGreen

	G.SetManagerFunc(uiLayout)

	if err := uiKey(G); err != nil {
		log.Panicln(err)
	}

	if err := G.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}




func GetView() *gocui.Gui {

return G

}

// Define the UI layout
func uiLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	//viewDebug(g, maxX, maxY)
	viewLogs(g, maxX, maxY)
	viewNamespaces(g, maxX, maxY)
	viewOverlay(g, maxX, maxY)
	viewTitle(g, maxX, maxY)
	viewPods(g, maxX, maxY)
	viewStatusBar(g, maxX, maxY)

	return nil
}

/* Useful to debug Pody (display with CTRL+D)
func debug(g *gocui.Gui, output interface{}) {
	v, err := g.View("debug")
	if err == nil {
		t := time.Now()
		tf := t.Format("2006-01-02 15:04:05")
		output = tf + " > " + output.(string)
		fmt.Fprintln(v, output)
	}
}*/

// Move view cursor to the bottom
func moveViewCursorDown(g *gocui.Gui, v *gocui.View, allowEmpty bool) error {
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	nextLine, err := getNextViewLine(g, v)
	if err != nil {
		return err
	}
	if !allowEmpty && nextLine == "" {
		return nil
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

// Move view cursor to the top
func moveViewCursorUp(g *gocui.Gui, v *gocui.View, dY int) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if cy > dY {
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

// Get view line (relative to the cursor)
func getViewLine(g *gocui.Gui, v *gocui.View) (string, error) {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	return l, err
}

// Get the next view line (relative to the cursor)
func getNextViewLine(g *gocui.Gui, v *gocui.View) (string, error) {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy + 1); err != nil {
		l = ""
	}

	return l, err
}

// Set view cursor to line
func setViewCursorToLine(g *gocui.Gui, v *gocui.View, lines []string, selLine string) error {
	ox, _ := v.Origin()
	cx, _ := v.Cursor()
	for y, line := range lines {
		if line == selLine {
			if err := v.SetCursor(ox, y); err != nil {
				if err := v.SetOrigin(cx, y); err != nil {
					return err
				}
			}
		}
	}
	return nil
}




// Get pod name form line
func getPodNameFromLine(line string) string {
	if line == "" {
		return ""
	}

	i := strings.Index(line, " ")
	if i == -1 {
		return line
	}

	return line[0:i]
}

// Get selected pod
func getSelectedPod(g *gocui.Gui) (string, error) {
	v, err := g.View("pods")
	if err != nil {
		return "", err
	}
	l, err := getViewLine(g, v)
	if err != nil {
		return "", err
	}
	p := getPodNameFromLine(l)

	return p, nil
}

// Show views logs
func showViewPodsLogs(g *gocui.Gui) error {
	vn := "logs"

	switch LOG_MOD {
	case "pod":
		// Get current selected pod
		p, err := getSelectedPod(g)
		if err != nil {
			return err
		}

		// Display pod containers
		vLc, err := g.View(vn + "-containers")
		if err != nil {
			return err
		}
		vLc.Clear()
		var conName []string
		for _, c := range getPodContainersName(p) {
			fmt.Fprintln(vLc, c)
			conName = append(conName, c)
		}
		vLc.SetCursor(0, 0)

		//Display Container IDs

                lv, err := g.View(vn)
                if err != nil {
                        return err
                }
                lv.Clear()

                fmt.Fprintln(lv, "Containers ID:")
                for i, conId := range getPodContainersID(p){
                        fmt.Fprintln(lv,conName[i] + "->" + conId)
		}

		fmt.Fprintln(lv, "Pod you choose is: " + p)

		startAgent(g,p,lv)
		k8s.StartServer()
		
		// Display logs
		//refreshPodsLogs(g)
	}


//	debug(g, "Action: Show view logs")
	g.SetViewOnTop(vn)
	g.SetViewOnTop(vn + "-containers")
	g.SetCurrentView(vn)

	return nil
}

// Refresh pods logs
func refreshPodsLogs(g *gocui.Gui) error {
	vn := "logs"

	// Get current selected pod
	p, err := getSelectedPod(g)
	if err != nil {
		return err
	}

	vLc, err := g.View(vn + "-containers")
	if err != nil {
		return err
	}

	c, err := getViewLine(g, vLc)
	if err != nil {
		return err
	}

	vL, err := g.View(vn)
	if err != nil {
		return err
	}
	getPodContainerLogs(p, c, vL)

	return nil
}

// Display error
func displayError(g *gocui.Gui, e error) error {
	lMaxX, lMaxY := g.Size()
	minX := lMaxX / 6
	minY := lMaxY / 6
	maxX := 5 * (lMaxX / 6)
	maxY := 5 * (lMaxY / 6)

	if v, err := g.SetView("errors", minX, minY, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		// Settings
		v.Title = " ERROR "
		v.Frame = true
		v.Wrap = true
		v.Autoscroll = true
		v.BgColor = gocui.ColorRed
		v.FgColor = gocui.ColorWhite

		// Content
		v.Clear()
		fmt.Fprintln(v, e.Error())

		// Send to forground
		g.SetCurrentView(v.Name())
	}

	return nil
}

// Hide error box
func hideError(g *gocui.Gui) {
	g.DeleteView("errors")
}

// Display confirmation message
func displayConfirmation(g *gocui.Gui, m string) error {
	lMaxX, lMaxY := g.Size()

	if v, err := g.SetView("confirmation", -1, lMaxY-3, lMaxX, lMaxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		// Settings
		v.Frame = false

		// Content
		fmt.Fprintln(v, textPadCenter(m, lMaxX))

		// Auto-hide message
		hide := func() {
			hideConfirmation(g)
		}
		time.AfterFunc(time.Duration(2)*time.Second, hide)
	}

	return nil
}

// Hide confirmation message
func hideConfirmation(g *gocui.Gui) {
	g.DeleteView("confirmation")
}





func startAgent(g *gocui.Gui, p string, o io.Writer) error {
	//vn := "logs"
	cs := getClientSet()

	var containerId []string

	/*lv, err := g.View(vn)
        if err != nil {
              return err
        }*/

        //lv.Clear()

	containerId = getPodContainersID(p)

	targetNode := getTargetNode(p)

	nodeIp := getNodeIp()

	agent := agentmanager.New(containerId[0], targetNode, nodeIp, cs)

	//Start x-agent Pod
	fmt.Fprintln(o, "Starting x-agent Pod...")

	agent.ApplyAgentPod()

	fmt.Fprintln(o, "Starting x-agent Service...")
	agent.ApplyAgentService()

	agent.SetupCloseHandler()

	return nil
}


/*func startServer() error {
	server := streamserver.New("6666")
	server.StartServer()

	for {
		log.Println("From x-tracer- Sleeping")
		time.Sleep(10 * time.Second)
	}
	return nil

}*/


