package main

import (
		"os"
    "log"
    "time"
    "fmt"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"
    
    "github.com/sclevine/agouti"
    "github.com/PuerkitoBio/goquery"
    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
)

type Schedule struct {
	Name  string
	Start time.Time
	End   time.Time
}

var schedule [69][]Schedule
var defRoom [69]Schedule

func init() {
	t := time.Now()
	for i, _ := range defRoom {
		var s_hour int
		var s_min int
		var e_hour int
		var e_min int
		if i < 26 || (29 < i && i < 32) || (i == 37) || (39 < i && i < 102) {
			s_hour = 8; s_min = 30; e_hour = 24; e_min = 0
		} else if (25 < i && i < 28) || (i == 29) {
			s_hour = 8; s_min = 30; e_hour = 17; e_min = 0
		} else if (i == 28) || (i == 32) || (i == 38) {
			s_hour = 0; s_min = 0; e_hour = 24; e_min = 0
		} else if (i == 29) || (33 < i && i < 37) ||i == 39 {
			s_hour = 8; s_min = 30; e_hour = 18; e_min = 0
		} else if (i == 33) {
			s_hour = 8; s_min = 30; e_hour = 20; e_min = 30
		}
		defRoom[i].Start = time.Date(t.Year(), t.Month(), t.Day(), s_hour, s_min, 0, 0, time.Local)
		defRoom[i].End = time.Date(t.Year(), t.Month(), t.Day(), e_hour, e_min, 0, 0, time.Local)
	}
	//fmt.Println(defRoom)
}

func getSchedule() {
	driver := agouti.ChromeDriver(
				    	agouti.ChromeOptions("args", []string {
				    		"--headless",
				    	}),
				    	agouti.Debug,
		    		)
  if err := driver.Start(); err != nil {
      log.Fatalf("Failed to start driver:%v", err)
  }
  defer driver.Stop()
  

  page, err := driver.NewPage(agouti.Browser("chrome"))
  if err != nil {
      log.Fatalf("Failed to open page:%v", err)
  }

  if err := page.Navigate("https://csweb.u-aizu.ac.jp/campusweb/campusportal.do"); err != nil {
      log.Fatalf("Failed to navigate:%v", err)
  }
  // ID, Passの要素を取得し、値を設定
  identity := page.FindByID("LoginFormSimple").FindByName("userName")
  password := page.FindByID("LoginFormSimple").FindByName("password")
  identity.Fill("s1240236")
  password.Fill("kanta01mtyan")
  // formをサブミット
  if err := page.FindByID("wf_PTW0000011_20120827233559-form").Submit(); err != nil {
      log.Fatalf("Failed to login:%v", err)
  }
  time.Sleep(1 * time.Second)
  if err := page.Navigate("https://csweb.u-aizu.ac.jp/campusweb/campussquare.do?_flowId=KHW0001300-flow"); err != nil {
      log.Fatalf("Failed to navigate:%v", err)
  }
  curContentsDom, err := page.HTML()
  if err != nil {
    log.Printf("Failed to get html: %v", err)
  }
  readerCurContents := strings.NewReader(curContentsDom)
	contentsDom, _ := goquery.NewDocumentFromReader(readerCurContents) // ページ内容が変化しているので再取得
	contentsDom.Find("table.kyuko-shisetsu > tbody").Each(func(_ int, s *goquery.Selection) {  // 繰り返し取得したい部分までのセレクタを入れる
		i := 0
		k := 0
		l := 36
	  s.Find("tr").Each(func(_ int, se *goquery.Selection) {
	  	if (2 < i && i < 23) || (25 < i && i < 46) || (48 < i && i < 69) || (71 < i) {
		  	se.Find("td").Each(func(_ int, c *goquery.Selection) {
		  		class, _ := c.Attr("class")
		  		if class != "kyuko-shi-shisetsunm" {
		  			col, status := c.Attr("colspan")
		  			j, _ := strconv.Atoi(col)
		  			if status {
		  				t := time.Now()
		  				var plane Schedule
		  				plane.Name = c.Text()
		  				plane.Start = time.Date(t.Year(), t.Month(), t.Day(), l/6, l%6*10, 0, 0, time.Local)
		  				plane.End = time.Date(t.Year(), t.Month(), t.Day(), (l+j)/6, (l+j)%6*10, 0, 0, time.Local)
			  			schedule[k] = append(schedule[k], plane)
		  			}
		  			l++
		  		}
		  	})
		  	k++
		  	l=36
	  	}
	  	i++
	  })
	})
}

func main() {
	port := os.Getenv("PORT")
	getSchedule()
  r := mux.NewRouter()
  r.HandleFunc("/api/UoAizu/room/{id}", roomStatus).Methods("GET")
  routerWithCORS := handlers.CORS(
        handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "PATCH"}),
        handlers.AllowedOrigins([]string{"*"}),
                handlers.AllowedHeaders([]string{"Content-Type", "application/json", ""}),
    )(r)
  log.Fatal(http.ListenAndServe(port, routerWithCORS))
	fmt.Println(getRoomStatus(47))
}

func roomStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	i := mux.Vars(r)
	id, _ := strconv.Atoi(i["id"]) 
	response := getRoomStatus(id)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func getRoomStatus(id int) (status bool) {
	// now := time.Date(2018, 12, 2, 12, 0, 0, 0, time.Local)
	now := time.Now()
	if !(defRoom[id].Start).After(now) && !now.After(defRoom[id].End) {
		if len(schedule[id]) != 0 {
			for _, s := range schedule[id] {
				if !(!(s.Start).After(now) && !now.After(s.End)) { 
					status = true 
				} else {
					return false
				}
			}
		} else {
			return true
		}
	} else {
		return false
	}
	return
}