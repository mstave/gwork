package worker

import (
	"testing"
	"fmt"
)

func TestSimple(t *testing.T) {
	t.Log("test")
}

func TestComboRun(t *testing.T) {
	tJob, sampleOutChan := getSampleJob()
	outChan := make(chan string)
	job2 := StringChanJob{inChan: sampleOutChan,
		outChan: outChan,
		customWork: func(inChan, outChan chan string) {
			for s := range inChan {
				fmt.Println(s)
			}
			outChan <- "Done"
			close(outChan)
		}}
	job3 := trivialIntJob{3}
	go RunJobs(&tJob, &job2, &job3)
	<-outChan
}

func TestFullPipe(t *testing.T) {
	dirs := make(chan string, 1)
	files := make(chan string)
	done := make(chan string)
	const cwd = ".."
	dirs <- cwd
	dirFinder := FileFinderJob{dirChan:dirs, fileChan:files, followLinks:true}
	strEcho := StringChanJob{inChan: files, outChan: done, customWork: func(inC, outC chan string) {
		for s := range inC {
			fmt.Println(s)
		}
		outC <- "done"
	}}
	RunJobs(&dirFinder, &strEcho)
	<-done

}

func TestWorker_Stop(t *testing.T) {
	j, out  := getSampleJob()
	p := RunJobs(&j)
	p.DrainPool()
	for range out {
	}
}
