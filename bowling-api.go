package main

import (
	"encoding/json"
	"net/http"
)

// Game contains the state of a bowling game.
type Game struct {
	rolls   []int
	current int
}

// NewGame allocates and starts a new game of bowling.
func NewGame() *Game {
	game := new(Game)
	game.rolls = make([]int, maxThrowsPerGame)
	return game
}

// Roll rolls the ball and knocks down the number of pins specified by pins.
func (gm *Game) Roll(pins int) {
	gm.rolls[gm.current] = pins
	gm.current++
}

// Score calculates and returns the player's current score.
func (gm *Game) Score() (sum int) {
	for throw, frame := 0, 0; frame < framesPerGame; frame++ {
		if gm.isStrike(throw) {
			sum += gm.strikeBonusFor(throw)
			throw += 1
		} else if gm.isSpare(throw) {
			sum += gm.spareBonusFor(throw)
			throw += 2
		} else {
			sum += gm.framePointsAt(throw)
			throw += 2
		}
	}
	return sum
}

// isStrike determines if a given throw is a strike or not.
// A strike is knocking down all pins in one throw.
func (gm *Game) isStrike(throw int) bool {
	return gm.rolls[throw] == allPins
}

// strikeBonusFor calculates and returns the strike bonus for a throw.
func (gm *Game) strikeBonusFor(throw int) int {
	return allPins + gm.framePointsAt(throw+1)
}

// isSpare determines if a given frame is a spare or not.
// A spare is knocking down all pins in one frame with two throws.
func (gm *Game) isSpare(throw int) bool {
	return gm.framePointsAt(throw) == allPins
}

// spareBonusFor calculates and returns the spare bonus for a throw.
func (gm *Game) spareBonusFor(throw int) int {
	return allPins + gm.rolls[throw+2]
}

// framePointsAt computes and returns the score in a frame specified by throw.
func (gm *Game) framePointsAt(throw int) int {
	return gm.rolls[throw] + gm.rolls[throw+1]
}

// testing utilities:

func (gm *Game) rollMany(times, pins int) {
	for x := 0; x < times; x++ {
		gm.Roll(pins)
	}
}
func (gm *Game) rollSpare() {
	gm.Roll(5)
	gm.Roll(5)
}
func (gm *Game) rollStrike() {
	gm.Roll(10)
}

const (
	// allPins is the number of pins allocated per fresh throw.
	allPins = 10

	// framesPerGame is the numer of frames per bowling game.
	framesPerGame = 10

	// maxThrowsPerGame is the maximum number of throws possible in a single game.
	maxThrowsPerGame = 21
)

// endpoint handlers:

// RollHandler handles the "POST /roll" endpoint.
func RollHandler(gm *Game) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Parse the pins from the request body
		var roll struct {
			Pins int `json:"pins"`
		}
		if err := json.NewDecoder(r.Body).Decode(&roll); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		gm.Roll(roll.Pins)
		w.WriteHeader(http.StatusCreated)

		score := gm.Score()

		//Convert the score to a JSON response
		response := struct {
			Score int `json:"score"`
		}{
			Score: score,
		}
		json.NewEncoder(w).Encode(response)
	}
}

// ScoreHandler handles the "GET /score" endpoint.
func ScoreHandler(gm *Game) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		score := gm.Score()

		//Convert the score to a JSON response
		response := struct {
			Score int `json:"score"`
		}{
			Score: score,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	gm := NewGame()
	http.HandleFunc("/roll", RollHandler(gm))
	http.HandleFunc("/score", ScoreHandler(gm))
	http.ListenAndServe(":8080", nil)
}
