package worker

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
)

func getSampleJob() (StringChanJob, chan string) {
	inChan := make(chan string, 3)
	outChan := make(chan string)
	inChan <- "Larry"
	inChan <- "Moe"
	inChan <- "Curly"
	close(inChan)
	theJob := StringChanJob{inChan: inChan, outChan: outChan, customWork: func(inChan, outChan chan string) {
		for s := range inChan {
			outChan <- "Stooge: " + s
		}
		close(outChan)
	}}
	return theJob, outChan
}
func TestFileFinderJob_WalkFuncSrcData(t *testing.T) {
	dChan := make(chan string, 1)
	fChan := make(chan string)
	dChan <- ".."
	fJ := FileFinderJob{dirChan: dChan, fileChan: fChan, followLinks: false}
	go RunJob(&fJ)
	const golden = "worker_test.go"
	evalStrChannel(fChan, golden, t)
}

func evalStrChannel(fChan chan string, golden string, t *testing.T) string {
	var allFiles string
	for aFile := range fChan {
		allFiles += aFile + "\n"
	}
	if len(allFiles) == 0 || !strings.Contains(allFiles, golden) {
		t.Error("expecting to find", golden, "in ", allFiles)
	}
	return allFiles
}

func TestFileFinderJob_Work_WalkFunc_TmpFile(t *testing.T) {
	dChan := make(chan string, 1000)
	fChan := make(chan string)
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatalf("Error creating temp dir for tests: %v", err)
	}
	defer os.RemoveAll(dir)
	os.MkdirAll(dir+"/subdir/123", 0777)
	const golden = "worker_test.go"
	f, err := os.OpenFile(dir+"/subdir/123/"+golden, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	dChan <- dir
	fJ := FileFinderJob{dirChan: dChan, fileChan: fChan, followLinks: true}
	go RunJob(&fJ)
	evalStrChannel(fChan, golden, t)
}

func TestFileFinderJob_Work_WalkFunc_TmpFileSymLinks(t *testing.T) {
	dChan := make(chan string, 1000)
	fChan := make(chan string)
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatalf("Error creating temp dir for tests: %v", err)
	}
	defer os.RemoveAll(dir)
	nested := dir + "/subdir/nested"
	os.MkdirAll(nested, 0777)
	const golden = "target.file"
	f, err := os.OpenFile(nested+string(os.PathSeparator)+golden, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	os.Symlink(nested, dir+"/nestLink")
	dChan <- dir
	fJ := FileFinderJob{dirChan: dChan, fileChan: fChan, followLinks: true}
	go RunJob(&fJ)
	var allFiles string
	var count int
	for aFile := range fChan {
		if strings.Contains(aFile, golden) {
			count++
		}
		allFiles += aFile + "\n"
	}
	if count != 2 {
		t.Error("expected 2 copies of ", golden, "in ", allFiles, " found: ", count)
	}
}

func TestSampleJob(t *testing.T) {
	j, out := getSampleJob()
	go RunJob(&j)
	for s := range out {
		t.Log(s)
	}

}
func TestTrivialIntJob_Work(t *testing.T) {
	RunJob(&trivialIntJob{3})
}

var bCount int // to discourage the compiler from cheating on benchmarks

func BenchmarkFileFinderJob_WalkFunc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dChan := make(chan string, 1)
		fChan := make(chan string)
		dChan <- runtime.GOROOT()
		//b.ResetTimer()
		fJ := FileFinderJob{dirChan: dChan, fileChan: fChan, followLinks: false}
		go RunJob(&fJ)

		//var allFiles []string
		var fcount int
		for range fChan {
			for range fChan {
				//allFiles = append(allFiles, aFile)
				fcount++
			}
		}
		bCount = fcount
	}
}


