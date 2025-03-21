package types

type Language struct {
	HexColour string
	Name      string
}

type Repository struct {
	Name        string
	Owner       string
	Description string
	Url         string
	Languages   []Language
}
