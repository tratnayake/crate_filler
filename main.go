package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli"
)

var (
	y2o       = "youtube-dl"
	mode      = ""
	inputFile = ""
	outputDir = ""
	urlList   = []string{}
)

func main() {
	app := cli.NewApp()
	app.Name = "crateFiller"
	app.Copyright = "Copyright 2018 Hugh Brown and licensed under the GPL v3. Original Source: https://gitlab.com/saintaardvark/prinbox/blob/master/prinbox.go"
	app.Usage = "Download mp3s from Youtube and keep current on playlists/channels"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "mode,m",
			Usage:       "What mode should cratefiller operate in? SONGS,CHANNELS or FETCH",
			Value:       mode,
			Destination: &mode,
		},
		cli.StringFlag{
			Name:        "input,i",
			Usage:       "The input file that will contain the list of songs or playlists",
			Value:       inputFile,
			Destination: &inputFile,
		},

		cli.StringFlag{
			Name:        "output,o",
			Usage:       "The output directory that downloaded mp3's should be downloaded into",
			Value:       outputDir,
			Destination: &outputDir,
		},
	}
	app.Action = func(c *cli.Context) error {
		if mode == "" {
			err := fmt.Errorf("A mode must be specified")
			return err
		}

		fmt.Println("Checking output directory")
		if outputDir == "" {
			err := fmt.Errorf("An output directory must be specified")
			return err
		}

		switch mode {
		case "file":
			grabAudio(c)
		}

		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

//Functions

func grabAudio(c *cli.Context) {
	checkForPrereqs()
	var err error
	//Check for file
	if inputFile != "" {
		if urlList, err = readFromFile(inputFile); err != nil {
			log.Println("[INFO] Can't open list of files: " + err.Error())
		}
		fmt.Println(urlList)

		for _, url := range urlList {
			log.Println("[INFO] Grabbing " + url)
			log.Println("[INFO] " + buildOutputArg(outputDir)) // output for debug
			cmd := &exec.Cmd{}
			// youtube-dl -x --audio-format mp3 --audio-quality 0 -o "~/Music/%(title)s-%(upload_date)s.%(ext)s" --embed-thumbnail https://www.youtube.com/watch\?v\=LMZ5RDHsvks
			if strings.Contains(url, "youtube") {
				cmd = exec.Command(y2o,
					"--extract-audio",
					"--audio-format",
					"mp3",
					"--restrict-filenames",
					"--embed-thumbnail",
					buildOutputArg(outputDir),
					url)
			} else if strings.Contains(url, "mp3") {
				cmd = exec.Command("wget",
					"--directory-prefix",
					outputDir,
					url)
			} else if url == "" {
				log.Println("[INFO] Skipping empty line")
				continue
			} else {
				log.Println("[INFO] I don't recognize that line, skipping it")
				continue
			}
			if cmdOut, err := cmd.Output(); err != nil {
				log.Printf("[ERROR] Unknown error: %+v\n", err)
				log.Printf("[ERROR] Command was: %+v\n", cmd)
				log.Println("[ERROR] Output of : " + string(cmdOut) + string(err.(*exec.ExitError).Stderr))
				log.Printf("[ERROR] Output of %s: %s %s\n",
					string(cmd.Path),
					string(cmdOut),
					string(err.(*exec.ExitError).Stderr))
				log.Fatal(err)
			} else {
				log.Println("[INFO] " + string(cmdOut))
			}
			// if sleepPlease == true {
			// 	log.Println("[INFO] Sleeping a bit between fetches")
			// 	time.Sleep(time.Duration(rand.Intn(15)) * time.Second)
			// }
		}

	} else {
		if c.Args().Get(0) == "" {
			log.Println("[INFO] No URL, skipping download")
			return
		}
		log.Println("[INFO] Looks like I should be able to get " + c.Args().Get(0))
		urlList = append(urlList, c.Args().Get(0))
	}
}

// checkForPrereqs looks for the programs we need for youtube-dl
func checkForPrereqs() {
	if _, err := exec.LookPath("youtube-dl"); err != nil {
		log.Fatal("Can't find youtube-dl in PATH -- please install!\n")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		if _, err := exec.LookPath("avconv"); err != nil {
			log.Fatal("Can't find ffmpeg or avconv in PATH -- please install!\n")
		}
	}
}

// readFromFile reads a list of URLs from a file and returns them
func readFromFile(f string) ([]string, error) {
	content, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

// buildOutputArg builds the -o argument for youtube-dl
func buildOutputArg(dir string) (arg string) {
	arg = fmt.Sprintf("-o%s/%%(title)s-%%(upload_date)s.%%(ext)s", dir)
	return arg
}
