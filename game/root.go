package game

import (
	"cmp"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"rpg/audio"
	"rpg/common"
	"rpg/defs"
	"rpg/files"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/ldtk"
	"rpg/stats"
	"rpg/util"
	"slices"
	"strings"
	"time"
)

type GameInstance struct {
	// base infos / config
	title                       string
	deviceIndependentScreenSize geometry.Point
	configuration               common.Configuration
	startInfo                   common.StartInfo

	// dependencies
	uiController common.UserInterface
	audioPlayer  *audio.Player

	// world state
	state        common.State
	actionBroker *common.Broker // registered hooks

	// random
	randomSource *rand.Rand
	worldTicks   uint64

	// rendering

	// input
	mousePosInPixels geometry.Point
	mousePosOnMap    geometry.Point

	// textserver
	serverStats     *TextServer
	serverInventory *TextServer
	serverEquipment *TextServer
	serverLog       *TextServer

	// player state
	isResting bool

	// game state
	isPaused         bool
	shouldQuitGame   bool
	quitToAscii      bool
	debugModeEnabled bool

	// event handler
	executeAfterTurn []func()

	// data definitions
	dataDefinitions defs.DataContainer

	actions *Actions
	factory *common.Factory

	// announcedToday keeps one performance of a duty from being reported on every
	// tick the Townsperson stands there. Cleared at the Day boundary.
	announcedToday   map[string]bool
	unmetObligations []*common.Obligation

	// The vampire, the town's inhabitants, and where the bodies are.
	blood           *common.Blood
	townspeople     []*common.Townsperson
	bodies          map[string]geometry.Point
	ending          RunEnding
	warnedAboutDawn bool
	isBurning       bool
	daywalker       bool
	zones           []*TownZone
	hunter          *common.Actor
	privacyRefuge   map[string]geometry.Point
	idleDrift       map[string]*idleDrift

	logHook     func(string)
	logBuffer   []string
	onMapChange func()
}

func (g *GameInstance) SetConsoleMode(consoleMode bool) {
	// only disable audio for now
	g.audioPlayer.SetDisablePlayback(consoleMode)
}
func (g *GameInstance) GetTicks() uint64 {
	return g.worldTicks
}

func (g *GameInstance) GetActionBroker() *common.Broker {
	return g.actionBroker
}
func (g *GameInstance) GetMoveTowardsPlayer(from geometry.Point) geometry.Point {
	return g.getMoveOnPlayerDijkstraMap(from, true)
}

func (g *GameInstance) GetMoveAwayFromPlayer(from geometry.Point) geometry.Point {
	return g.getMoveOnPlayerDijkstraMap(from, false)
}
func (g *GameInstance) getMoveOnPlayerDijkstraMap(from geometry.Point, towardsPlayer bool) geometry.Point {
	dijkstraMap := g.state.GetPlayerDijkstraMap()
	if dijkstraMap == nil {
		return from
	}
	currentDistanceToPlayer, exist := dijkstraMap[from]
	if !exist {
		return from
	}
	var compareFunc func(a, b int) bool
	if towardsPlayer {
		compareFunc = func(a, b int) bool { return a < b }
	} else {
		compareFunc = func(a, b int) bool { return a > b }
	}
	currentMap := g.state.GetMap()
	neighbors := currentMap.GetFilteredNeighborsForMovement(from, func(pos geometry.Point) bool { // choose a possible next step
		if !currentMap.IsCurrentlyPassable(pos) { // only walk on passable tiles
			return false
		}
		if _, existsTransition := currentMap.GetTransitionAt(pos); existsTransition { // don't walk on transitions
			return false
		}
		neighborDist, neighborExists := dijkstraMap[pos]
		if !neighborExists {
			return false
		}
		neighborsWithTransition := currentMap.GetFilteredNeighbors(pos, func(pos geometry.Point) bool { // in fact, don't walk on tiles next to transitions either
			if !currentMap.IsCurrentlyPassable(pos) {
				return false
			}
			if _, existsTransition := currentMap.GetTransitionAt(pos); existsTransition {
				return true
			}
			return false
		})
		if len(neighborsWithTransition) > 0 {
			return false
		}
		return compareFunc(neighborDist, currentDistanceToPlayer) // depends on rolling direction on our dijkstra map
		//return neighborDist < currentDistanceToPlayer // depends on rolling direction on our dijkstra map
	})

	if len(neighbors) == 0 {
		return from
	}
	// we found some locations we can move to, that also bring us closer to the player
	// which of these is the closest to the player?

	nearestDist := math.MaxInt
	nearestPos := from

	for _, neighbor := range neighbors {
		neighborDist, _ := dijkstraMap[neighbor]
		//if neighborDist < nearestDist {  // depends on rolling direction on our dijkstra map
		if compareFunc(neighborDist, nearestDist) {
			nearestDist = neighborDist
			nearestPos = neighbor
		}
	}

	return nearestPos
}

