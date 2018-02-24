package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	//"strconv"
	"strings"
	"time"
	//"github.com/gorilla/handlers"
	"database/sql"
	//	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	//"html/template"
)

type Login struct {
	Username   string
	Password   string
	Userid     int
	IsLoggedIn bool
	Result     bool
	ErrMsg     error
}

type Scorebook struct {
	Id     int
	Name   string
	Events []Event
}

type Team struct {
	Id       int
	Location string
	Mascot   string
	Logo     string
}

type Event struct {
	Id        int
	Name      string
	Accountid int
	Location  string
	Hometeam  Team
	Awayteam  Team
	Scorebook Scorebook
}

type Possession struct {
	PossessionId string
	EventId      int
	OffTeam      Team
	DefTeam      Team
	OffPlay      Play
	DefPlay      Play
	OffPlayer    Player
	DefPlayer    Player
	Period       string
	Notes        string
	ResultId 	int
	ResultType   string
	ResultPoints int 
	Attributes   []PossAttributes
}

type PossAttributes struct {
	PossessionId string
	Label        string
	Value        string
}

type Play struct {
	PlayId     int
	PlayName   string
	PlayAbbrev string
	SideOfBall string
	Team       Team
}

type Stat struct {
	Id           int
	Event        Event
	Possessionid int
	Statx        int
	Staty        int
	Stattype     int //offense vs defense
	Player       Player
	//opponentPlayer int
}

type Player struct {
	Id       int
	Name     string
	Number   string
	Position string
	Team     Team
}

type SaveResult struct {
	Success bool
	ErrMsg  error
}

type GenericOption struct {
	Value string
	Label string
}

func corsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		var dat map[string]interface{}
		_ = json.Unmarshal(body, &dat)
		username, _ := dat["username"].(string)
		password, _ := dat["password"].(string)
		sqlLogin := `SELECT userid, count(*) isLoggedIn FROM user WHERE username='` + username + `' AND password='` + password + `' group by userid LIMIT 1;`
		db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
		var loginResult Login
		loginResult.Result = false
		loginResult.IsLoggedIn = false
		if err != nil {
			fmt.Println(w, "DB CONNECT fail [%s]", err)
			loginResult.ErrMsg = err
		}
		rows, err := db.Query(sqlLogin)
		if err != nil {
			fmt.Println(w, "Query fail [%s]", err)
			loginResult.ErrMsg = err
		}

		var isLoggedIn int64
		var userid int
		for rows.Next() {
			err = rows.Scan(&userid, &isLoggedIn)
			fmt.Println("Is Logged in? %s | %s", sqlLogin, (isLoggedIn))
			if err != nil {
				//log.Fatal(err)
			}
		}
		if isLoggedIn == 1 {
			loginResult.Result = true
			loginResult.IsLoggedIn = true
			loginResult.Username = username
			loginResult.Userid = userid
		}
		resJson, _ := json.Marshal(loginResult)
		fmt.Fprintf(w, string(resJson))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	body, _ := ioutil.ReadAll(r.Body)
	var dat map[string]interface{}
	_ = json.Unmarshal(body, &dat)
	username, _ := dat["username"].(string)
	password, _ := dat["password"].(string)
	firstname, _ := dat["firstname"].(string)
	lastname, _ := dat["lastname"].(string)
	email, _ := dat["email"].(string)
	sqlRegister := `INSERT INTO user(username, password, email, firstname, lastname) VALUES('` + string(username) + `','` + string(password) + `','` + string(email) + `','` + string(firstname) + `','` + string(lastname) + `');`
	fmt.Println(sqlRegister)
	var registerResult SaveResult
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		registerResult.Success = false
		registerResult.ErrMsg = err
		fmt.Println(w, "DB CONNECT fail [%s]", err)
		resJson, _ := json.Marshal(registerResult)
		fmt.Fprintf(w, string(resJson))
		return
	} else {
		//	fmt.Println(w, "Successfully connected to dB", err)
	}
	_, err = db.Query(sqlRegister)
	if err != nil {
		registerResult.Success = false
		registerResult.ErrMsg = err
		fmt.Println("Query fail [%s]", err)
		resJson, _ := json.Marshal(registerResult)
		fmt.Fprintf(w, string(resJson))
		return
	} else {
		//	fmt.Println(w, "Query succeeded?", err)
	}
	registerResult.Success = true
	resJson, _ := json.Marshal(registerResult)
	fmt.Fprintf(w, string(resJson))
}

