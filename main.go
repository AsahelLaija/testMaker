package main
import (
	"log"
	"os"
	"text/template"
	"fmt"
	"github.com/xuri/excelize/v2"
	"regexp"
	"strings"
)

const test = `
type {{ .Name }} struct {
	{{ range .ArgsNames }}{{.}}, {{ end }}expect	float64
}
var {{ .Name }}s = []{{ .Name }}{
	{{ .Name }}{ {{ range .Args }}{{.}}, {{ end }} {{.Expect}} },
}
func TestCalc{{ .Name }}(t *testing.T){
	for _, test := range {{ .Name }}s{
		out := Calc{{ .Name }}({{ range .ArgsNames }}test.{{ . }}, {{ end }})
		exp := test.expect *1.001 > out && out > test.expect /1.001
		if !exp{
			t.Errorf("output %v, expected %v", out, test.expect)
		}
	}
}


func Calc{{ .Name }}( {{ range .ArgsNames }}{{.}}, {{end}} float64 ) float64 {
	return {{ .Formula }}
}
`

type testValues struct {
	Name	string
	Expect	string
	Formula	string
	ArgsNames	[]string
	Args	[]string
}
func main() {
	//cellToBeTested := "P51"
	var name, cellToBeTested string

	fmt.Printf("type the name of the test: ")
	if _, err := fmt.Scanln(&name); err != nil {
		log.Fatal(err)
	}


	fmt.Printf("type the cell is constructed the test: ")
	if _, err := fmt.Scanln(&cellToBeTested); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")

	joistSS	:= "/home/renan/projects/JoistCalc/CalculusF.xlsx"
	sheet	:= "Joist_2"
	f, err := excelize.OpenFile(joistSS)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	expected, err := f.GetCellValue(sheet, cellToBeTested)
	if err != nil {
		fmt.Println(err)
		return
	}

	rawFormula, err := f.GetCellFormula(sheet, cellToBeTested)
	if err != nil {
		fmt.Println(err)
		return
	}

	formula := strings.Replace(rawFormula, "$", "", -1)
	fmt.Printf("%v\n\n", formula)
	pattern := regexp.MustCompile(`([A-Z][A-Z][A-Z]|[A-Z][A-Z]|[A-Z])+(\d\d\d|\d\d|\d)`)
	rawCells := pattern.FindAllString(formula, -1)
	cells := unique(rawCells)
	var args []string
	for _, cellVal := range cells {
		c, err := f.GetCellValue(sheet, cellVal)
		if c == "" { c = "0.0" }
		fmt.Printf("%v: \t%v\n", cellVal, c)
		args = append(args, c)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	var ULVx testValues
	ULVx.Name	= name
	ULVx.Expect	= expected
	ULVx.Formula	= formula
	ULVx.ArgsNames	= cells
	ULVx.Args	= args

	t := template.Must(template.New("").Parse(test))
	err = t.Execute(os.Stdout, ULVx)
	if err != nil {
		log.Println("executing template:", err)
	}
}

func unique(arr []string) []string {
	result := []string{}
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range arr {
		encountered[arr[v]] = true
	}

	// Place all unique keys from
	// the map into the results array.
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}
