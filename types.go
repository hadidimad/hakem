package main

type user struct {
	ID        int
	Username  string
	Password  string
	Email     string
	GamesWin  int
	GamesLose int
}
type game struct {
	GameStatus int //0:chosing hokm 1:divide cards 2:playing
	Side1ID    int
	Side2ID    int
	Turn       int //0:no One 1:player1 2:player2
	Wins1      int
	Wins2      int
	Hokm       int //hokm card mode 1:clubs 2:diamonds 3:hearts 4:spades
	Finished   bool
	Card1      card
	Card2      card
	Side1Cards []card
	Side2Cards []card
	OutCards   []card
}
type showGames struct {
	GameID        int
	Side1Username string
	Side2Username string
	Wins1         int
	Wins2         int
	Finished      bool
}