func (g *GameInstance) GetAudio() *audio.Player {
	return g.audioPlayer
}

func NewGameInstance(startInfo common.StartInfo, uiController common.UserInterface) *GameInstance {
	seed := startInfo.RandomSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	g := &GameInstance{
		startInfo:    startInfo,
		uiController: uiController,
		randomSource: rand.New(rand.NewSource(seed)),
		audioPlayer:  audio.NewAudioPlayer(startInfo.FileSystem, audio.SampleRate44100),
	}
	println(fmt.Sprintf("GameInstance created.."))
	println(fmt.Sprintf("Random source loaded.."))
	println(fmt.Sprintf("Audio player loaded.."))
	g.dataDefinitions = defs.LoadGameData(func(err string) {
		println(err)
	})
	println(fmt.Sprintf("Data definitions loaded.."))
	g.ResetGame(g.dataDefinitions.ClassDefinitions[0])
	return g
}

// nextActionFor is what the TimeKeeper asks of everyone it moves: what they do
// next and how many in-game seconds it costs them.
func (g *GameInstance) nextActionFor(being common.TimedBeing) (float64, func()) {
	if actor, isActor := being.(*common.Actor); isActor {
		return g.getScheduledAction(actor)
	}
	return 0, nil
}
func (g *GameInstance) giveAndEquipItemToActor(actor *common.Actor, item *common.Item) {
	actor.GetInventory().Add(item)
	g.actionBroker.Execute(common.NewEquipItemAction(common.AbstractCause("Game Start"), actor, item))
}

// getActionContext requires the following fields to be initialized:
// g.audioPlayer, g.dataDefinitions, g.state
func (g *GameInstance) getActionContext() common.ActionContext {
	return common.ActionContext{
		GameState:     g.GetWorldState(),
		UserInterface: g.GetUserInterface,
		AudioPlayer:   g.audioPlayer,
		Data:          g.dataDefinitions,
		Msg:           g.PlayerMessage,
		Speak:         g.ActorSpeak,
	}
}

