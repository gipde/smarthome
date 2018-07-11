package tests

import (
	"github.com/revel/revel/testing"
)

type MiriamTest struct {
	testing.TestSuite
}

func (t *MiriamTest) Bevore() {
	println()
}

func (t *MiriamTest) TestLoop2() {
	for i := 0; i < 3; i++ {
		moveNorth()
		getNectar()
		moveNorth()
		makeHoney()
	}
}

func (t *MiriamTest) TestLoop1() {
	for count2 := 0; count2 < 2; count2++ {
		moveNorth()
		for count := 0; count < 10; count++ {
			getNectar()
		}
	}
	for count4 := 0; count4 < 2; count4++ {
		moveNorth()
		for count3 := 0; count3 < 10; count3++ {
			makeHoney()
		}
	}

}

func getNectar() {
	println("NEKTAR")
}
func makeHoney() {
	println("HONIG")
}

func goSouth() {
	println("SÜDEN")
}
func goEast() {
	println("OSTEN")
}

func goWest() {
	println("WESTEN")
}
func moveNorth() {
	println("NORDEN")
}
