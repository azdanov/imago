package database

import "embed"

const MigrationsDir = "migrations"

//go:embed migrations/*
var FS embed.FS