func (g *GameInstance) ResetGame(playerStartDefinition defs.LoadoutDefinition) {
	// we have loaded at this point:
	// - random source
	// - audio player
	// - data definitions
	g.actions = nil
	g.actionBroker = nil
	g.factory = nil
	//g.uiController =

	g.configuration = g.startInfo.UserConfiguration
	g.deviceIndependentScreenSize = g.startInfo.ScreenSize

	if g.configuration.ServeLog {
		g.serverLog = NewTextServer("rogue_log", 1)
	}
	if g.configuration.ServeStats {
		g.serverStats = NewTextServer("rogue_stats", 2)
	}
	if g.configuration.ServeInventory {
		g.serverInventory = NewTextServer("rogue_inventory", 3)
	}
	if g.configuration.ServeEquipment {
		g.serverEquipment = NewTextServer("rogue_equipment", 4)
	}

	// WE DO NOT HAVE A UI YET
	if g.configuration.ServeLog {
		g.SetLogHook(g.serverLog.Write)
	}

	ldtkFileName := g.startInfo.LDTKMapFile
	ldtkFile, _ := files.ReadFile(ldtkFileName)
	ldtkProject, _ := ldtk.Open(ldtkFile)

	timeKeeper := common.NewTimeKeeper(g.nextActionFor)
	println(fmt.Sprintf("Timekeeper created.."))

	g.state = common.NewState(ldtkProject)
	g.announcedToday = make(map[string]bool)
	g.unmetObligations = nil
	g.blood = common.NewBlood()
	g.townspeople = nil
	g.bodies = make(map[string]geometry.Point)
	g.privacyRefuge = make(map[string]geometry.Point)
	g.idleDrift = make(map[string]*idleDrift)
	g.hunter = nil
	g.ending = RunOngoing
	g.state.SetRandomSource(g.randomSource)
	g.state.SetTimeMultiplier(g.configuration.TimeMultiplier)
	g.state.SetPlayerMessage(g.PlayerMessage)
	g.state.SetTimeKeeper(timeKeeper)
	g.state.SelectedClass = playerStartDefinition

	g.actionBroker = common.NewBroker(g.getActionContext())
	g.actionBroker.RegisterHook(common.NewOnHealthChangedHook())

	g.actionBroker.RegisterHook(common.NewMainOnMoveHook(g.configuration.AutoPickup)) // update fov, etc.
	g.actionBroker.RegisterHook(common.NewMainOnKillHook())
	g.actionBroker.RegisterHook(common.NewMainOnDropHook())
	g.actionBroker.RegisterHook(common.NewMainOnEquipHook())

	g.factory = common.NewFactory(common.FactoryContext{
		Definitions:  g.dataDefinitions,
		RandomSource: g.randomSource,
		State:        g.GetWorldState(),
		ActionBroker: g.actionBroker,
		Msg:          g.PlayerMessage,
	})
	g.factory.SetUserInterface(g.GetUserInterface)
	g.factory.SetQueueAfterTurn(g.queueAfterTurn)

	g.state.SetPlayerMessage(g.PlayerMessage)
	g.state.SetAnimator(g.actionBroker)
	g.state.SetActionBroker(g.actionBroker)
	g.state.SetFactory(g.factory)

	g.createPlayer()
	g.actions = NewActions(PlayerActionsContext{
		Player:         g.state.GetPlayer(),
		ActionBroker:   g.actionBroker,
		State:          g.GetWorldState(),
		UI:             g.GetUserInterface,
		Msg:            g.PlayerMessage,
		Animator:       g.actionBroker,
		Speak:          g.ActorSpeak,
		Configuration:  g.configuration,
		QueueAfterTurn: g.queueAfterTurn,
	})
	g.state.SetOnMapChanged(g.onMapChanged)
	g.state.MapTransition(gridmap.Transition{
		TargetMap:      string(defs.LevelNameTown),
		TargetLocation: "spawn",
	})
	g.populateTown()

	townSpawn := g.state.GetMap().GetNamedLocation("spawn")
	g.state.SetTownSpawn(townSpawn)

	//g.spawnActorOnMap(g.state.GetPlayer(), townSpawn, playerStartDefinition)

	// load audio
	if g.configuration.AudioEnabled {
		sfxPath := path.Join("assets", "sfx")
		cues := []string{
			"heavy_door",
			"missile_hit_human",
			"heavy_hit_human",
			"levelup",
			"bat_attack",
			"skeleton_hit",
			"skeleton_death",
		}
		sfxFiles := make([]string, len(cues))
		for i, sfxFile := range cues {
			sfxFiles[i] = path.Join(sfxPath, sfxFile+".ogg")
		}
		g.audioPlayer.RegisterSoundCues(sfxFiles)
		g.audioPlayer.RegisterRandomizedSoundCues([]string{
			path.Join(sfxPath, "bow_fire"),
			//path.Join(sfxPath, "close_door"),
			path.Join(sfxPath, "open_door"),
			path.Join(sfxPath, "melee_attack"),
			path.Join(sfxPath, "human_hit"),
		})

		g.audioPlayer.PreLoadCuesIntoMemory(cues)
	}

	g.state.GetFlags().SetTrue("game_started")

	g.actionBroker.Execute(common.NewUpdateFoVAction(common.AbstractCause("Game Start"), g.state.GetPlayer()))

	g.uiController.UpdateStatsLabel()

	g.state.GetFlags().SetTrue(string(common.FlagStoryModeEnabled))
}

// SetAndInitUserInterface will be called after the game has finished initialization
// So we have to inform every component about the ui
func (g *GameInstance) SetAndInitUserInterface(ui common.UserInterface) {
	g.uiController = ui
	ui.SetPlayerMessage(g.PlayerMessage)
	ui.OnMapChanged()
	ui.SetDebugActions(g.getDebugActions)
}
func (g *GameInstance) queueAfterTurn(call func()) {
	g.executeAfterTurn = append(g.executeAfterTurn, call)
}
func (g *GameInstance) ActorSpeak(actor *common.Actor, trigger defs.ActorSpeechTrigger) {
	textLines := actor.GetSpeech(g.randomSource, trigger)
	if len(textLines) == 0 {
		return
	}

	if oldAnim, exists := g.actionBroker.GetAnimation(func(anim common.Animation) bool {
		if speechAnim, isSpeechAnim := anim.(*common.SpeechAnimation); isSpeechAnim {
			return speechAnim.GetActor() == actor
		}
		return false
	}); exists {
		oldAnim.Cancel()
	}

	speechAnim := common.NewSpeechAnimation(actor, textLines)
	g.actionBroker.AddAnimation(speechAnim)

	g.uiController.Say(actor, textLines)
	g.PlayerMessage(fmt.Sprintf("%s says: %s", actor.Name(), strings.Join(textLines, "\n")))
}

