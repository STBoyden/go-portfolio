package types

type Language struct {
	HexColour string
	Name      string
}

type Repository struct {
	Name        string
	Description string
	Url         string
	Languages   []Language
}
