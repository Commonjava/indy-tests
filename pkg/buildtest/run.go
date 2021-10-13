package buildtest

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	common "github.com/commonjava/indy-tests/pkg/common"
)

const (
	TMP_DOWNLOAD_DIR = "/tmp/download"
	TMP_UPLOAD_DIR   = "/tmp/upload"
	BUILD_TEST_      = "build-test-"
)

var (
	versionRegexp = regexp.MustCompile(`redhat-([0-9]+)`)
)

func Run(originalIndy, foloId, replacement, targetIndy, buildType string, processNum int) {
	origIndy := originalIndy
	if !strings.HasPrefix(origIndy, "http://") {
		origIndy = "http://" + origIndy
	}
	foloTrackContent := common.GetFoloRecord(origIndy, foloId)
	DoRun(originalIndy, replacement, targetIndy, buildType, foloTrackContent, processNum)
}

func DoRun(originalIndy, replacement, targetIndy, buildType string, foloTrackContent common.TrackedContent, processNum int) string {
	_, validated := common.ValidateTargetIndy(originalIndy)
	if !validated {
		os.Exit(1)
	}
	targetIndyHost, validated := common.ValidateTargetIndy(targetIndy)
	if !validated {
		os.Exit(1)
	}

	newBuildName := generateRandomBuildName()

	// Prepare the indy repos for the whole testing
	buildMeta := decideMeta(buildType)
	if !prepareIndyRepos("http://"+targetIndyHost, newBuildName, *buildMeta) {
		os.Exit(1)
	}

	prepareCacheDirectories()

	downloads := prepareDownloadEntriesByFolo(targetIndy, newBuildName, foloTrackContent)
	downloadFunc := func(originalArtiURL, targetArtiURL string) bool {
		fileLoc := path.Join(TMP_DOWNLOAD_DIR, path.Base(targetArtiURL))
		return common.DownloadFile(targetArtiURL, fileLoc)
	}
	broken := false
	if len(downloads) > 0 {
		fmt.Println("Start handling downloads artifacts.")
		fmt.Printf("==========================================\n\n")
		if processNum > 1 {
			broken = !concurrentRun(processNum, downloads, downloadFunc)
		} else {
			for _, urls := range downloads {
				broken = !downloadFunc(urls[0], urls[1])
				if broken {
					break
				}
			}
		}
		fmt.Println("==========================================")
		if broken {
			fmt.Printf("Build test failed due to some downloading errors. Please see above logs to see the details.\n\n")
			os.Exit(1)
		}
		fmt.Printf("Downloads artifacts handling finished.\n\n")
	}

	uploadFunc := func(originalArtiURL, targetArtiURL string) bool {
		cacheFile := path.Join(TMP_UPLOAD_DIR, path.Base(originalArtiURL))
		downloaded := common.DownloadUploadFileForCache(originalArtiURL, cacheFile)
		if downloaded {
			return common.UploadFile(targetArtiURL, cacheFile)
		}
		return false
	}

	uploads := prepareUploadEntriesByFolo(originalIndy, targetIndy, newBuildName, foloTrackContent)

	if len(uploads) > 0 {
		fmt.Println("Start handling uploads artifacts.")
		fmt.Printf("==========================================\n\n")
		if processNum > 1 {
			broken = !concurrentRun(processNum, uploads, uploadFunc)
		} else {
			for _, artiURLs := range uploads {
				broken = !uploadFunc(artiURLs[0], artiURLs[1])
				if broken {
					break
				}
			}
		}
		fmt.Println("==========================================")
		if broken {
			fmt.Printf("Build test failed due to some uploadig errors. Please see above logs to see the details.\n\n")
			os.Exit(1)
		}

		fmt.Printf("Uploads artifacts handling finished.\n\n")
	}
	if !broken {
		targIndy := targetIndy
		if !strings.HasPrefix(targIndy, "http://") {
			targIndy = "http://" + targIndy
		}
		if common.SealFoloRecord(targIndy, newBuildName) {
			fmt.Printf("Folo record sealing succeeded for %s", newBuildName)
		} else {
			fmt.Printf("Warning: folo record sealing failed for %s", newBuildName)
		}
	}

	return newBuildName
}

// For downloads entries, we will get the paths and inject them to the final url of target indy
// as they should be directly download from target indy.
func prepareDownloadEntriesByFolo(targetIndyURL, newBuildId string, foloRecord common.TrackedContent) map[string][]string {
	targetIndy := normIndyURL(targetIndyURL)
	result := make(map[string][]string)
	for _, down := range foloRecord.Downloads {
		downUrl := fmt.Sprintf("%sapi/folo/track/%s/maven/group/%s%s", targetIndy, newBuildId, newBuildId, down.Path)
		downTuple := []string{"", downUrl}
		result[down.Path] = downTuple
	}
	return result
}