func (g *GameInstance) GetUserInterface() common.UserInterface {
	return g.uiController
}

func (g *GameInstance) handleLightSource() {
	player := g.state.GetPlayer()
	if player.GetEquipment().HasLightSourceEquipped() && !g.state.PlayerStandsInLitArea() {
		lightSource := player.GetEquipment().GetLightSource()
		asGameItem := lightSource
		asGameItem.UseCharge()
	}
}

// Advance moves the simulation forward by some number of in-game seconds. It is
// called from the fixed-timestep update rather than from a player action, which
// is the whole of the real-time conversion — see ADR 0001. One tick is one
// in-game second at the default Time Multiplier, making a Day twenty-four
// minutes of real time.
func (g *GameInstance) Advance(inGameSeconds float64) {
	if g.isPaused || g.ending != RunOngoing {
		return
	}
	scaled := inGameSeconds * g.state.GetTimeMultiplier()

	daysPassed := g.state.GetClock().Advance(scaled)
	g.state.GetTimeKeeper().AdvanceBySeconds(scaled)
	g.tickVampire(scaled)
	g.checkForTrespass()
	g.checkForBodyDiscovery()

	for i := 0; i < daysPassed; i++ {
		g.onDayBoundary()
	}

	for _, event := range g.executeAfterTurn {
		event()
	}
	g.executeAfterTurn = make([]func(), 0)

	player := g.state.GetPlayer()
	if player != nil && player.IsAlive() {
		g.handleLightSource()
		g.actions.walkPlayer(scaled)
		g.actions.nextStepOnPlayerPath()
	}

	g.uiController.OnPlayerTurnTaken()
}

// onDayBoundary is where anything reckoned in whole Days happens: Schedules are
// recomputed here rather than mid-street, and this is where Absence counting and
// Reassignment will hang.
func (g *GameInstance) onDayBoundary() {
	g.announcedToday = make(map[string]bool)
	g.privacyRefuge = make(map[string]geometry.Point)

	g.reassignVoidedObligations()

	for _, actor := range g.state.GetMap().Actors() {
		if person := actor.GetTownsperson(); person != nil {
			person.RecomputeSchedule()
		}
	}

	g.PlayerMessage(g.state.GetClock().String())
	g.showUnmetObligations()
}

func (g *GameInstance) SetPaused(paused bool) { g.isPaused = paused }
func (g *GameInstance) IsPaused() bool        { return g.isPaused }
func (g *GameInstance) TogglePause() bool {
	g.isPaused = !g.isPaused
	return g.isPaused
}

func (g *GameInstance) procFireAtLocation(location geometry.Point, startAnimating func() bool) {
	if g.state.GetMap().IsActorAt(location) {
		actor := g.state.GetMap().ActorAt(location)
		action := common.NewChangeResourceAction(common.AbstractCause("Fire"), actor, stats.Health, -3)
		action.SetAnimationStartPredicate(startAnimating)
		g.actionBroker.Execute(action)
	}
}
func (g *GameInstance) updateStatsDisplay() {
	g.uiController.UpdateStatsLabel()
	g.sendStatsToWindowClient()
}

// createPlayer takes nothing from the class loadout any more. The vampire has
// his own sprite and no stat block, so there is nothing left of the dungeon
// crawler's classes for him to be built out of.
func (g *GameInstance) createPlayer() *common.Actor {
	player := common.NewActor(defs.IconVampire)
	player.SetAsciiIcon('@')
	player.SetCP437Icon(int32(util.UnicodeToCP437Byte('☺')))
	player.SetColor(defs.White)
	player.SetName("Player")
	g.state.SetPlayer(player)

	// He is described entirely by his Blood, which is his health, his clock and
	// his motive at once. What is left in the attribute holder is engine
	// plumbing — how fast he walks, how far he sees.
	player.SetLife(g.blood)
	player.GetStats().SetOnChangeHandler(g.updateStatsDisplay)
	player.GetEquipment().SetOnChangeHandler(g.sendEquipmentToWindowClient)
	player.GetInventory().SetOnChangeHandler(g.sendInventoryToWindowClient)
	return player
}

