package main

import (
	"bufio"
	"errors"
	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inancgumus/screen"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"path"
	"rpg/common"
	"rpg/console"
	"rpg/defs"
	"rpg/files"
	"rpg/game"
	"rpg/gui"
	"rpg/renderer"
	"rpg/util"
	"strings"
	"time"
)

func main() {
	/*
	cpuProfileFile, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	defer cpuProfileFile.Close()

	// start CPU profiling
	if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()
	*/
	if len(os.Args) > 1 {
		argCmd := strings.ToLower(os.Args[1])
		switch argCmd {
		case "ascii":
			runGameWithConsole(createGameAndUIContext())
		case "title":
			showTitleScreen()
		}
	} else {
		runGameWithEbiten(createGameAndUIContext())
	}
}

func showTitleScreen() {
	logoFile := util.MustOpenReader(path.Join("assets","logo.txt"))
	defer logoFile.Close()
	//reader := bufio.NewReader(logoFile)
	scanner := bufio.NewScanner(logoFile)
	logoHeight := 0
	logoWidth := 0
	var logoLines []string
	for scanner.Scan() {
		line := scanner.Text()
		logoLines = append(logoLines, line)
		logoHeight++
		logoWidth = max(logoWidth, len(line))
	}

	s, _ := tcell.NewScreen()
	s.Init()
	s.Clear()
	w, h := s.Size()

	// center logo
	logoStartY := 3
	for y := 0; y < logoHeight; y++ {
		line := logoLines[y]
		for x := 0; x < len(line); x++ {
			s.SetContent((w-logoWidth)/2+x, logoStartY+y, rune(line[x]), nil, tcell.StyleDefault)
		}
	}

	lineAfterLogo := logoStartY + logoHeight + 1
	credits := "a game by Felix Ruzzoli"
	centeredCredits := (w - len(credits)) / 2
	for x, c := range credits {
		s.SetContent(centeredCredits+x, lineAfterLogo, c, nil, tcell.StyleDefault)
	}

	// center prompt vertically
	promptStartY := ((h - 1) / 2) + 2
	prompt := "Do you want graphics? (y/n)"
	centeredPrompt := (w - len(prompt)) / 2
	for x, c := range prompt {
		s.SetContent(centeredPrompt+x, promptStartY, c, nil, tcell.StyleDefault)
	}
	s.ShowCursor(centeredPrompt + len(prompt), promptStartY)
	s.SetCursorStyle(tcell.CursorStyleBlinkingUnderline)
	s.Show()
	graphics := false
	outerloop:
	for event := s.PollEvent(); event != nil; event = s.PollEvent() {
		if ev, ok := event.(*tcell.EventKey); ok {
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'y':
					fallthrough
				case 'Y':
					graphics = true
					break outerloop
				case 'n':
					fallthrough
				case 'N':
					graphics = false
					break outerloop
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.Clear()
	s.Fini()
	if graphics {
		runGameWithEbiten(createGameAndUIContext())
	} else {
		runGameWithConsole(createGameAndUIContext())
	}

	screen.Clear()
	screen.MoveTopLeft()

	messageFile := util.MustOpenReader(path.Join("assets","endMessage.txt"))
	defer messageFile.Close()
	//reader := bufio.NewReader(logoFile)
	msgScanner := bufio.NewScanner(messageFile)
	for msgScanner.Scan() {
		line := msgScanner.Text()
		println(line)
	}
}

func createGameAndUIContext() (*game.GameInstance, common.UIContext) {
	userConfig := common.NewConfigurationFromRecFile("config.rec")
	startInfos := common.StartInfo{
		Title:       "RPG",
		ScreenSize:  userConfig.WindowSize,
		TileScale:   4.0,
		LDTKMapFile: "assets/maps.ldtk",
		FontInfos: common.FontInfo{
			TTFFile: userConfig.DefaultFont,
			TTFSize: userConfig.DefaultFontSize,
		},
		UserConfiguration: userConfig,
		FileSystem: files.GetFileSystem(),
	}
	theGame := game.NewGameInstance(startInfos, game.NewNullUI())
	uiContext := common.UIContext{
		GameState:    theGame.GetWorldState,
		PlayerAction: theGame.GetActions,
		AudioPlayer:  theGame.GetAudio(),
		ResetGame:    theGame.ResetGame,
		Animator: func() common.Animator {
			return theGame.GetActionBroker()
		},
		GetLog:      theGame.GetLog,
		StartInfos:  startInfos,
		Definitions: theGame.GetGameData(),
	}
	return theGame, uiContext
}

func runGameWithConsole(theGame *game.GameInstance, context common.UIContext) {
	theGame.SetConsoleMode(true)
	con := console.NewConsole(context)
	theGame.SetAndInitUserInterface(con)

	terminationCode := 0
	loopConfig := util.LoopConfig{
		TargetFPS:        30,
		IdleThreshold:    5,
		CurrentTimeFunc:  func() float64 { return float64(time.Now().UnixNano()) * 1e-9 },
		ProcessInputFunc: con.HandleInput,
		UpdateFunc:       con.Update,
		RenderFunc:       con.Draw,
	}

	loop := util.CreateLoop(loopConfig, &terminationCode)
	loop()

}

// runGameWithEbiten will start the infinitely running game loop
// game will receive the update and draw calls
func runGameWithEbiten(theGame *game.GameInstance, context common.UIContext) {
	startInfos := context.StartInfos
	tileRenderer := renderer.NewTileRenderer(startInfos.TileScale)
	tileRenderer.SetColorFromName(defs.ColorFromName)
	tileRenderer.SetDefaultAtlas(theGame.GetWorldState().GetWorldTexture())
	tileRenderer.SetDefaultColors(defs.DarkTextColor, color.Black)
	tileRenderer.SetWhiteTile(defs.IconWhite)

	tileMapAnimations := map[int32]gui.TilemapAnimInfo{
		28:                       {FrameCount: 2, DelayInSeconds: 1.5},
		32:                       {FrameCount: 2, DelayInSeconds: 0.2},
		defs.IconEffectFireSmall: {FrameCount: 2, DelayInSeconds: 0.2},
	}
	uiController := gui.NewUIController(tileRenderer, context, startInfos.UserConfiguration)
	uiController.SetTileMapAnimations(tileMapAnimations)
	uiController.SetTTFFont(util.MustOpen(startInfos.FontInfos.TTFFile), startInfos.FontInfos.TTFSize)
	uiController.SetAsciiAtlas(renderer.NewTextureAtlas(path.Join("assets", "Bm437_IBM_CGA.png"), 8, 8))
	theGame.SetAndInitUserInterface(uiController)

	screenSize := uiController.GetDeviceIndependentScreenSize()

	ebiten.SetWindowTitle("Lux Aeterna")
	ebiten.SetWindowSize(screenSize.X, screenSize.Y)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSizeLimits(640, 480, -1, -1)
	ebiten.SetWindowClosingHandled(false)

	err := ebiten.RunGameWithOptions(uiController, &ebiten.RunGameOptions{
		GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
	})
	if err != nil && !errors.Is(err, ebiten.Termination) {
		log.Fatal(err)
	}
}

