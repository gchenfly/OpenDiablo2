package d2gamescreen

import (
	"image/color"
	"math"
	"os"
	"strings"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"

	"github.com/OpenDiablo2/OpenDiablo2/d2game/d2player"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2clientconnectiontype"

	"github.com/hajimehoshi/ebiten"

	"github.com/OpenDiablo2/OpenDiablo2/d2common"
	dh "github.com/OpenDiablo2/OpenDiablo2/d2common"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2audio"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2render"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2screen"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

type CharacterSelect struct {
	background             *d2ui.Sprite
	newCharButton          d2ui.Button
	convertCharButton      d2ui.Button
	deleteCharButton       d2ui.Button
	exitButton             d2ui.Button
	okButton               d2ui.Button
	deleteCharCancelButton d2ui.Button
	deleteCharOkButton     d2ui.Button
	selectionBox           *d2ui.Sprite
	okCancelBox            *d2ui.Sprite
	d2HeroTitle            d2ui.Label
	deleteCharConfirmLabel d2ui.Label
	charScrollbar          d2ui.Scrollbar
	characterNameLabel     [8]d2ui.Label
	characterStatsLabel    [8]d2ui.Label
	characterExpLabel      [8]d2ui.Label
	characterImage         [8]*d2mapentity.Player
	gameStates             []*d2player.PlayerState
	selectedCharacter      int
	mouseButtonPressed     bool
	showDeleteConfirmation bool
	connectionType         d2clientconnectiontype.ClientConnectionType
	connectionHost         string
}

func CreateCharacterSelect(connectionType d2clientconnectiontype.ClientConnectionType, connectionHost string) *CharacterSelect {
	return &CharacterSelect{
		selectedCharacter: -1,
		connectionType:    connectionType,
		connectionHost:    connectionHost,
	}
}

func (v *CharacterSelect) OnLoad() error {
	d2audio.PlayBGM(d2resource.BGMTitle)

	animation, _ := d2asset.LoadAnimation(d2resource.CharacterSelectionBackground, d2resource.PaletteSky)
	v.background, _ = d2ui.LoadSprite(animation)
	v.background.SetPosition(0, 0)

	v.newCharButton = d2ui.CreateButton(d2ui.ButtonTypeTall, d2common.CombineStrings(dh.SplitIntoLinesWithMaxWidth("CREATE NEW CHARACTER", 15)))
	v.newCharButton.SetPosition(33, 468)
	v.newCharButton.OnActivated(func() { v.onNewCharButtonClicked() })
	d2ui.AddWidget(&v.newCharButton)

	v.convertCharButton = d2ui.CreateButton(d2ui.ButtonTypeTall, d2common.CombineStrings(dh.SplitIntoLinesWithMaxWidth("CONVERT TO EXPANSION", 15)))
	v.convertCharButton.SetPosition(233, 468)
	v.convertCharButton.SetEnabled(false)
	d2ui.AddWidget(&v.convertCharButton)

	v.deleteCharButton = d2ui.CreateButton(d2ui.ButtonTypeTall, d2common.CombineStrings(dh.SplitIntoLinesWithMaxWidth("DELETE CHARACTER", 15)))
	v.deleteCharButton.OnActivated(func() { v.onDeleteCharButtonClicked() })
	v.deleteCharButton.SetPosition(433, 468)
	d2ui.AddWidget(&v.deleteCharButton)

	v.exitButton = d2ui.CreateButton(d2ui.ButtonTypeMedium, "EXIT")
	v.exitButton.SetPosition(33, 537)
	v.exitButton.OnActivated(func() { v.onExitButtonClicked() })
	d2ui.AddWidget(&v.exitButton)

	v.deleteCharCancelButton = d2ui.CreateButton(d2ui.ButtonTypeOkCancel, "NO")
	v.deleteCharCancelButton.SetPosition(282, 308)
	v.deleteCharCancelButton.SetVisible(false)
	v.deleteCharCancelButton.OnActivated(func() { v.onDeleteCharacterCancelClicked() })
	d2ui.AddWidget(&v.deleteCharCancelButton)

	v.deleteCharOkButton = d2ui.CreateButton(d2ui.ButtonTypeOkCancel, "YES")
	v.deleteCharOkButton.SetPosition(422, 308)
	v.deleteCharOkButton.SetVisible(false)
	v.deleteCharOkButton.OnActivated(func() { v.onDeleteCharacterConfirmClicked() })
	d2ui.AddWidget(&v.deleteCharOkButton)

	v.okButton = d2ui.CreateButton(d2ui.ButtonTypeMedium, "OK")
	v.okButton.SetPosition(625, 537)
	v.okButton.OnActivated(func() { v.onOkButtonClicked() })
	d2ui.AddWidget(&v.okButton)

	v.d2HeroTitle = d2ui.CreateLabel(d2resource.Font42, d2resource.PaletteUnits)
	v.d2HeroTitle.SetPosition(320, 23)
	v.d2HeroTitle.Alignment = d2ui.LabelAlignCenter

	v.deleteCharConfirmLabel = d2ui.CreateLabel(d2resource.Font16, d2resource.PaletteUnits)
	lines := dh.SplitIntoLinesWithMaxWidth("Are you sure that you want to delete this character? Take note: this will delete all versions of this Character.", 29)
	v.deleteCharConfirmLabel.SetText(strings.Join(lines, "\n"))
	v.deleteCharConfirmLabel.Alignment = d2ui.LabelAlignCenter
	v.deleteCharConfirmLabel.SetPosition(400, 185)

	animation, _ = d2asset.LoadAnimation(d2resource.CharacterSelectionSelectBox, d2resource.PaletteSky)
	v.selectionBox, _ = d2ui.LoadSprite(animation)
	v.selectionBox.SetPosition(37, 86)

	animation, _ = d2asset.LoadAnimation(d2resource.PopUpOkCancel, d2resource.PaletteFechar)
	v.okCancelBox, _ = d2ui.LoadSprite(animation)
	v.okCancelBox.SetPosition(270, 175)

	v.charScrollbar = d2ui.CreateScrollbar(586, 87, 369)
	v.charScrollbar.OnActivated(func() { v.onScrollUpdate() })
	d2ui.AddWidget(&v.charScrollbar)

	for i := 0; i < 8; i++ {
		xOffset := 115
		if i&1 > 0 {
			xOffset = 385
		}
		v.characterNameLabel[i] = d2ui.CreateLabel(d2resource.Font16, d2resource.PaletteUnits)
		v.characterNameLabel[i].Color = color.RGBA{R: 188, G: 168, B: 140, A: 255}
		v.characterNameLabel[i].SetPosition(xOffset, 100+((i/2)*95))
		v.characterStatsLabel[i] = d2ui.CreateLabel(d2resource.Font16, d2resource.PaletteUnits)
		v.characterStatsLabel[i].SetPosition(xOffset, 115+((i/2)*95))
		v.characterExpLabel[i] = d2ui.CreateLabel(d2resource.Font16, d2resource.PaletteStatic)
		v.characterExpLabel[i].Color = color.RGBA{R: 24, G: 255, A: 255}
		v.characterExpLabel[i].SetPosition(xOffset, 130+((i/2)*95))
	}
	v.refreshGameStates()

	return nil
}

func (v *CharacterSelect) onScrollUpdate() {
	v.moveSelectionBox()
	v.updateCharacterBoxes()
}

func (v *CharacterSelect) updateCharacterBoxes() {
	expText := "EXPANSION CHARACTER"
	for i := 0; i < 8; i++ {
		idx := i + (v.charScrollbar.GetCurrentOffset() * 2)
		if idx >= len(v.gameStates) {
			v.characterNameLabel[i].SetText("")
			v.characterStatsLabel[i].SetText("")
			v.characterExpLabel[i].SetText("")
			v.characterImage[i] = nil
			continue
		}
		v.characterNameLabel[i].SetText(v.gameStates[idx].HeroName)
		v.characterStatsLabel[i].SetText("Level 1 " + v.gameStates[idx].HeroType.String())
		v.characterExpLabel[i].SetText(expText)
		// TODO: Generate or load the object from the actual player data...
		v.characterImage[i] = d2mapentity.CreatePlayer("", "", 0, 0, 0,
			v.gameStates[idx].HeroType,
			d2inventory.HeroObjects[v.gameStates[idx].HeroType],
		)
	}
}

func (v *CharacterSelect) onNewCharButtonClicked() {
	d2screen.SetNextScreen(CreateSelectHeroClass(v.connectionType, v.connectionHost))
}

func (v *CharacterSelect) onExitButtonClicked() {
	mainMenu := CreateMainMenu()
	mainMenu.SetScreenMode(ScreenModeMainMenu)
	d2screen.SetNextScreen(mainMenu)
}

func (v *CharacterSelect) Render(screen d2render.Surface) error {
	v.background.RenderSegmented(screen, 4, 3, 0)
	v.d2HeroTitle.Render(screen)
	actualSelectionIndex := v.selectedCharacter - (v.charScrollbar.GetCurrentOffset() * 2)
	if v.selectedCharacter > -1 && actualSelectionIndex >= 0 && actualSelectionIndex < 8 {
		v.selectionBox.RenderSegmented(screen, 2, 1, 0)
	}
	for i := 0; i < 8; i++ {
		idx := i + (v.charScrollbar.GetCurrentOffset() * 2)
		if idx >= len(v.gameStates) {
			continue
		}
		v.characterNameLabel[i].Render(screen)
		v.characterStatsLabel[i].Render(screen)
		v.characterExpLabel[i].Render(screen)
		screen.PushTranslation(v.characterNameLabel[i].X-40, v.characterNameLabel[i].Y+50)
		v.characterImage[i].Render(screen)
		screen.Pop()
	}
	if v.showDeleteConfirmation {
		screen.DrawRect(800, 600, color.RGBA{A: 128})
		v.okCancelBox.RenderSegmented(screen, 2, 1, 0)
		v.deleteCharConfirmLabel.Render(screen)
	}

	return nil
}

func (v *CharacterSelect) moveSelectionBox() {
	if v.selectedCharacter == -1 {
		v.d2HeroTitle.SetText("")
		return
	}
	bw := 272
	bh := 92
	selectedIndex := v.selectedCharacter - (v.charScrollbar.GetCurrentOffset() * 2)
	v.selectionBox.SetPosition(37+((selectedIndex&1)*bw), 86+(bh*(selectedIndex/2)))
	v.d2HeroTitle.SetText(v.gameStates[v.selectedCharacter].HeroName)
}

func (v *CharacterSelect) Advance(tickTime float64) error {
	if !v.showDeleteConfirmation {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if !v.mouseButtonPressed {
				v.mouseButtonPressed = true
				mx, my := ebiten.CursorPosition()
				bw := 272
				bh := 92
				localMouseX := mx - 37
				localMouseY := my - 86
				if localMouseX > 0 && localMouseX < bw*2 && localMouseY >= 0 && localMouseY < bh*4 {
					adjustY := localMouseY / bh
					selectedIndex := adjustY * 2
					if localMouseX > bw {
						selectedIndex += 1
					}
					if (v.charScrollbar.GetCurrentOffset()*2)+selectedIndex < len(v.gameStates) {
						v.selectedCharacter = (v.charScrollbar.GetCurrentOffset() * 2) + selectedIndex
						v.moveSelectionBox()
					}
				}
			}
		} else {
			v.mouseButtonPressed = false
		}
	}
	for _, hero := range v.characterImage {
		if hero != nil {
			hero.AnimatedComposite.Advance(tickTime)
		}
	}

	return nil
}

