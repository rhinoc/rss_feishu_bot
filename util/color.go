package util

func GetColorByIndex(index int) string {
	colors := []string{
		"green",
		"yellow",
		"red",
		"purple",
		"carmine",
		"blue",
		"turquoise",
		"lime",
		"orange",
		"violet",
		"indigo",
		"wathet",
	}
	return colors[index%len(colors)]
}
