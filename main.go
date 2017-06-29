package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	InitRender()
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./statics"))))
	router.Methods("GET").Path("/").Handler(http.HandlerFunc(indexGetHandler))
	router.Methods("GET").Path("/signup").Handler(http.HandlerFunc(signupGetHandler))
	router.Methods("POST").Path("/signup").Handler(http.HandlerFunc(signupPostHandler))
	router.Methods("GET").Path("/login").Handler(http.HandlerFunc(loginGetHandler))
	router.Methods("POST").Path("/login").Handler(http.HandlerFunc(loginPostHandler))
	router.Methods("GET").Path("/user").Handler(http.HandlerFunc(userPageGetHandler))
	router.Methods("GET").Path("/logout").Handler(http.HandlerFunc(logoutGetHandler))
	router.Methods("GET").Path("/gamemaker").Handler(http.HandlerFunc(gameMakerGetHandler))
	router.Methods("GET").Path("/getuser").Handler(http.HandlerFunc(getUsersGetHandler))
	router.Methods("GET").Path("/startgame").Handler(http.HandlerFunc(startgameGetHanlder))
	router.Methods("GET").Path("/deletegame").Handler(http.HandlerFunc(deleteGameGetHandler))
	router.Methods("GET").Path("/games").Handler(http.HandlerFunc(gameGetHandler))
	router.Methods("GET").Path("/gamesdowncard").Handler(http.HandlerFunc(getDownCardsHandler))
	router.Methods("GET").Path("/endgame").Handler(http.HandlerFunc(endGameGetHandler))
	err := http.ListenAndServe(":8080", router)
	fmt.Println(err)
}