func (v *CharacterSelect) onDeleteCharButtonClicked() {
	v.toggleDeleteCharacterDialog(true)
}

func (v *CharacterSelect) onDeleteCharacterConfirmClicked() {
	_ = os.Remove(v.gameStates[v.selectedCharacter].FilePath)
	v.charScrollbar.SetCurrentOffset(0)
	v.refreshGameStates()
	v.toggleDeleteCharacterDialog(false)
	v.deleteCharButton.SetEnabled(len(v.gameStates) > 0)
	v.okButton.SetEnabled(len(v.gameStates) > 0)
}

func (v *CharacterSelect) onDeleteCharacterCancelClicked() {
	v.toggleDeleteCharacterDialog(false)
}

func (v *CharacterSelect) toggleDeleteCharacterDialog(showDialog bool) {
	v.showDeleteConfirmation = showDialog
	v.okButton.SetEnabled(!showDialog)
	v.deleteCharButton.SetEnabled(!showDialog)
	v.exitButton.SetEnabled(!showDialog)
	v.newCharButton.SetEnabled(!showDialog)
	v.deleteCharOkButton.SetVisible(showDialog)
	v.deleteCharCancelButton.SetVisible(showDialog)
}

func (v *CharacterSelect) refreshGameStates() {
	v.gameStates = d2player.GetAllPlayerStates()
	v.updateCharacterBoxes()
	if len(v.gameStates) > 0 {
		v.selectedCharacter = 0
		v.d2HeroTitle.SetText(v.gameStates[0].HeroName)
		v.charScrollbar.SetMaxOffset(int(math.Ceil(float64(len(v.gameStates)-8) / float64(2))))
	} else {
		v.selectedCharacter = -1
		v.charScrollbar.SetMaxOffset(0)
	}
	v.moveSelectionBox()

}

func (v *CharacterSelect) onOkButtonClicked() {
	gameClient, _ := d2client.Create(v.connectionType)
	switch v.connectionType {
	case d2clientconnectiontype.LANClient:
		gameClient.Open(v.connectionHost, v.gameStates[v.selectedCharacter].FilePath)
	default:
		gameClient.Open("", v.gameStates[v.selectedCharacter].FilePath)
	}

	d2screen.SetNextScreen(CreateGame(gameClient))
}
