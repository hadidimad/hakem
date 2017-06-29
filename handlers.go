package main

import (
	"crypto/sha512"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"strconv"

	"encoding/json"

	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
)

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	cookie, err := getCookie("login_cookie", r)
	if err == nil {
		usrIDstr := cookie
		usrID, _ := strconv.Atoi(usrIDstr)
		onuser := getUserByID(usrID)
		m["user"] = onuser
	}
	Render(w, "index", m)
}

func signupGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	_, err1 := getCookie("login_cookie", r)
	if err1 == nil {
		http.Redirect(w, r, "/user", http.StatusFound)
	}
	err := r.URL.Query().Get("err")
	if err == "passnotmatch" {
		m["passnotmatch"] = true
	}
	if err == "takenusername" {
		m["takenusername"] = true
	}
	if err == "takenemail" {
		m["takenemail"] = true
	}
	if err == "emptyfield" {
		m["emptyfield"] = true
	}
	if err == "invalidUsername" {
		m["invalidUsername"] = true
	}

	Render(w, "signup", m)
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	var formIsValid bool
	formIsValid = true
	if r.FormValue("password") == "" || r.FormValue("username") == "" || r.FormValue("email") == "" {
		formIsValid = false
		http.Redirect(w, r, "/signup?err=emptyfield", http.StatusFound)
	}
	if strings.ContainsAny(r.FormValue("username"), "/=*'><") {
		formIsValid = false
		http.Redirect(w, r, "/signup?err=invalidUsername", http.StatusFound)
	}
	if !(r.FormValue("password") == r.FormValue("password-repeat")) {
		formIsValid = false
		http.Redirect(w, r, "/signup?err=passnotmatch", http.StatusFound)
	}
	db, _ := sql.Open("sqlite3", "./database/database.db")
	if formIsValid {
		rows, _ := db.Query("SELECT * FROM userinfo WHERE username='" + r.FormValue("username") + "';")
		for rows.Next() {
			formIsValid = false
			http.Redirect(w, r, "/signup?err=takenusername", http.StatusFound)
		}
	}
	if formIsValid {
		rows, _ := db.Query("SELECT * FROM userinfo WHERE email='" + r.FormValue("email") + "';")
		for rows.Next() {
			formIsValid = false
			http.Redirect(w, r, "/signup?err=takenemail", http.StatusFound)
		}
	}
	if formIsValid {
		passHash := sha512.New()
		io.WriteString(passHash, r.FormValue("password"))

		stmt, _ := db.Prepare("INSERT INTO userinfo(username,password,email,gameswin,gameslose) values(?,?,?,?,?)")
		stmt.Exec(r.FormValue("username"), string(passHash.Sum(nil)), r.FormValue("email"), 0, 0)
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("Image")
		if err == nil {
			defer file.Close()
			f, err := os.OpenFile("./userImages/"+r.FormValue("username"), os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			io.Copy(f, file)
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	if r.URL.Query().Get("err") == "invalid" {
		m["invalid"] = true
	}
	Render(w, "login", m)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	var onuser user
	var finded bool
	finded = false
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM userinfo WHERE username='" + r.FormValue("username") + "'")
	for rows.Next() {
		rows.Scan(&onuser.ID, &onuser.Username, &onuser.Password, &onuser.Email, &onuser.GamesWin, &onuser.GamesLose)
		passHash := sha512.New()
		io.WriteString(passHash, r.FormValue("password"))
		if onuser.Username == r.FormValue("username") && onuser.Password == string(passHash.Sum(nil)) {
			finded = true
			break
		}
	}
	rows.Close()
	if finded {
		setCookie("login_cookie", strconv.Itoa(onuser.ID), w)
		m["username"] = onuser.Username
		fmt.Println("you are in")
		http.Redirect(w, r, "/user", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login?err=invalid", http.StatusFound)
	}
}

func userPageGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	cookie, err := getCookie("login_cookie", r)
	var usrID int
	if err == nil {
		usrIDstr := cookie
		usrID, _ = strconv.Atoi(usrIDstr)
		onuser := getUserByID(usrID)
		m["user"] = onuser
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM games WHERE side1ID=" + strconv.Itoa(usrID) + " OR side2ID=" + strconv.Itoa(usrID) + ";")
	var games []showGames
	var tempGame game
	var tempShowGame showGames
	var gameID int
	for rows.Next() {
		rows.Scan(&gameID, &tempGame.Side1ID, &tempGame.Side2ID)
		file, _ := ioutil.ReadFile("./games/" + strconv.Itoa(gameID) + "/game.json")
		json.Unmarshal(file, &tempGame)
		tempShowGame.Side1Username = getUserByID(tempGame.Side1ID).Username
		tempShowGame.Side2Username = getUserByID(tempGame.Side2ID).Username
		tempShowGame.Wins1 = tempGame.Wins1
		tempShowGame.Wins2 = tempGame.Wins2
		tempShowGame.Finished = tempGame.Finished
		tempShowGame.GameID = gameID
		games = append(games, tempShowGame)
	}
	m["games"] = games
	Render(w, "user", m)

}

func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("login_cookie")
	if err == nil {
		cookie.Value = "delete"
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func gameMakerGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	cookie, err := getCookie("login_cookie", r)
	if err == nil {
		usrIDstr := cookie
		usrID, _ := strconv.Atoi(usrIDstr)
		onuser := getUserByID(usrID)
		m["user"] = onuser
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	Render(w, "gameMaker", m)
}

func getUsersGetHandler(w http.ResponseWriter, r *http.Request) {
	userreq := r.URL.Query().Get("user")
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM userinfo WHERE username LIKE '" + userreq + "%';")
	var users []user
	var usr user
	for rows.Next() {
		rows.Scan(&usr.ID, &usr.Username, &usr.Password, &usr.Email, &usr.GamesWin, &usr.GamesLose)
		usr.ID = -3
		usr.Password = ""
		users = append(users, usr)
	}
	jsonUsers, _ := json.Marshal(users)
	fmt.Fprint(w, string(jsonUsers))
}

func startgameGetHanlder(w http.ResponseWriter, r *http.Request) {
	//m := make(map[string]interface{})
	cookie, err := getCookie("login_cookie", r)
	var side1ID int
	if err == nil {
		usrIDstr := cookie
		usrID, _ := strconv.Atoi(usrIDstr)
		side1ID = usrID
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	side2username := r.URL.Query().Get("side2")
	db, _ := sql.Open("sqlite3", "./database/database.db")
	if strings.ContainsAny(side2username, "OR<>''=") {
		http.Redirect(w, r, "/gamemaker", http.StatusFound)
		return
	}
	rows, _ := db.Query("SELECT * FROM userinfo WHERE username='" + side2username + "';")
	var user2 user
	for rows.Next() {
		rows.Scan(&user2.ID, &user2.Username, &user2.Password, &user2.Email, &user2.GamesWin, &user2.GamesLose)
	}
	rows, _ = db.Query("SELECT * FROM userinfo WHERE uid=" + strconv.Itoa(side1ID) + ";")
	var user1 user
	for rows.Next() {
		rows.Scan(&user1.ID, &user1.Username, &user1.Password, &user1.Email, &user1.GamesWin, &user1.GamesLose)
	}
	if user1.Username == user2.Username {
		http.Redirect(w, r, "/gamemaker", http.StatusFound)
		return
	}
	rows, _ = db.Query("SELECT * FROM games WHERE side1ID=" + strconv.Itoa(user1.ID) + " AND side2ID=" + strconv.Itoa(user2.ID) + ";")
	for rows.Next() {
		http.Redirect(w, r, "/gamemaker", http.StatusFound)
		return
	}
	stmt, _ := db.Prepare("INSERT INTO games(side1ID,side2ID) values(?,?)")
	res, _ := stmt.Exec(user1.ID, user2.ID)
	gameID, _ := res.LastInsertId()
	os.MkdirAll("./games/"+strconv.Itoa(int(gameID)), 0777)
	var tempgame game
	tempgame.Side1ID = user1.ID
	tempgame.Side2ID = user2.ID
	tempgame.Turn = 1
	tempgame.Hokm = 1
	tempgame.Wins1 = 0
	tempgame.Wins2 = 0
	tempgame.GameStatus = 2
	allCards := make(map[int]card)
	var k int
	for i := 1; i <= 4; i++ {
		for j := 1; j <= 13; j++ {
			allCards[k] = card{Mode: i, Value: j}
			k++
		}
	}
	for i := 0; i < 13; {
		temp := rand.Intn(52)
		if allCards[temp].Mode != 0 {
			tempgame.Side1Cards = append(tempgame.Side1Cards, allCards[temp])
			allCards[temp] = card{Mode: 0, Value: 0}
			i++
		}
	}
	for i := 0; i < 13; {
		temp := rand.Intn(52)
		if allCards[temp].Mode != 0 {
			tempgame.Side2Cards = append(tempgame.Side2Cards, allCards[temp])
			allCards[temp] = card{Mode: 0, Value: 0}
			i++
		}
	}
	jsonVal, _ := json.Marshal(tempgame)
	ioutil.WriteFile("./games/"+strconv.Itoa(int(gameID))+"/game.json", jsonVal, 0644)
	http.Redirect(w, r, "/games?game="+strconv.Itoa(int(gameID)), http.StatusFound)
}

func deleteGameGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	var onuser user
	cookie, err := getCookie("login_cookie", r)
	if err == nil {
		usrIDstr := cookie
		usrID, _ := strconv.Atoi(usrIDstr)
		onuser = getUserByID(usrID)
		m["user"] = onuser
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	gameIDstr := r.URL.Query().Get("game")
	if strings.ContainsAny(gameIDstr, "></*=OR ") || gameIDstr == "" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM games WHERE id=" + gameIDstr + ";")
	var tempGame game
	var gameID int
	var finded bool
	for rows.Next() {
		rows.Scan(&gameID, &tempGame.Side1ID, &tempGame.Side2ID)
		finded = true
	}
	if !finded {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	if onuser.ID == tempGame.Side1ID || onuser.ID == tempGame.Side2ID {
		onuser.GamesLose++
		stmt, _ := db.Prepare("UPDATE userinfo SET gameslose=? WHERE uid=?")
		stmt.Exec(onuser.GamesLose, onuser.ID)
		stmt, _ = db.Prepare("DELETE FROM games WHERE id=?")
		stmt.Exec(gameID)
		stmt.Close()
		os.RemoveAll("./games/" + gameIDstr)
		http.Redirect(w, r, "/user", http.StatusFound)
	} else {
		http.Redirect(w, r, "/user", http.StatusFound)
	}
}

func gameGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	var onuser user
	cookie, err := getCookie("login_cookie", r)
	if err == nil {
		usrIDstr := cookie
		usrID, _ := strconv.Atoi(usrIDstr)
		onuser = getUserByID(usrID)
		m["user"] = onuser
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	var thisGame game
	gameIDstr := r.URL.Query().Get("game")
	fmt.Println(gameIDstr)
	if strings.ContainsAny(gameIDstr, "></*=OR ") || gameIDstr == "" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	var finded bool
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM games WHERE id=" + gameIDstr)
	for rows.Next() {
		rows.Scan(nil, &thisGame.Side1ID, &thisGame.Side2ID)
		finded = true
	}
	if !finded {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	JsonFile, err := ioutil.ReadFile("./games/" + gameIDstr + "/game.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(JsonFile, &thisGame)
	if !(onuser.ID == thisGame.Side1ID || onuser.ID == thisGame.Side2ID) {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	m["GameID"] = gameIDstr
	m["UserNameSide1"] = getUserByID(thisGame.Side1ID).Username
	m["UserNameSide2"] = getUserByID(thisGame.Side2ID).Username
	m["Wins1"] = thisGame.Wins1
	m["Wins2"] = thisGame.Wins2
	if onuser.ID == thisGame.Side1ID {
		m["YourCards"] = thisGame.Side1Cards
	}
	if onuser.ID == thisGame.Side2ID {
		m["YourCards"] = thisGame.Side2Cards
	}
	if thisGame.GameStatus == 0 {

	} else if thisGame.GameStatus == 1 {

	} else if thisGame.GameStatus == 2 {
		if thisGame.Wins1 == 7 {
			thisGame.Finished = true
			jsonVal, err := json.Marshal(thisGame)
			if err != nil {
				fmt.Println(err)
			}
			ioutil.WriteFile("./games/"+gameIDstr+"/game.json", jsonVal, 0644)
			m["info"] = "game finished " + getUserByID(thisGame.Side1ID).Username + " wins game"
			Render(w, "game", m)
			return
		} else if thisGame.Wins2 == 7 {
			thisGame.Finished = true
			jsonVal, err := json.Marshal(thisGame)
			if err != nil {
				fmt.Println(err)
			}
			ioutil.WriteFile("./games/"+gameIDstr+"/game.json", jsonVal, 0644)
			m["info"] = "game finished " + getUserByID(thisGame.Side2ID).Username + " wins game"
			Render(w, "game", m)
			return
		}
		SendedCardStr := r.URL.Query().Get("card")
		if SendedCardStr != "" {
			if thisGame.Turn == 1 && onuser.ID == thisGame.Side1ID {
				if SendedCardStr == "" {
					m["info"] = "please choose a card"
					Render(w, "game", m)
					return
				}
				var thisCard card
				json.Unmarshal([]byte(SendedCardStr), &thisCard)
				fmt.Println("this Card:", thisCard)
				var havingCard bool
				for _, i := range thisGame.Side1Cards {
					if i == thisCard {
						havingCard = true
						break
					}
				}
				if !havingCard {
					m["info"] = "you dont have this card"
					Render(w, "game", m)
					return
				} ////////////////////
				if thisGame.Card2.Mode != 0 {
					if thisCard.Mode != thisGame.Card2.Mode {
						var finded bool
						for _, i := range thisGame.Side1Cards {
							if i.Mode == thisGame.Card2.Mode {
								finded = true
								break
							}
						}
						if finded {
							m["info"] = "you have this card type in your cards you cant place another type when you have it"
							Render(w, "game", m)
							return
						}
					}
				}
				thisGame.Card1 = thisCard
				if thisGame.Card2.Mode != 0 {
					var Card1Index int
					var Card2Index int
					for i := 0; i < len(thisGame.Side1Cards); i++ {
						if thisGame.Side1Cards[i] == thisGame.Card1 {
							Card1Index = i
						}
					}
					for i := 0; i < len(thisGame.Side2Cards); i++ {
						if thisGame.Side2Cards[i] == thisGame.Card2 {
							Card2Index = i
						}
					}
					if thisGame.Card1.Mode == thisGame.Card2.Mode {
						if thisGame.Card1.Value > thisGame.Card2.Value {
							thisGame.Wins1++
							thisGame.Turn = 1
						} else {
							thisGame.Wins2++
							thisGame.Turn = 2
						}
					} else if thisGame.Card1.Mode == thisGame.Hokm {
						thisGame.Wins1++
						thisGame.Turn = 1
					} else if thisGame.Card2.Mode == thisGame.Hokm {
						thisGame.Wins2++
						thisGame.Turn = 2
					} else {
						thisGame.Wins2++
						thisGame.Turn = 2
					}
					thisGame.Side1Cards = deleteCard(thisGame.Side1Cards, Card1Index)
					thisGame.Side2Cards = deleteCard(thisGame.Side2Cards, Card2Index)
					thisGame.Card1.Mode = 0
					thisGame.Card1.Value = 0
					thisGame.Card2.Mode = 0
					thisGame.Card2.Value = 0
				} else {
					thisGame.Turn = 2
				}
				////////////////////
				jsonVal, err := json.Marshal(thisGame)
				if err != nil {
					fmt.Println(err)
				}
				ioutil.WriteFile("./games/"+gameIDstr+"/game.json", jsonVal, 0644)
			} else if thisGame.Turn == 2 && onuser.ID == thisGame.Side2ID {
				if SendedCardStr == "" {
					m["info"] = "please choose a card"
					Render(w, "game", m)
					return
				}
				var thisCard card
				json.Unmarshal([]byte(SendedCardStr), &thisCard)
				fmt.Println("this Card:", thisCard)
				var havingCard bool
				for _, i := range thisGame.Side2Cards {
					if i == thisCard {
						havingCard = true
						break
					}
				}
				if !havingCard {
					m["info"] = "you dont have this card"
					Render(w, "game", m)
					return
				}
				if thisGame.Card1.Mode != 0 {
					if thisCard.Mode != thisGame.Card1.Mode {
						var finded bool
						for _, i := range thisGame.Side2Cards {
							if i.Mode == thisGame.Card1.Mode {
								finded = true
								break
							}
						}
						if finded {
							m["info"] = "you have this card type in your cards you cant place another type when you have it"
							Render(w, "game", m)
							return
						}
					}
				}
				thisGame.Card2 = thisCard
				if thisGame.Card1.Mode != 0 {
					var Card1Index int
					var Card2Index int
					for i := 0; i < len(thisGame.Side1Cards); i++ {
						if thisGame.Side1Cards[i] == thisGame.Card1 {
							Card1Index = i
						}
					}
					for i := 0; i < len(thisGame.Side2Cards); i++ {
						if thisGame.Side2Cards[i] == thisGame.Card2 {
							Card2Index = i
						}
					}
					if thisGame.Card1.Mode == thisGame.Card2.Mode {
						if thisGame.Card1.Value == 1 {
							thisGame.Wins1++
							thisGame.Turn = 1
						} else if thisGame.Card1.Value == 1 {
							thisGame.Wins2++
							thisGame.Turn = 2
						} else if thisGame.Card1.Value > thisGame.Card2.Value {
							thisGame.Wins1++
							thisGame.Turn = 1
						} else {
							thisGame.Wins2++
							thisGame.Turn = 2
						}
					} else if thisGame.Card1.Mode == thisGame.Hokm {
						thisGame.Wins1++
						thisGame.Turn = 1
					} else if thisGame.Card2.Mode == thisGame.Hokm {
						thisGame.Wins2++
						thisGame.Turn = 2
					} else {
						thisGame.Wins2++
						thisGame.Turn = 2
					}
					thisGame.Side1Cards = deleteCard(thisGame.Side1Cards, Card1Index)
					thisGame.Side2Cards = deleteCard(thisGame.Side2Cards, Card2Index)
					thisGame.Card1.Mode = 0
					thisGame.Card1.Value = 0
					thisGame.Card2.Mode = 0
					thisGame.Card2.Value = 0
				} else {
					thisGame.Turn = 1
				}
				jsonVal, err := json.Marshal(thisGame)
				if err != nil {
					fmt.Println(err)
				}
				ioutil.WriteFile("./games/"+gameIDstr+"/game.json", jsonVal, 0644)
			} else {
				m["info"] = "not your turn"
			}
		}
	}
	m["Wins1"] = thisGame.Wins1
	m["Wins2"] = thisGame.Wins2
	if onuser.ID == thisGame.Side1ID {
		m["YourCards"] = thisGame.Side1Cards
	}
	if onuser.ID == thisGame.Side2ID {
		m["YourCards"] = thisGame.Side2Cards
	}
	Render(w, "game", m)
}

func getDownCardsHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	_, err := getCookie("login_cookie", r)
	if err == nil {
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	gameIDstr := r.URL.Query().Get("game")
	if strings.ContainsAny(gameIDstr, "></*=OR ") || gameIDstr == "" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	var thisGame game
	jsonFile, err := ioutil.ReadFile("./games/" + gameIDstr + "/game.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(jsonFile, &thisGame)
	if thisGame.Card1.Mode != 0 {
		m["DownCard1"] = thisGame.Card1
	}
	if thisGame.Card2.Mode != 0 {
		m["DownCard2"] = thisGame.Card2
	}
	m["Wins1"] = thisGame.Wins1
	m["Wins2"] = thisGame.Wins2
	Render(w, "downCards", m)
}

func endGameGetHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	var onuser user
	cookie, err := getCookie("login_cookie", r)
	if err == nil {
		usrIDstr := cookie
		usrID, _ := strconv.Atoi(usrIDstr)
		onuser = getUserByID(usrID)
		m["user"] = onuser
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	gameIDstr := r.URL.Query().Get("game")
	if strings.ContainsAny(gameIDstr, "></*=OR ") || gameIDstr == "" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM games WHERE id=" + gameIDstr + ";")
	var tempGame game
	var gameID int
	var finded bool
	for rows.Next() {
		rows.Scan(&gameID, &tempGame.Side1ID, &tempGame.Side2ID)
		finded = true
	}
	if !finded {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	if onuser.ID == tempGame.Side1ID || onuser.ID == tempGame.Side2ID {
		JsonFile, err := ioutil.ReadFile("./games/" + gameIDstr + "/game.json")
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(JsonFile, &tempGame)
		if tempGame.Finished == false {
			http.Redirect(w, r, "/user", http.StatusFound)
			return
		}
		user1 := getUserByID(tempGame.Side1ID)
		user2 := getUserByID(tempGame.Side2ID)
		if tempGame.Wins1 == 7 {
			user1.GamesWin++
			user2.GamesLose++
		}
		if tempGame.Wins2 == 7 {
			user2.GamesWin++
			user1.GamesLose++
		}
		stmt, _ := db.Prepare("UPDATE userinfo SET gameslose=? WHERE uid=?")
		stmt.Exec(user1.GamesLose, user1.ID)
		stmt, _ = db.Prepare("UPDATE userinfo SET gameswin=? WHERE uid=?")
		stmt.Exec(user1.GamesWin, user1.ID)
		stmt, _ = db.Prepare("UPDATE userinfo SET gameslose=? WHERE uid=?")
		stmt.Exec(user2.GamesLose, user2.ID)
		stmt, _ = db.Prepare("UPDATE userinfo SET gameswin=? WHERE uid=?")
		stmt.Exec(user2.GamesWin, user2.ID)
		stmt, _ = db.Prepare("DELETE FROM games WHERE id=?")
		stmt.Exec(gameID)
		stmt.Close()
		os.RemoveAll("./games/" + gameIDstr)
		http.Redirect(w, r, "/user", http.StatusFound)
	} else {
		http.Redirect(w, r, "/user", http.StatusFound)
	}
}

func setCookie(name string, value string, w http.ResponseWriter) {

	encval, err := encrypt(value)
	if err != nil {
		fmt.Println(err)
	}
	cookie := &http.Cookie{
		Name:   name,
		Value:  encval,
		MaxAge: 0,
	}
	http.SetCookie(w, cookie)
}

func getCookie(name string, r *http.Request) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	val := cookie.Value
	decVal, err := decrypt(val)
	if err != nil {
		return "", err
	}
	return decVal, nil
}

func getUserByID(id int) user {
	var usr user
	db, _ := sql.Open("sqlite3", "./database/database.db")
	rows, _ := db.Query("SELECT * FROM userinfo WHERE uid='" + strconv.Itoa(id) + "';")
	for rows.Next() {
		rows.Scan(&usr.ID, &usr.Username, &usr.Password, &usr.Email, &usr.GamesWin, &usr.GamesLose)
	}
	return usr
}
