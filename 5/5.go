package main

import (
	"fmt"
	"math"
)

// ЗАДАНИЕ 1

type Person struct {
	name string
	age int
}

func (p Person) info() {
	fmt.Printf("Имя: %v\nВозраст: %v\n", p.name, p.age)
}

// ЗАДАНИЕ 2

func (p *Person) birthday() {
	p.age += 1
	fmt.Printf("%v стал на год старше\n", p.name)
}

// ЗАДАНИЕ 3

type Circle struct {
	radius float64
}

func (c Circle) area() float64 {
	return math.Pi * math.Pow(c.radius, 2)
}

// ЗАДАНИЕ 4

type Shape interface {
	area() float64
}

type Rectangle struct {
	width float64
	height float64
}

func (r Rectangle) area() float64 {
	return r.width * r.height
}

func calcArea(s Shape) {
	fmt.Println(s.area())
}

// ЗАДАНИЕ 5

func calcSomeAreas(list []Shape) {
	for _, value := range list {
		fmt.Println(value.area())
	}
}

// ЗАДАНИЕ 6

type Stringer interface {
	info()
	get_price()
}

type Book struct {
	author string
	name string
	genre string
	year int
	price int
}

func (b Book) info() {
	fmt.Printf("Автор: %v\nНазвание: %v\nЖанр: %v\nГод: %v\n", b.author, b.name, b.genre, b.year)
}
func (b Book) get_price() {
	fmt.Println("Цена:", b.price)
}

func main() {
	fmt.Println("====================================================")

	fmt.Println("\nЗАДАНИЯ 1-2")

	person := Person{"Семён", 20}
	person.info()
	person.birthday()
	person.info()

	//

	fmt.Println("\nЗАДАНИЕ 3")

	circle := Circle{9}
	fmt.Println("Радиус:", circle.area())

	// 

	fmt.Println("\nЗАДАНИЕ 4")

	calcArea(&Rectangle{5, 20})
	calcArea(&circle)

	fmt.Println("\nЗАДАНИЕ 5")

	list := []Shape{
		Shape(&Rectangle{2,3}), 
		Shape(&Circle{4}), 
		Shape(&Rectangle{8,6}), Shape(&Rectangle{1,2}),
		Shape(&Circle{15}),
	}
	calcSomeAreas(list)

	fmt.Println("\nЗАДАНИЕ 6")

	book1 := Book{
		author: "Антуан де Сент-Экзюпери",
		name: "Маленький Принц",
		genre: "Философская сказка",
		year: 1942,
		price: 999,
	}

	book2 := Book{
		author: "Марк Аврелий",
		name: "К самому себе",
		genre: "-",
		year: 170,
		price: 259,
	}

	s1 := Stringer(&book1)
	s2 := Stringer(&book2)

	s1.info()
	s1.get_price()
	
	fmt.Println()

	s2.info()
	s2.get_price()

	fmt.Println("====================================================")
}