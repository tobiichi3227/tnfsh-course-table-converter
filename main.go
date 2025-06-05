package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/extrame/xls"
)

//go:embed templates/*
//go:embed static/*
var templateFiles embed.FS

type (
	Class struct {
		ClassID string
		Grade   int
		Number  int
		Courses [6][9]map[string][]string
	}
	Teacher struct {
		TeacherID string
		Name      string
		Courses   [6][9]map[string][]string
	}
	Timeinfo struct {
		Number    string
		StartTime string
		EndTime   string
	}
)

type (
	Teachers map[string]*Teacher
	Classes  map[string]*Class
)

var (
	NUMBER_CHINESE [4]string    = [4]string{"", "一", "二", "三"}
	TIMEINFO       [10]Timeinfo = [10]Timeinfo{
		{},
		{"一", "0800", "0850"},
		{"二", "0900", "0950"},
		{"三", "1010", "1100"},
		{"四", "1110", "1200"},
		{},
		{"五", "1310", "1400"},
		{"六", "1410", "1500"},
		{"七", "1510", "1600"},
		{"八", "1610", "1700"},
	}
	TH [10]int = [10]int{-1, 1, 2, 3, 4, -1, 5, 6, 7, 8}
)

func convert(reader io.ReadSeeker) (*bytes.Buffer, error) {
	wb, err := xls.OpenReader(reader, "utf-8")
	if err != nil {
		return nil, err
	}

	classes := make(Classes)
	teachers := make(Teachers)
	{
		t := new(Teacher)
		t.Name = ""
		t.TeacherID = "empty"
		teachers[t.TeacherID] = t
	}
	classNumToClass := make(map[int]*Class)

	ws := wb.GetSheet(0)
	for rowNum := range int(ws.MaxRow) {
		row := ws.Row(rowNum)
		classID := row.Col(0)
		classNum, err := strconv.Atoi(row.Col(1))
		if err != nil {
			continue
		}
		c, ok := classes[classID]
		if !ok {
			c = new(Class)
			c.ClassID = classID
			c.Grade = (classNum / 100)
			c.Number = (classNum % 100)
			classes[classID] = c
			for i := range len(c.Courses) {
				for j := range len(c.Courses[i]) {
					c.Courses[i][j] = make(map[string][]string)
				}
			}
			classNumToClass[classNum] = c
		}

		teacherID := row.Col(9)
		t, ok := teachers[teacherID]
		if !ok {
			t = new(Teacher)
			t.Name = row.Col(10)
			t.TeacherID = teacherID
			teachers[teacherID] = t
			for i := range len(t.Courses) {
				for j := range len(t.Courses[i]) {
					t.Courses[i][j] = make(map[string][]string)
				}
			}
		}

		week, err := strconv.Atoi(row.Col(2))
		if err != nil {
			continue
		}
		th, err := strconv.Atoi(row.Col(3))
		if err != nil {
			continue
		}

		courseName := row.Col(5)
		if teacherID == "" {
			teacherID = "empty"
		}

		c.Courses[week][th][courseName] = append(c.Courses[week][th][courseName], teacherID)
		slices.Sort(c.Courses[week][th][courseName])

		t.Courses[week][th][courseName] = append(t.Courses[week][th][courseName], classID)
		slices.Sort(t.Courses[week][th][courseName])
	}

	funcMap := template.FuncMap{
		"Find": strings.IndexAny,
		"Add": func(a, b int) int {
			return a + b
		},
		"Mul": func(a, b int) int {
			return a * b
		},
		"Mod": func(a, b int) int {
			if b == 0 {
				panic("b should not equal to 0.")
			}
			return a % b
		},
		"EqInt": func(a, b int) bool {
			return a == b
		},
		"Utf8Len": utf8.RuneCountInString,
		"ToString": func(ch byte) string {
			return string(ch)
		},
	}
	ct := template.New("base").Funcs(funcMap)
	ct, err = ct.ParseFS(templateFiles, "templates/golang_class.html", "templates/golang_teacher.html",
		"templates/golang_classindex.html", "templates/golang_teacherindex.html")
	if err != nil {
		return nil, err
	}

	zipBuf := new(bytes.Buffer)
	zip := zip.NewWriter(zipBuf)
	defer zip.Close()

	updateTimestamp := time.Now().Format("2006/01/02 15:04:05")

	for classID, class := range classes {
		f, err := zip.Create(fmt.Sprintf("C%v.HTML", classID))
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = ct.ExecuteTemplate(f, "golang_class.html", struct {
			UpdateTimestamp string
			Class           *Class
			Classes         Classes
			Teachers        Teachers
			TIMEINFO        []Timeinfo
			NUMBER_CHINESE  []string
			TH              []int
		}{
			UpdateTimestamp: updateTimestamp,
			Class:           class,
			Classes:         classes,
			Teachers:        teachers,
			TIMEINFO:        TIMEINFO[:],
			NUMBER_CHINESE:  NUMBER_CHINESE[:],
			TH:              TH[:],
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	for teacherID, teacher := range teachers {
		if teacherID == "empty" {
			continue
		}
		f, err := zip.Create(fmt.Sprintf("T%v.HTML", teacherID))
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = ct.ExecuteTemplate(f, "golang_teacher.html", struct {
			UpdateTimestamp string
			Teacher         *Teacher
			Classes         Classes
			TIMEINFO        []Timeinfo
			NUMBER_CHINESE  []string
			TH              []int
		}{
			UpdateTimestamp: updateTimestamp,
			Teacher:         teacher,
			Classes:         classes,
			TIMEINFO:        TIMEINFO[:],
			NUMBER_CHINESE:  NUMBER_CHINESE[:],
			TH:              TH[:],
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	{
		f, err := zip.Create("_ClassIndex.html")
		if err != nil {
			return nil, err
		}

		err = ct.ExecuteTemplate(f, "golang_classindex.html", struct {
			UpdateTimestamp string
			ClassNumToClass map[int]*Class
			Grades          [3]string
		}{
			UpdateTimestamp: updateTimestamp,
			ClassNumToClass: classNumToClass,
			Grades:          [3]string{"高一", "高二", "高三"},
		})
		if err != nil {
			return nil, err
		}
	}

	{
		f, err := zip.Create("_TeachIndex.html")
		if err != nil {
			return nil, err
		}

		type Subject struct {
			Name     string
			Range    []string
			Teachers []string
		}
		subjects := [7]Subject{
			{"國文科", []string{"A"}, []string{}},
			{"英文科", []string{"B"}, []string{}},
			{"數學科", []string{"C"}, []string{}},
			{"社會科", []string{"D", "E", "F"}, []string{}},
			{"自然科", []string{"G", "H", "I"}, []string{}},
			{"藝能科", []string{"J", "K", "L"}, []string{}},
			{"外聘教師", []string{}, []string{}},
		}
		rs := map[string]*Subject{
			"A": &subjects[0],
			"B": &subjects[1],
			"C": &subjects[2],
			"D": &subjects[3],
			"E": &subjects[3],
			"F": &subjects[3],
			"G": &subjects[4],
			"H": &subjects[4],
			"I": &subjects[4],
			"J": &subjects[5],
			"K": &subjects[5],
			"L": &subjects[5],
		}

		delete(teachers, "empty")
		keys := make([]string, 0, len(teachers))
		for teacherID := range teachers {
			keys = append(keys, teacherID)
		}
		slices.Sort(keys)
		for _, teacherID := range keys {
			if teacherID == "" {
				continue
			}

			group := string(teacherID[0])
			val, ok := rs[group]
			if ok {
				val.Teachers = append(val.Teachers, teacherID)
			} else {
				subjects[6].Teachers = append(subjects[6].Teachers, teacherID)
			}
		}

		err = ct.ExecuteTemplate(f, "golang_teacherindex.html", struct {
			UpdateTimestamp string
			Teachers        Teachers
			Subjects        []Subject
		}{
			UpdateTimestamp: updateTimestamp,
			Teachers:        teachers,
			Subjects:        subjects[:],
		})
		if err != nil {
			return nil, err
		}
	}

	err = zip.Close()
	if err != nil {
		return nil, err
	}
	return zipBuf, nil
}

func openBrowser(url string) {
	cmd := exec.Command("xdg-open", url)
	if err := cmd.Start(); err != nil {
		cmd = exec.Command("open", url) // macOS
		if err := cmd.Start(); err != nil {
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) // Windows
			cmd.Start()
		}
	}
}

var indexTmpl = template.Must(template.ParseFS(templateFiles, "templates/index.html"))

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := indexTmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		buf, err := convert(file)
		if err != nil {
			log.Println("Convert error:", err)
			http.Error(w, "Conversion failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=output.zip")
		w.Header().Set("Content-Type", "application/zip")
		io.Copy(w, buf)
	}
}

func main() {
	fs := http.FileServerFS(templateFiles)
	http.Handle("/static/", fs)
	http.HandleFunc("/", handler)
	go openBrowser("http://localhost:5000")
	fmt.Println("Server started at http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
