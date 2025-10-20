package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Game struct {
	Rows, Cols    int
	Board         [][]int
	CurrentPlayer int
	Mode          int // 1 = vs IA, 2 = deux joueurs
	GameOver      bool
	AiDifficulty  string
	Winner        int // 0 = none, 1 or 2
	Player1       string
	Player2       string
}

var (
	game      = &Game{}
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"loop": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
	}).ParseGlob("templates/*.tmpl"))
)

func (g *Game) initBoard(rows, cols int) {
	g.Rows = rows
	g.Cols = cols
	g.Board = make([][]int, rows)
	for r := range g.Board {
		g.Board[r] = make([]int, cols)
	}
	g.CurrentPlayer = 1
	g.GameOver = false
	g.Winner = 0
}

func (g *Game) isColumnFull(col int) bool {
	return g.Board[0][col] != 0
}

func (g *Game) isBoardFull() bool {
	for _, val := range g.Board[0] {
		if val == 0 {
			return false
		}
	}
	return true
}

func (g *Game) dropPiece(col, player int) (int, int, bool) {
	for r := g.Rows - 1; r >= 0; r-- {
		if g.Board[r][col] == 0 {
			g.Board[r][col] = player
			return r, col, true
		}
	}
	return -1, -1, false
}

func (g *Game) checkWin(player int) bool {
	// horizontal
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols-3; c++ {
			if g.Board[r][c] == player && g.Board[r][c+1] == player &&
				g.Board[r][c+2] == player && g.Board[r][c+3] == player {
				return true
			}
		}
	}
	// vertical
	for c := 0; c < g.Cols; c++ {
		for r := 0; r < g.Rows-3; r++ {
			if g.Board[r][c] == player && g.Board[r+1][c] == player &&
				g.Board[r+2][c] == player && g.Board[r+3][c] == player {
				return true
			}
		}
	}
	// diagonal \
	for r := 0; r < g.Rows-3; r++ {
		for c := 0; c < g.Cols-3; c++ {
			if g.Board[r][c] == player && g.Board[r+1][c+1] == player &&
				g.Board[r+2][c+2] == player && g.Board[r+3][c+3] == player {
				return true
			}
		}
	}
	// diagonal /
	for r := 3; r < g.Rows; r++ {
		for c := 0; c < g.Cols-3; c++ {
			if g.Board[r][c] == player && g.Board[r-1][c+1] == player &&
				g.Board[r-2][c+2] == player && g.Board[r-3][c+3] == player {
				return true
			}
		}
	}
	return false
}

func (g *Game) aiPlay() {
	col := rand.Intn(g.Cols)
	for g.isColumnFull(col) {
		col = rand.Intn(g.Cols)
	}
	g.dropPiece(col, 2)
	if g.checkWin(2) {
		g.GameOver = true
		g.Winner = 2
	} else if g.isBoardFull() {
		g.GameOver = true
		g.Winner = 0
	} else {
		g.CurrentPlayer = 1
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html.tmpl", nil)
	})

	http.HandleFunc("/newgame", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		player1 := r.FormValue("player1")
		player2 := r.FormValue("player2")
		difficulty := r.FormValue("difficulty")

		game.Player1 = player1
		game.Player2 = player2
		game.AiDifficulty = difficulty

		switch difficulty {
		case "Easy":
			game.initBoard(6, 7)
		case "Normal":
			game.initBoard(6, 9)
		case "Hard":
			game.initBoard(7, 8)
		default:
			game.initBoard(6, 7)
		}

		if player2 == "" {
			game.Mode = 1
		} else {
			game.Mode = 2
		}

		http.Redirect(w, r, "/game", http.StatusSeeOther)
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "game.html.tmpl", game)
	})

	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/game", http.StatusSeeOther)
			return
		}
		if game.GameOver {
			http.Redirect(w, r, "/game", http.StatusSeeOther)
			return
		}
		colStr := r.FormValue("col")
		col, err := strconv.Atoi(colStr)
		if err != nil || col < 0 || col >= game.Cols {
			http.Redirect(w, r, "/game", http.StatusSeeOther)
			return
		}

		if game.isColumnFull(col) {
			http.Redirect(w, r, "/game", http.StatusSeeOther)
			return
		}

		game.dropPiece(col, game.CurrentPlayer)

		if game.checkWin(game.CurrentPlayer) {
			game.GameOver = true
			game.Winner = game.CurrentPlayer
		} else if game.isBoardFull() {
			game.GameOver = true
			game.Winner = 0
		} else {
			game.CurrentPlayer = 3 - game.CurrentPlayer
			if game.Mode == 1 && game.CurrentPlayer == 2 && !game.GameOver {
				time.Sleep(300 * time.Millisecond)
				game.aiPlay()
			}
		}
		http.Redirect(w, r, "/game", http.StatusSeeOther)
	})

	http.HandleFunc("/rematch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/game", http.StatusSeeOther)
			return
		}
		game.initBoard(game.Rows, game.Cols)
		http.Redirect(w, r, "/game", http.StatusSeeOther)
	})

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("✅ Serveur Power4 lancé sur http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}
