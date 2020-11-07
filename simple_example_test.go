package selenium_go

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	// These paths will be different on your system.
	seleniumPath    = "./tmp/selenium-server.jar"
	geckoDriverPath = "./tmp/geckodriver"
	seleniumPort    = 8080
)

func TestMain(m *testing.M) {
	_, errSelenium := os.Stat(seleniumPath)
	_, errGecko := os.Stat(geckoDriverPath)
	if errSelenium != nil || errGecko != nil {
		cmd := exec.Command("go", "run", "../pkg/selenium_base/init.go", "--alsologtostderr", "--download_browsers", "--download_latest")
		cmd.Dir = "./tmp"
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
	code := m.Run()
	os.Exit(code)
}

func TestExample(t *testing.T) {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, seleniumPort, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", seleniumPort))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	t.Run("Should execute code in playground of golang", func(t *testing.T) {
		// Navigate to the simple playground interface.
		if err := wd.Get("http://play.golang.org/?simple=1"); err != nil {
			panic(err)
		}

		// Get a reference to the text box containing code.
		elem, err := wd.FindElement(selenium.ByCSSSelector, "#code")
		if err != nil {
			panic(err)
		}
		// Remove the boilerplate code already in the text box.
		if err := elem.Clear(); err != nil {
			panic(err)
		}

		// Enter some new code in text box.
		err = elem.SendKeys(`
		package main
		import "fmt"

		func main() {
			fmt.Println("Hello Selenium Go!")
		}`)
		if err != nil {
			panic(err)
		}

		// Click the run button.
		btn, err := wd.FindElement(selenium.ByCSSSelector, "#run")
		if err != nil {
			panic(err)
		}
		if err := btn.Click(); err != nil {
			panic(err)
		}

		// Wait for the program to finish running and get the output.
		outputDiv, err := wd.FindElement(selenium.ByCSSSelector, "#output")
		if err != nil {
			panic(err)
		}

		var output string
		for {
			output, err = outputDiv.Text()
			if err != nil {
				panic(err)
			}
			if output != "Waiting for remote server..." {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		fmt.Printf("%s", strings.Replace(output, "\n\n", "\n", -1))

		assert.Equal(t, output, "Hello Selenium Go!\n\nProgram exited.")
	})
}