func saveSB(sbBody map[string]interface{}) (result SaveResult, errorMsg error) {
	scorebookname, _ := sbBody["name"].(string)
	userid, _ := sbBody["userid"].(string)
	sqlSaveSB := "INSERT INTO scorebook(name, userid) values('" + scorebookname + "','" + userid + "') ;"
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		result.Success = false
		result.ErrMsg = err
		return result, err
	}
	_, err = db.Query(sqlSaveSB)
	if err != nil {
		result.Success = false
		result.ErrMsg = err
		return result, err
	}
	result.Success = true
	return result, nil
}

func sbHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	fmt.Printf("\n Request Method: (%s)", r.Method)
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		var dat map[string]interface{}
		_ = json.Unmarshal(body, &dat)
		res, _ := saveSB(dat)
		resJson, _ := json.Marshal(res)
		fmt.Fprintf(w, string(resJson))
		return
	}
	var sqlScorebook string
	path_parts := strings.Split(r.URL.Path, "/")
	userid := path_parts[2]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if len(path_parts) > 3 {
		scorebookid := path_parts[3]
		sqlScorebook = `SELECT scorebookid, name FROM scorebook WHERE scorebookid=` + scorebookid + ` limit 1`
	} else {
		sqlScorebook = `SELECT scorebookid, name FROM scorebook WHERE userid=` + userid + ` limit 10`
	}
	rows, err := db.Query(sqlScorebook)
	if err != nil {
		fmt.Println("ERROR: ", err.Error())
		return
	}
	var sbReturn []Scorebook
	for rows.Next() {
		var sb Scorebook
		var scorebookid int
		var name string
		err = rows.Scan(&scorebookid, &name)
		sb.Id, sb.Name = scorebookid, name
		fmt.Println(w, "%v \n sc %v, name: %v", sqlScorebook, scorebookid, name)
		sbReturn = append(sbReturn, sb)
	}
	sbjson, _ := json.Marshal(sbReturn)
	fmt.Fprintf(w, string(sbjson))
}

func savePoss(possBody map[string]interface{}) (result SaveResult, errorMsg error) {
	fmt.Printf("Body: (%+v)", possBody)
	eventId, _ := possBody["eventid"].(string)
	offTeam, _ := possBody["offTeam"].(string)
	offPlay, _ := possBody["offPlay"].(string)
	offPlayer, _ := possBody["offPlayer"].(string)
	defTeam, _ := possBody["defTeam"].(string)
	defPlay, _ := possBody["defPlay"].(string)
	defPlayer, _ := possBody["defPlayer"].(string)
	period, _ := possBody["period"].(string)
	resultid, _ := possBody["result"].(string)
	notes, _ := possBody["notes"].(string)
	unixTime := time.Now().Unix()
	possIdSeed := fmt.Sprintf("%s%s", unixTime, eventId)
	hasher := md5.New()
	hasher.Write([]byte(possIdSeed))
	possessionId := hex.EncodeToString(hasher.Sum(nil))

	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	sqlSavePoss := "INSERT INTO possession(possessionId, eventId, offTeam, offPlay, offPlayer, defTeam, defPlay, defPlayer, period, notes, resultid) values('" + possessionId + "','" + eventId + "','" + offTeam + "','" + offPlay + "','" + offPlayer + "','" + defTeam + "','" + defPlay + "','" + defPlayer + "','" + period + "','" + notes + "','" + resultid + "' ) ;"
	fmt.Printf("\nSQL: (%s)", sqlSavePoss)
	if err != nil {
		result.Success = false
		result.ErrMsg = err
		return result, err
	}
	_, err = db.Query(sqlSavePoss)
	if err != nil {
		result.Success = false
		result.ErrMsg = err
		return result, err
	}
	result.Success = true
	return result, nil
}

func saveEv(evBody map[string]interface{}) (result SaveResult, errorMsg error) {
	fmt.Printf("Body: (%+v)", evBody)
	evName, _ := evBody["name"].(string)
	scorebookid, _ := evBody["scorebookid"].(string)
	evDate, _ := evBody["date"].(string)
	evLocation, _ := evBody["location"].(string)
	evHomeTeamId, _ := evBody["hometeamid"].(string)
	fmt.Println("evBody: " + evBody["hometeamid"].(string))
	fmt.Println("evVar: " + evHomeTeamId)
	evAwayTeamId, _ := evBody["awayteamid"].(string)
	sqlSaveEv := "INSERT INTO event(name, scorebookid, dt, location, hometeamid, awayteamid) values('" + evName + "','" + scorebookid + "','" + evDate + "','" + evLocation + "','" + evHomeTeamId + "','" + evAwayTeamId + "') ;"
	fmt.Printf("\nSQL: (%+v)", sqlSaveEv)
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		result.Success = false
		result.ErrMsg = err
		return result, err
	}
	_, err = db.Query(sqlSaveEv)
	if err != nil {
		result.Success = false
		result.ErrMsg = err
		return result, err
	}
	result.Success = true
	return result, nil
}

func evHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		var dat map[string]interface{}
		_ = json.Unmarshal(body, &dat)
		res, _ := saveEv(dat)
		resJson, _ := json.Marshal(res)
		fmt.Fprintf(w, string(resJson))
		return
	}
	var sqlEvent string
	path_parts := strings.Split(r.URL.Path, "/")
	fmt.Println(w, "\n in evHandler\n %v ", path_parts)
	scorebookid := path_parts[3]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if len(path_parts) > 4 {
		eventid := path_parts[4]
		sqlEvent = `select eventid, event.name, event.location, hometeamid, team_home.location, team_home.mascot, awayteamid, team_away.location, team_away.mascot from event inner join team as team_home on event.hometeamid=team_home.teamid inner join team as team_away on event.awayteamid=team_away.teamid  where eventid=` + eventid + ` limit 1`
	} else {
		sqlEvent = `select eventid, event.name, event.location, hometeamid, team_home.location, team_home.mascot, awayteamid, team_away.location, team_away.mascot from event inner join team as team_home on event.hometeamid=team_home.teamid inner join team as team_away on event.awayteamid=team_away.teamid  where event.scorebookid=` + scorebookid + ` limit 10`
		fmt.Println(w, "\n in evHandler\n %s ", sqlEvent)
	}
	rows, err := db.Query(sqlEvent)
	if err != nil {
		fmt.Println("ERROR: ", err.Error())
		return
	}
	var sbReturn []Event
	//sbReturn = make([]Scorebook, 10)
	for rows.Next() {
		var ev Event
		var eventid int
		var eventname string
		var eventlocation string
		var hometeamid int
		var hometeamloc string
		var hometeammascot string
		var awayteamid int
		var awayteamloc string
		var awayteammascot string
		err = rows.Scan(&eventid, &eventname, &eventlocation, &hometeamid, &hometeamloc, &hometeammascot, &awayteamid, &awayteamloc, &awayteammascot)
		hometeam := Team{Id: hometeamid, Location: hometeamloc, Mascot: hometeammascot}
		awayteam := Team{Id: awayteamid, Location: awayteamloc, Mascot: awayteammascot}
		ev.Id, ev.Name, ev.Location, ev.Hometeam, ev.Awayteam = eventid, eventname, eventlocation, hometeam, awayteam
		sbReturn = append(sbReturn, ev)
	}
	sbjson, _ := json.Marshal(sbReturn)
	fmt.Fprintf(w, string(sbjson))
}

func teamHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	if r.Method == "GET" {
		fmt.Println("Getting Teams")
		var sqlTeam string
		path_parts := strings.Split(r.URL.Path, "/")
		userid := path_parts[2]
		scorebookid := path_parts[3]
		db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
		if err != nil {
			fmt.Println(w, "DB CONNECT fail [%s]", err)
		}
		if len(path_parts) > 4 {
			eventid := path_parts[4]
			sqlTeam = `select teamid, location, mascot, logo FROM team  where (eventid=` + eventid + `) limit 3`
		} else {
			sqlTeam = `select teamid, location, mascot, logo FROM team  where (userid=` + userid + `) AND (scorebookid=` + scorebookid + `) `
		}
		fmt.Println("SQL TEAM: ", sqlTeam)
		rows, err := db.Query(sqlTeam)
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			return
		}
		var sbReturn []Team
		for rows.Next() {
			var teamid int
			var location string
			var mascot string
			var logo string
			err = rows.Scan(&teamid, &location, &mascot, &logo)
			tm := Team{Id: teamid, Location: location, Mascot: mascot}
			sbReturn = append(sbReturn, tm)
		}
		sbjson, _ := json.Marshal(sbReturn)
		fmt.Fprintf(w, string(sbjson))
	}
	if r.Method == "POST" {

	}
}

func statTypesHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	var sqlStatTypes string
	path_parts := strings.Split(r.URL.Path, "/")
	userid := path_parts[2]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if r.Method == "GET" {
		if len(path_parts) > 1 {
			sqlStatTypes = `select statTypeId, statType from statTypes where userid=` + userid + ` ;`
			rows, err := db.Query(sqlStatTypes)
			if err != nil {
				fmt.Println("ERROR: ", err.Error())
				return
			}
			var statOptions []GenericOption
			for rows.Next() {
				var stattypeid string
				var stattype string
				err = rows.Scan(&stattypeid, &stattype)
				st := GenericOption{Value: stattypeid, Label: stattype}
				statOptions = append(statOptions, st)
			}
			sbjson, _ := json.Marshal(statOptions)
			fmt.Fprintf(w, string(sbjson))
		} else {
			//DEFAULT
		}
	}
}

func resultTypesHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	var sqlResultTypes string
	path_parts := strings.Split(r.URL.Path, "/")
	userid := path_parts[2]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if r.Method == "GET" {
		if len(path_parts) > 1 {
			sqlResultTypes = `select resultTypeId, resultType from resultTypes where userid=` + userid + ` ;`
			rows, err := db.Query(sqlResultTypes)
			if err != nil {
				fmt.Println(sqlResultTypes, "| ERROR: ", err.Error())
				return
			}
			var resultOptions []GenericOption
			for rows.Next() {
				var resulttypeid string
				var resulttype string
				err = rows.Scan(&resulttypeid, &resulttype)
				rt := GenericOption{Value: resulttypeid, Label: resulttype}
				resultOptions = append(resultOptions, rt)
			}
			sbjson, _ := json.Marshal(resultOptions)
			fmt.Fprintf(w, string(sbjson))
		} else {
			//DEFULAT OPTIONS
		}
	}
}

func playerHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	var sqlPlayers string
	path_parts := strings.Split(r.URL.Path, "/")
	teamid := path_parts[2]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if r.Method == "GET" {
		if len(path_parts) > 1 {
			sqlPlayers = `select playerId, CONCAT(COALESCE(jerseyNumber,'X'), ' - ', firstname,' ',lastname) playerLabel from player where teamid=` + teamid + ` ;`
			rows, err := db.Query(sqlPlayers)
			if err != nil {
				fmt.Println(sqlPlayers, "| ERROR: ", err.Error())
				return
			}
			var resultOptions []GenericOption
			for rows.Next() {
				var id string
				var name string
				err = rows.Scan(&id, &name)
				rt := GenericOption{Value: id, Label: name}
				resultOptions = append(resultOptions, rt)
			}
			sbjson, _ := json.Marshal(resultOptions)
			fmt.Fprintf(w, string(sbjson))
		} else {
			//DEFULAT OPTIONS
		}
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	var sqlPlay string
	path_parts := strings.Split(r.URL.Path, "/")
	userid := path_parts[2]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if r.Method == "GET" {
		if len(path_parts) > 1 {
			sqlPlay = `select playId, playName from play where userid=` + userid + ` ;`
			rows, err := db.Query(sqlPlay)
			if err != nil {
				fmt.Println(sqlPlay, "| ERROR: ", err.Error())
				return
			}
			var resultOptions []GenericOption
			for rows.Next() {
				var id string
				var name string
				err = rows.Scan(&id, &name)
				rt := GenericOption{Value: id, Label: name}
				resultOptions = append(resultOptions, rt)
			}
			sbjson, _ := json.Marshal(resultOptions)
			fmt.Fprintf(w, string(sbjson))
		} else {
			//DEFULAT OPTIONS
		}
	}
}

func possHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, r)
	var sqlPoss string
	path_parts := strings.Split(r.URL.Path, "/")
	eventid := path_parts[3]
	db, err := sql.Open("mysql", "root:Sc0r3b00k?@/scorebook")
	if err != nil {
		fmt.Println(w, "DB CONNECT fail [%s]", err)
	}
	if r.Method == "GET" {
		if len(path_parts) > 4 {
			possessionid := path_parts[4]
			sqlPoss = `select away on event.awayteamid=team_away.teamid  where possessionid=` + possessionid + ` limit 1`
		} else {
			sqlPoss = `
SELECT 
possessionid, period, p.offteam, p.offplay, oplay.playname, offplayer, CONCAT(oplayer.firstname,' ', oplayer.lastname) as offplayername, p.defteam, p.defplay,  defplayer, postfeed, reversal, pieceofpaint, notes, resultid, res.resultType, res.points
FROM
possession as p left outer join player as oplayer on p.offplayer = oplayer.playerid left outer join player as dplayer on p.defplayer=dplayer.playerid left outer join play as oplay on p.offplay = oplay.playid left outer join play as dplay on p.defplay = dplay.playid left outer join resultTypes as res on p.resultid = res.resultTypeId 
WHERE 
eventid='` + eventid + `';`;
		}
		fmt.Printf("\nPath Parts: [%+v]", path_parts)
		fmt.Println("\nSQL: ", sqlPoss)
		rows, err := db.Query(sqlPoss)
		if err != nil {
			fmt.Printf("\nPath Parts: [%+v]", path_parts)
			fmt.Println("\nSQL: ", sqlPoss, " ERROR: ", err.Error())
			return
		}
		var possReturn []Possession
		for rows.Next() {
			var possid string
			var period string
			var offteamid int
			var offplayid int
			var offplayname string
			var offplayerid int
			var offplayername string
			var defteamid int
			var defplayid int
			var defplayerid int
			var postfeed sql.NullString
			var reversal sql.NullString
			var pieceofpaint sql.NullString
			var notes sql.NullString
			var resultid int
			var resulttype sql.NullString
			var resultpoints int

			/*
			var offteamlocation string
			var offteammascot string
			var defteamlocation string
			var defteammascot string
			*/
			// Attributes []PossAttributes
			err = rows.Scan(&possid, &period, &offteamid, &offplayid, &offplayname, &offplayerid, &offplayername, &defteamid, &defplayid, &defplayerid, &postfeed, &reversal, &pieceofpaint, &notes, &resultid, &resulttype, &resultpoints)
			if err != nil {
				fmt.Printf("ERROR: %s", err.Error())
			}
			var real_resulttype string
			if resulttype.Valid {
				real_resulttype = resulttype.String 
			}
			var real_notes string
			if notes.Valid {
				real_notes = notes.String
			}
fmt.Printf("RES: [%d] [%s]", resultid, resulttype)
			oTeam := Team{Id: offteamid}
			dTeam := Team{Id: defteamid}
			oPlay := Play{PlayId: offplayid, PlayName: offplayname}
			dPlay := Play{PlayId: defplayid }
			oPlayer := Player{Id: offplayerid, Name: offplayername }
			dPlayer := Player{Id: defplayerid }
			poss := Possession{PossessionId: possid, OffTeam: oTeam, OffPlay: oPlay, OffPlayer: oPlayer, DefTeam: dTeam, DefPlay: dPlay, DefPlayer: dPlayer, Period: period, Notes: real_notes, ResultId: resultid, ResultType: real_resulttype, ResultPoints: resultpoints}
			fmt.Printf("\nPOSSESSION: [%+v]", poss)
			possReturn = append(possReturn, poss)
		}
		possjson, _ := json.Marshal(possReturn)
		fmt.Fprintf(w, string(possjson))
		return
	}
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		var dat map[string]interface{}
		_ = json.Unmarshal(body, &dat)
		res, _ := savePoss(dat)
		resJson, _ := json.Marshal(res)
		fmt.Fprintf(w, string(resJson))
		return
	}
}

func main() {
	scRouter := mux.NewRouter()
	//scRouter := mux.NewRouter().StrictSlash(true)
	scRouter.HandleFunc("/login", loginHandler)
	scRouter.HandleFunc("/register", registerHandler)
	scRouter.HandleFunc("/scorebooks/{userid}/{scorebookid}", sbHandler)
	scRouter.HandleFunc("/scorebooks/{userid}", sbHandler)
	scRouter.HandleFunc("/events/{scorebookid}/{eventid}", evHandler)
	scRouter.HandleFunc("/events/{scorebookid}", evHandler)
	scRouter.HandleFunc("/possessions/{eventid}/{possessionid}", possHandler)
	scRouter.HandleFunc("/possessions/{eventid}", possHandler)
	scRouter.HandleFunc("/teams/{userid}/{scorebookid}/{eventid}", teamHandler)
	scRouter.HandleFunc("/players/{teamid}", playerHandler)
	scRouter.HandleFunc("/plays/{userid}", playHandler)
	scRouter.HandleFunc("/resulttypes/{userid}", resultTypesHandler)
	scRouter.HandleFunc("/stattypes/{userid}", statTypesHandler)
	http.ListenAndServe(":8080", scRouter)
	//http.ListenAndServe(":8080", handlers.CORS()(scRouter))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
