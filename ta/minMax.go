package ta

func Max(numbers []float64) float64 {
	var largerNumber, temp float64

	for _, n := range numbers {
		if n > temp {
			temp = n
			largerNumber = temp
		}
	}
	return largerNumber
}

func Min(numbers []float64) float64 {
	temp := numbers[0]
	smallerNumber := numbers[0]

	for _, n := range numbers {
		if n < temp {
			temp = n
			smallerNumber = temp
		}
	}
	return smallerNumber
}
