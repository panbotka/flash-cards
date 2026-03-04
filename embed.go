package flashcards

import "embed"

//go:embed frontend/dist/*
var FrontendDist embed.FS