// For uploads entries, firstly they should be downloaded from original indy server as they may not
// exist in target indy server, so need to use original indy server to make the download url, and
// use the target indy server to make the upload url
func prepareUploadEntriesByFolo(originalIndyURL, targetIndyURL, newBuildId string, foloRecord common.TrackedContent) map[string][]string {
	originalIndy := normIndyURL(originalIndyURL)
	targetIndy := normIndyURL(targetIndyURL)
	result := make(map[string][]string)
	for _, up := range foloRecord.Uploads {
		orgiUpUrl, targUpUrl := createUploadUrls(originalIndy, targetIndy, newBuildId, up)
		result[up.Path] = []string{orgiUpUrl, targUpUrl}
	}
	return result
}

func createUploadUrls(originalIndy, targetIndy, newBuildId string, up common.TrackedContentEntry) (string, string) {
	storePath := common.StoreKeyToPath(up.StoreKey) // original store, e.g, maven/hosted/build-1234
	uploadPath := path.Join("api/content", storePath, up.Path)
	orgiUpUrl := fmt.Sprintf("%s%s", originalIndy, uploadPath)                    // original url to retrieve artifact
	alteredUploadPath := alterUploadPath(up.Path, newBuildId[len(BUILD_TEST_):])  // replace version number
	toks := strings.Split(storePath, "/")                                         // get package/type, e.g., maven/hosted
	targetStorePath := path.Join(toks[0], toks[1], newBuildId, alteredUploadPath) // e.g, maven/hosted/build-913413/org/...
	targUpUrl := fmt.Sprintf("%sapi/folo/track/%s/%s", targetIndy, newBuildId, targetStorePath)
	return orgiUpUrl, targUpUrl
}

func alterUploadPath(rawPath, buildNumber string) string {
	return versionRegexp.ReplaceAllString(rawPath, "redhat-"+buildNumber) // replace with same build number
}

func normIndyURL(indyURL string) string {
	indy := indyURL
	if !strings.HasPrefix(indy, "http://") {
		indy = "http://" + indy
	}
	if !strings.HasSuffix(indy, "/") {
		indy = indy + "/"
	}
	return indy
}

func prepareCacheDirectories() {
	if !common.FileOrDirExists(TMP_DOWNLOAD_DIR) {
		os.Mkdir(TMP_DOWNLOAD_DIR, os.FileMode(0755))
	}
	if !common.FileOrDirExists(TMP_DOWNLOAD_DIR) {
		fmt.Printf("Error: cannot create directory %s for file downloading.\n", TMP_DOWNLOAD_DIR)
		os.Exit(1)
	}
	if !common.FileOrDirExists(TMP_UPLOAD_DIR) {
		os.Mkdir(TMP_UPLOAD_DIR, os.FileMode(0755))
	}
	if !common.FileOrDirExists(TMP_UPLOAD_DIR) {
		fmt.Printf("Error: cannot create directory %s for caching uploading files.\n", TMP_UPLOAD_DIR)
		os.Exit(1)
	}
}

// generate a random 5 digit  number for a build repo like "build-test-9xxxxx"
func generateRandomBuildName() string {
	rand.Seed(time.Now().UnixNano())
	min := 900000
	max := 999999
	return fmt.Sprintf(BUILD_TEST_+"%v", rand.Intn(max-min)+min)
}

func concurrentRun(numWorkers int, artifacts map[string][]string, job func(originalArtiURL, targetArtiURL string) bool) bool {
	fmt.Printf("Start to run job in concurrent mode with thread number %v\n", numWorkers)
	ch := make(chan []string, numWorkers*5) // This buffered number of chan can be anything as long as it's larger than numWorkers
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results = []bool{}

	// This starts numWorkers number of goroutines that wait for something to do
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				a, ok := <-ch
				if !ok { // if there is nothing to do and the channel has been closed then end the goroutine
					wg.Done()
					return
				}
				mu.Lock()
				results = append(results, job(a[0], a[1]))
				mu.Unlock()
			}
		}()
	}

	// Now the jobs can be added to the channel, which is used as a queue
	for _, artifact := range artifacts {
		ch <- artifact // add artifact to the queue
	}

	close(ch) // This tells the goroutines there's nothing else to do
	wg.Wait() // Wait for the threads to finish

	finalResult := true
	for _, result := range results {
		if finalResult = result; !finalResult {
			break
		}
	}

	return finalResult
}
