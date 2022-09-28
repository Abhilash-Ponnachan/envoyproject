package main

import (
	"log"
	"os"
	"sync"
)

// env var keys
const portEnv = "PORT"
const appNameEnv = "APPNAME"
const bgColorEnv = "BGCOLOR"
const fgColorEnv = "FGCOLOR"

// const values
const assetsDir = "./assets"
const templatesDir = "./templates"

// default values
const portDef = "8080"
const appNameDef = "Demo App - 1"
const bgColorDef = "#aeeaf2"
const fgColorDef = "#121e59"

type configData struct {
	AssetsDir    string
	TemplatesDir string
	Port         string
	AppName      string
	BgColor      string
	FgColor      string
}

var once sync.Once
var cf *configData

func config() *configData {
	if cf == nil {
		once.Do(
			func() {
				cf = &configData{}
				cf.load()
				cf.log()
			})
	}
	return cf
}

func (cf *configData) load() {

	// nested function to set check and set fields from env
	fnSetVal := func(field *string, defVal string, envKey string) {
		*field = defVal
		e := os.Getenv(envKey)
		if e != "" {
			*field = e
		}
	}

	fnSetVal(&cf.Port, portDef, portEnv)
	fnSetVal(&cf.AppName, appNameDef, appNameEnv)
	fnSetVal(&cf.BgColor, bgColorDef, bgColorEnv)
	fnSetVal(&cf.FgColor, fgColorDef, fgColorEnv)

	cf.AssetsDir = assetsDir
	cf.TemplatesDir = templatesDir
}

func (cf *configData) log() {
	log.Printf("*** conf data *** \n %#v", *cf)
}
