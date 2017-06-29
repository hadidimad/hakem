package main

import (
	"strconv"
)

// 1:clubs 2:diamonds 3:hearts 4:spades
//value	==> 1-10=1-10 11:jack  12:queeen 13:king
type card struct {
	Mode  int
	Value int
}

func getCardImg(Card card) string {
	var path string
	switch Card.Mode {
	case 1:
		path = "/static/cards_img/clubs/"
	case 2:
		path = "/static/cards_img/diamonds/"
	case 3:
		path = "/static/cards_img/hearts/"
	case 4:
		path = "/static/cards_img/spades/"
	}
	switch Card.Value {
	case 11:
		path += "jack.png"
	case 12:
		path += "queen.png"
	case 13:
		path += "king.png"
	default:
		path += strconv.Itoa(Card.Value) + ".png"
	}
	return path
}

func deleteCard(cards []card, index int) []card {
	return append(cards[:index], cards[index+1:]...)
}
