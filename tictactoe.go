package main

import "fmt"

func Square(value string, onSquareClick func(Record) string) {

	buttonRef := El("button").Class("square").Content(value)
	buttonRef.OnWithParams("click", func(r Record) {
		buttonRef.SetTextContent(onSquareClick(r))
	}, Record{"value": buttonRef.getTextContent()},
	)
}

// https://react.dev/learn/tutorial-tic-tac-toe

func Board(xIsNext *DOMVariable[bool], squares *DOMVariable[[]string], onPlay func([]string)) {

	winner := newNamedDOMVariable("winner", "")
	status := newNamedDOMVariable("status", "Next player: X")

	handleClick := func(i int, value string) {
		// fmt.Println("i = ", i, "&&", "value = ", value, " -> xIsNext", xIsNext)
		squaresValue := squares.Value()
		if squaresValue[i] != "" {
			return
		}

		winner.SetValue(calculateWinner(squaresValue))
		// fmt.Println("*******************************************", squaresValue, winner.Value(), "________")

		if winner.Value() != "" {
			return 
		}
		onPlay(squaresValue)
		
	}

	winner.OnValueUpdated(func(newWinner string) {
		if newWinner != "" {
			status.SetValue("Winner: " + newWinner)
		} else {
			if xIsNext.Value() {
				status.SetValue("Next player: X")
			} else {
				status.SetValue("Next player: O")
			}
		}
	})
    
	El("div").Class("status").DOMContent(status)

	for i := 0; i < 3; i++ {
		El("div").Class("board-row").Inner(func() {

			for j := 0; j < 3; j++ {
				count := i*3 + j
				Square(squares.Value()[j], func(params Record) string { 
					value := params["value"]
					if value != "" {
						return value 
					}
					fmt.Println("_______________", value, "++++++++++++++++++++++++")
					var newValue string
					if xIsNext.Value() {
						newValue = "X"
					} else {
						newValue = "O"
					}
					squares.UpdateValue(func(s []string) []string {
						s[count] = newValue
						fmt.Println("******** squares=$", ToJsonString(s), "--", count, "$***")
						return s
					})
					handleClick(count, value)
					return newValue
				})
			}
		})
	}
}

func Game() {

	LoadCSS("./assets/tictactoe.css")
	// [][]*DOMVariable[{[]string{}}]
	history := newNamedDOMVariable("history", [][]string{{}})
	
	history.UpdateValue(func(s [][]string) [][]string {
		for i := 0; i < 9; i++ {
			s[0] = append(s[0], "")
		}
		return s
	})

	currentMove := newNamedDOMVariable("currentMove", 0)
	xIsNext := newNamedDOMVariable("xIsNext", currentMove.Value() %2 == 0)
	currentMove.OnValueUpdated(func(i int) {
		xIsNext.SetValue(i % 2 == 0)
		fmt.Println("CURRENT MOVE UPDATED @@@@@@", i, xIsNext.Value(), ")))")
	})
	currentSquares := newNamedDOMVariable("currentSquares", history.Value()[currentMove.Value()])

	handlePlay := func(nextSquares []string) {
		history.UpdateValue(func(s [][]string) [][]string {
			return append(s, nextSquares)
		})
		// currentMove.SetValue(len(history.Value()) - 1)
		currentMove.UpdateValue(func(i int) int { return i+1})
		fmt.Println("HISTORY UPDATED @@@@@@", nextSquares, "**", currentMove)

	}

	jumpTo := func(nextMove int) {
		fmt.Println("CURRENT MOVE JUMPTO @@@@@@", nextMove)
		currentMove.SetValue(nextMove)
	}

	moves := func() {

		var description string
		for move := range history.Value() {

			if move > 0 {
				description = fmt.Sprintf("Go to move #%d", move)
			} else {
				description = "Go to game start"
			}

			El("li").Attr("key", fmt.Sprintf("%d", move)).Inner(func() {
				El("button").Content(description).On("click", func() { jumpTo(move) })
			})
		}
	}

	El("div").Class("game").Inner(func() {
		El("div").Class("game-board").Inner(func() {
			Board(xIsNext, currentSquares, handlePlay)
		})
		El("div").Class("game-info").Inner(func() {
			El("ol").Inner(moves)
		})
	})

}

func calculateWinner(squares []string) string {

	lines := [8][3]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}

	for i := range lines {
		line := lines[i]
		a := line[0]
		b := line[1]
		c := line[2]
        win := squares[a] != "" && squares[a] == squares[b] && squares[a] == squares[c] 
		fmt.Println(line, []string{squares[a], squares[b], squares[c]}, "=>>>>>", win)
		if win {
			return squares[a]
		}
	}
	return ""
}

func TicTacToe() {
	CreateApp("TicTacToe", Game)
}
