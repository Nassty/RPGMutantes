package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"

	irc "github.com/thoj/go-ircevent"
)

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x

}
func Pow(a, b int) int {
	p := 1
	for b > 0 {
		if b&1 != 0 {
			p *= a
		}
		b >>= 1
		a *= a
	}
	return p
}

type Property struct {
	Name    string
	Current int
	Max     int
}

func (p Property) String() string {
	return fmt.Sprintf("%s: %d/%d", p.Name,
		p.Current, p.Max)
}

type Item struct {
	Name   string
	Prefix string
	AT     int
}

func NewItem(name, prefix string, at int) *Item {
	return &Item{
		name, prefix, at,
	}
}

func (i *Item) String() string {
	return fmt.Sprintf("%s +%d", i.Name, i.AT)
}

type Player struct {
	Name  string
	AP    *Property
	XP    float64
	HP    *Property
	AT    int
	Level int
	Items []*Item
}

func NewPlayer(name string, ap, hp, at int) *Player {
	return &Player{
		Name:  name,
		AP:    &Property{Name: "AP", Current: ap, Max: ap},
		XP:    0,
		HP:    &Property{Name: "HP", Current: hp, Max: hp},
		AT:    at,
		Level: 1,
	}
}

func (current *Player) Hit(hit int) int {
	left := Max(0, hit-current.AP.Current)
	current.AP.Current = Max(0, current.AP.Current-hit)
	current.HP.Current = Max(0, current.HP.Current-left)
	return hit
}

func (current *Player) AddItem(item *Item) {
	current.Items = append(current.Items, item)
}
func (current *Player) IncreaseXP(by int) (float64, bool) {
	var next_level int
	level_up := false
	increased := float64(by) * math.Pow(10.0, float64(-1*current.Level))
	current.XP = current.XP + increased
	next_level = Max(1, int(current.XP)/10+1)
	if next_level > current.Level {
		level_up = true
		current.HP.Max += rand.Intn(5 * next_level)
		current.HP.Current = current.HP.Max
		current.AP.Max += rand.Intn(5 * next_level)
		current.AP.Current = current.AP.Max
	}
	current.Level = next_level
	return increased, level_up
}

func (current *Player) Attack(target *Player) string {

	msg := "[Attack] %s hits %s for %d points "

	if current.HP.Current <= 0 {
		return fmt.Sprintf("%s has no HP, cannot attack", current.Name)
	}
	if target.HP.Current <= 0 {
		return fmt.Sprintf("%s has no HP, cannot be attacked", target.Name)
	}

	modifiers := 0
	if current.Items != nil {
		item := current.Items[rand.Intn(len(current.Items))]
		modifiers += item.AT
		msg = msg + fmt.Sprintf("with %s %s +%d ", item.Prefix, item.Name, item.AT)
	}
	msg += "earns  %.3fXP"

	if current.Level-target.Level < -1*current.Level {
		return fmt.Sprintf("%s cannot attack %s: big experience gap",
			current.Name, target.Name)
	}
	hit := Max(0, rand.Intn(current.AT+modifiers))
	att := target.Hit(hit)
	xp, level_up := current.IncreaseXP(att)
	if level_up {
		msg += " Level Up!!!"
	}
	return fmt.Sprintf(msg, current.Name, target.Name, att, xp)
}

func (current Player) String() string {
	items := ""
	for _, i := range current.Items {
		items += i.String() + " "
	}
	return fmt.Sprintf("%s (%d) [%s] [%s], [XP:%.1f] %s",
		current.Name, current.Level,
		current.AP, current.HP, current.XP, items)
}

type Arena struct {
	Players map[string]*Player
	Items   []*Item
}

func NewArena() *Arena {
	var items []*Item
	return &Arena{
		make(map[string]*Player),
		items,
	}
}

func (a *Arena) AddPlayer(player *Player) string {
	_, ok := a.Players[player.Name]
	if ok == true {
		return fmt.Sprintf("player %s already on the arena", player.Name)
	}
	a.Players[player.Name] = player
	return fmt.Sprintf("player %s joins the arena", player.Name)
}
func (a *Arena) AddItem(item *Item) string {
	a.Items = append(a.Items, item)
	return fmt.Sprintf("item %s added to the arena", item.Name)
}

func (a *Arena) Parse(author, input string) string {
	args := strings.Split(input, " ")
	action := args[0]
	targets := args[1:]
	switch action {
	case "join", "joins", "enter":
		new_player := NewPlayer(author, rand.Intn(5)+5,
			rand.Intn(5)+5,
			rand.Intn(3)+2)
		new_player.Items = a.Items
		return a.AddPlayer(new_player)
	case "attack", "attacks":
		player, ok := a.Players[author]
		if ok == false {
			return fmt.Sprintf("%s not in arena, please use the JOIN command first", author)
		}
		target, ok := a.Players[targets[0]]
		if ok == false {
			return fmt.Sprintf("%s not in arena", author)
		}
		return player.Attack(target)
	case "status":
		player, ok := a.Players[author]
		if ok == false {
			return fmt.Sprintf("%s not in arena, please use the JOIN command first", author)
		}
		return player.String()

	}
	return action
}

func main() {
	arena := NewArena()
	items := [...]string{"banana", "sword", "pineaple", "katana"}
	for _, item := range items {
		arena.AddItem(NewItem(item, "a", rand.Intn(3)+1))
	}
	roomName := "#mutant3s"

	con := irc.IRC("RPGMutantes", "RPGMutantes")
	err := con.Connect("irc.freenode.net:6667")
	if err != nil {
		log.Println("Failed connecting")
		return
	}
	con.AddCallback("001", func(e *irc.Event) {
		con.Join(roomName)
	})
	con.AddCallback("CTCP_ACTION", func(e *irc.Event) {
		con.Action(roomName, arena.Parse(e.Nick, e.Message()))
	})
	con.Loop()
}
