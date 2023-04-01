package main

import "fmt"

type P struct {
	X, Y float64
}

func F_new_point(x, y float64) P {
	return P{x, y}
}

func F_add_point(p1, p2 P) P {
	return P{p1.X + p2.X, p1.Y + p2.Y}
}

func F_mutate_point(p *P) {
	p.X = p.X * 2
	p.Y = p.Y * 2
}

// func F_mutate_point_array(p []P) {
// 	for i := range p {
// 		p[i].X = p[i].X * 2
// 		p[i].Y = p[i].Y * 2
// 	}
// }

func main() {
	fmt.Println("main")
}