func (g *GameInstance) spawnItem(mapPos geometry.Point, item *common.Item) {
	g.state.GetMap().AddItem(item, mapPos)
}

func (g *GameInstance) spawnItemInTown(namedLocation string, item *common.Item) {
	townMap := g.state.GetTownMap(string(defs.LevelNameTown))
	location := townMap.GetNamedLocation(namedLocation)
	townMap.AddItem(item, location)
}

func (g *GameInstance) GetWorldState() *common.State {
	return &g.state
}

func (g *GameInstance) killLevel() {
	actorsOnMap := g.state.GetMap().Actors()
	for i := len(actorsOnMap) - 1; i >= 0; i-- {
		actor := actorsOnMap[i]
		if actor == g.state.GetPlayer() || !actor.IsAlive() || actor.IsUnique() {
			continue
		}
		currentHP := actor.GetStats().GetResource(stats.Health).Value() * 2
		g.actionBroker.Execute(common.NewChangeResourceAction(common.NewActorCause("Kill All", g.state.GetPlayer()), actor, stats.Health, float64(-currentHP)))
	}
}

func (g *GameInstance) TransferRandomEnchantment(source *common.Item, destination *common.Item) {
	enchantments := source.GetAllEnchantments()

	if len(enchantments) == 0 {
		return
	}

	index := rand.Intn(len(enchantments))
	randomEnchantment := enchantments[index]

	source.RemoveEnchantment(randomEnchantment)
	destination.AddEnchantment(randomEnchantment)
}

func (g *GameInstance) onMapChanged(oldMapName string, newMapName string) {
	if defs.IsTownLevel(newMapName) {
		currentMap := g.state.GetMap()
		g.state.SetTownMap(currentMap)
		for _, actor := range currentMap.Actors() {
			if !actor.IsVendor() {
				continue
			}
			g.ensureVendorInventory(actor)
		}
	} else { // dungeon map
		g.state.SetDungeonMap(newMapName, g.state.GetMap())
	}

	if defs.IsTownLevel(oldMapName) {
		g.GetAudio().PlayCue("heavy_door")
	}

	g.actions.updateFoV()
	g.uiController.OnMapChanged()

	if g.onMapChange != nil {
		g.onMapChange()
	}
}

func (g *GameInstance) ensureVendorInventory(actor *common.Actor) {
	inventory := actor.GetInventory()
	if !inventory.IsEmpty() {
		return
	}
	var inventoryItems []*common.Item
	definition, exists := g.dataDefinitions.VendorDefinitions[string(actor.GetVendorType())]
	if !exists {
		g.ErrLog(fmt.Sprintf("No store definition for %s", actor.GetVendorType()))
		return
	}

	_ = definition

	for _, item := range inventoryItems {
		inventory.Add(item)
	}
}

func (g *GameInstance) getDisplacePositionForActor(victimToDisplace *common.Actor, direction geometry.Point) geometry.Point {
	currentMap := g.state.GetMap()
	victimPos := victimToDisplace.MapPosition()
	var displacePos geometry.Point
	stepsInDegrees := 45
	for turnBy := stepsInDegrees; turnBy < 180; turnBy += stepsInDegrees {
		curDir := geometry.RotateVector(direction, float64(turnBy))
		displacePos = victimPos.Add(curDir)
		if currentMap.IsCurrentlyPassable(displacePos) {
			return displacePos
		}

		curDir = geometry.RotateVector(direction, float64(-turnBy))
		displacePos = victimPos.Add(curDir)
		if currentMap.IsCurrentlyPassable(displacePos) {
			return displacePos
		}
	}
	displacePos = victimPos.Add(direction)
	if currentMap.IsCurrentlyPassable(displacePos) {
		return displacePos
	}
	// return source
	behindUs := victimPos.Add(direction.Mul(-1))
	return behindUs
}

