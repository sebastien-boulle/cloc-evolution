package main

import "fmt"
import "flag"
import "io/ioutil"
import "os"
import "path"
import "log"
import "sort"
import "regexp"
import "strconv"
import "bufio"
import "bytes"
import "strings"
import "text/template"
import "os/exec"

import "gopkg.in/libgit2/git2go.v28"

type VersionLoc struct {
	Version string
	Locs map[string]int64
	SumLocs int64
}

const templateHTML = `<!DOCTYPE html>
<html>
  <head>
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
    <script src="https://code.highcharts.com/highcharts.js"></script>
    <script src="https://code.highcharts.com/modules/exporting.js"></script>
    <script src="http://underscorejs.org/underscore.js"></script>
  </head>
  <body>
    <div id="graph"/>
  </body>
  <script type="text/javascript">
    $(function () {
        versionLOCs = [
        {{range .}}
          {
            version: "{{.Version}}",
            totalLOC: "{{.SumLocs}}",
            locs: {
              {{range $key, $value := .Locs}}
              "{{$key}}": {{$value}},
              {{end}}
            }
          },
        {{end}}
        ]

        // Cater for some languages appearing and disappearing
        languages = _.chain(versionLOCs).map(function(version) { return _.keys(version.locs);}).flatten().uniq().value()

        $('#graph').highcharts({
            chart: {
                type: 'spline'
            },
            title: {
                text: 'Lines of Code across source code versions'
            },
            xAxis: {
                categories: _.pluck(versionLOCs, "version")
            },
            yAxis: {
                title: {
                    text: 'Number of lines of code'
                }
            },
            plotOptions: {
                line: {
                    dataLabels: {
                        enabled: true
                    },
                    enableMouseTracking: true
                }
            },
            series: _.map(languages, function(language) { return {name: language, data: _.map(versionLOCs, function(version) { return version.locs[language];})}})
        });
    });
  </script>
</html>`

func getTargetDir() (string) {
	targetdir := flag.String("targetdir", path.Dir(os.Args[0]), "The location of the git repository")
	flag.Parse()

	return *targetdir
}

// golang doesn't define min for non-float64 types
func min(a, b int) (int) {
	if a < b {
		return a
	}

	return b
}

// Log and fail on errors. Note: This can potentially leave the git repo in a bad state.
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

// Define a custom type that implements the sort interface. Allows us to sort by semantic version number. See
// http://semver.org/
type BySemverNumber []string
func (s BySemverNumber) Swap (i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s BySemverNumber) Less(i, j int) bool {
	a := s[i]
	b := s[j]

	r := regexp.MustCompile("([0-9]+)\\.?")

	matches_a := r.FindAllStringSubmatch(a, -1)
	matches_b := r.FindAllStringSubmatch(b, -1)

	min_len := min(len(matches_a), len(matches_b))

	for i := 0; i < min_len; i++ {
		a_int, _ := strconv.Atoi(matches_a[i][1])
		b_int, _ := strconv.Atoi(matches_b[i][1])
		if a_int < b_int {
			return true
		} else {
			if b_int < a_int {
				return false
			}
		}
	}

	if len(matches_a) < len(matches_b) {
		return true
	}

	return false
}
func (s BySemverNumber) Len() int {

	return len(s)
}

func parseClocOutput(clocOutput []byte) (map[string]int64, int64) {
	parsedOutput := make(map[string]int64)

	scanner := bufio.NewScanner(bytes.NewReader(clocOutput))

	// Matches cloc's supported languages: http://cloc.sourceforge.net/#Languages
	r := regexp.MustCompile(`^([\w#\+\/\.:]*[ \d]+?)\s+\d+\s+\d+\s+\d+\s+(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		data := r.FindStringSubmatch(line)

		if len(data) == 3 {
			key := strings.TrimSpace(data[1])
			parsed, err := strconv.ParseInt(data[2], 10, 64)

			if key == "SUM:" {
				return parsedOutput, parsed
			}
			checkErr(err)
			parsedOutput[key] = parsed
		}
	}

	panic("Invalid cloc output detected. Perhaps this is an untested version?")
}

func getRepo(targetdir string) (*git.Repository) {
	repo, err := git.OpenRepository(targetdir)
	checkErr(err)
	return repo
}

// sort.Sort sorts inline
func getSortedTags(repo *git.Repository) ([]string) {
	tags, err := repo.Tags.List()
	checkErr(err)

	sort.Sort(BySemverNumber(tags))
	return tags
}

func checkOutAndCloc(repo *git.Repository, tag, targetdir string) ([]byte) {
	fmt.Printf("Checking out tag %v in %s\n", tag, targetdir)
	fmt.Printf("Running cloc on %s\n", targetdir)

	newHead := fmt.Sprintf("refs/tags/%s", tag)
	repo.SetHead(newHead)
	repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce,})

	cmd := exec.Command("cloc", targetdir)
	out, err := cmd.Output()
	checkErr(err)

	return out
}

func writeHTMLTemplateAndOpen(versionLocs []VersionLoc) {
	file, err := ioutil.TempFile(os.TempDir(), "cloc-evolution")
	checkErr(err)
	//defer os.Remove(file.Name())

	templater := template.Must(template.New("cloc").Parse(templateHTML))
	templater.Execute(file, versionLocs)
	checkErr(err)

	err = exec.Command("xdg-open", file.Name()).Start()
	checkErr(err)
}

func main() {
	targetdir := getTargetDir()
	repo := getRepo(targetdir)
	tags := getSortedTags(repo)

	var versionlocs []VersionLoc

	for index := range tags {
		clocOutput := checkOutAndCloc(repo, tags[index], targetdir)
		locs, maxLocs := parseClocOutput(clocOutput)

		fmt.Printf("Languages: %v\nTotal LOC: %v\n\n", locs, maxLocs)
		versionlocs = append(versionlocs, VersionLoc{Version: tags[index], Locs: locs, SumLocs: int64(maxLocs)})
	}

	writeHTMLTemplateAndOpen(versionlocs)
}
