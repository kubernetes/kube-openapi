package enumtype

// FruitType defines a kind of fruits.
// +enum
type FruitType string

// FruitApple is the Apple
const FruitApple FruitType = "apple"

// FruitBanana is the Banana
const FruitBanana FruitType = "banana"

// FruitRiceBall is the Rice ball that does not seem to belong to
// a fruits basket but has a long comment that is so long that it spans
// multiple lines
const FruitRiceBall FruitType = "onigiri"

// FruitsBasket is the type that contains the enum type.
// +k8s:openapi-gen=true
type FruitsBasket struct {
	Content FruitType `json:"content"`

	// +default=0
	Count  int `json:"count"`
}