func (g *GameInstance) sendStatsToWindowClient() {
	if !g.configuration.ServeStats {
		return
	}
	lines := g.GetWorldState().GetPlayer().GetFullStats()
	var asStrings []string
	for _, line := range lines {
		asStrings = append(asStrings, line.Text)
	}
	g.serverStats.Write(strings.Join(asStrings, "\n"))
}

func (g *GameInstance) sendInventoryToWindowClient() {
	if !g.configuration.ServeInventory {
		return
	}
	asStrings := g.state.GetPlayer().GetInventory().AsColorCodedStrings()
	g.serverInventory.Write(strings.Join(asStrings, "\n"))
}

func (g *GameInstance) sendEquipmentToWindowClient() {
	if !g.configuration.ServeEquipment {
		return
	}
	equipment := g.state.GetPlayer().GetEquipment()
	inventory := g.state.GetPlayer().GetInventory()
	items := sortEquipForUI(equipment)
	var asStrings []string

	maxItemLength := 0
	maxSlotLength := 0
	for _, item := range items {
		// index 0 => a, 1 => b, 2 => c, ...
		//itemColor := common.ColorCodeFromColor(item.TintColor())
		slotName := equipment.GetSlot(item)
		slotDesc := fmt.Sprintf("<-- %s", slotName.ToString())

		//itemLine := fmt.Sprintf("[:white]%c) %s%s", 'a'+index, itemColor, item.Name())
		shortcut := inventory.GetLetterForItem(item)
		itemLineWithoutColor := fmt.Sprintf("%c) %s", shortcut, item.BandStyleName())
		//minLineLength := len(itemLineWithoutColor) + len(slotDesc) + 1

		//asStrings = append(asStrings, itemLine)
		if len(itemLineWithoutColor) > maxItemLength {
			maxItemLength = len(itemLineWithoutColor)
		}
		if len(slotDesc) > maxSlotLength {
			maxSlotLength = len(slotDesc)
		}
	}

	for _, item := range items {
		// index 0 => a, 1 => b, 2 => c, ...
		itemColor := defs.ColorCodeFromColor(item.TintColor())

		slotName := equipment.GetSlot(item)
		slotDesc := fmt.Sprintf("[:white]<-- %s", slotName.ToString())
		shortcut := inventory.GetLetterForItem(item)
		itemLine := fmt.Sprintf("[:white]%c) %s%s", shortcut, itemColor, item.BandStyleName())

		withoutColor := common.RemoveColorCodes(slotDesc)
		padForSlot := (maxSlotLength + 1) - len(withoutColor)
		padForItem := maxItemLength - len(common.RemoveColorCodes(itemLine))
		slotDesc = util.LeftPadCount(slotDesc, padForSlot)
		itemLine = util.RightPadCount(itemLine, padForItem)

		asStrings = append(asStrings, itemLine+slotDesc)
	}
	g.serverEquipment.Write(strings.Join(asStrings, "\n"))
}
func sortEquipForUI(equip *common.Equipment) []*common.Item {
	itemList := equip.AllItems()

	slices.SortStableFunc(itemList, func(i, j *common.Item) int {
		slotNameI := equip.GetSlot(i)
		slotNameJ := equip.GetSlot(j)
		return cmp.Compare(slotNameI, slotNameJ)
	})
	return itemList
}
func (g *GameInstance) SetOnMapChange(onMapChange func()) {
	g.onMapChange = onMapChange
}

func (g *GameInstance) ErrLog(errorMessage string) {
	// write to stderr
	fmt.Fprintf(os.Stderr, errorMessage+"\n")
}

func (g *GameInstance) GetGameData() defs.DataContainer {
	return g.dataDefinitions
}

func (g *GameInstance) GetActions() common.PlayerActions {
	return g.actions
}

func (g *GameInstance) PlayerMessage(message string) {
	//println(text)
	rawText := common.RemoveColorCodes(message)
	if g.logHook != nil {
		g.logHook(rawText)
	}
	if len(g.logBuffer) > 200 {
		g.logBuffer = g.logBuffer[1:]
	}
	g.logBuffer = append(g.logBuffer, message)

	g.GetUserInterface().ShowLogMessage(message)
}

func (g *GameInstance) SetLogHook(hook func(string)) {
	g.logHook = hook
}

func (g *GameInstance) GetLog() []string {
	return g.logBuffer
}

func (g *GameInstance) getDebugActions() common.DebugActions {
	return common.DebugActions{
		KillLevel: g.killLevel,
		Menu:      g.debugMenu,
	}
}
