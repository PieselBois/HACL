package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	arg "github.com/alexflint/go-arg"
)

var args struct {
	APIURL string `arg:"-u" help:"api url"`
}

type pipeline struct {
	ID          string `json:"id"`
	CloneURL    string `json:"clone_url"`
	ProjectName string `json:"project_name"`
	CommitHash  string `json:"CommitHash"`
}

type warningInfo struct {
	Kind    string `json:"kind"`
	File    string `json:"file"`
	Lines   string `json:"lines"`
	Message string `json:"message"`
	Module  string `json:"module"`
}

func getPipeline() (pipeline, error) {
	req, err := http.NewRequest("GET", args.APIURL, nil)
	var res pipeline

	if err != nil {
		return res, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return res, err
	}

	err = json.Unmarshal(body, &res)

	if err != nil {
		return res, err
	}

	return res, nil
}

func run(p pipeline) error {
	c := exec.Command("git", "clone", p.CloneURL)
	err := c.Run()
	if err != nil {
		return err
	}
	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	err = os.Chdir(currentDir + "/" + p.ProjectName)
	if err != nil {
		return err
	}

	c = exec.Command("git", "checkout", p.CommitHash)
	err = c.Run()
	if err != nil {
		return err
	}

	c = exec.Command("halcy", "-c", ".halcy.yml")
	err = c.Run()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	arg.MustParse(&args)

	for {
		p, err := getPipeline()

		if err == nil {
			_ = run(p)
		} else {
			log.Print(err)
		}

		time.Sleep(60 * time.Second)
	}
}
