package ui

import (
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

func ShowWelcomeScreen() {
	figure.NewFigure("S3 Auditor", "slant", true).Print()
	color.Cyan("\nWelcome to the AWS S3 Bucket Auditor!\n")
}

func ShowError(format string, args ...interface{}) {
	color.Red(format, args...)
}

func ShowSuccess(format string, args ...interface{}) {
	color.Green(format, args...)
}
