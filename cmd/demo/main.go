package main

import (
	"fmt"
	"github.com/luowensheng/dream/internal/demo"
	"github.com/luowensheng/dream"
)

func main(){
	dream.CreateApp("t", 9090, demo.TicTacToe)
	fmt.Println("Yellow")
}